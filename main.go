package main

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/gorilla/sessions"
)

type ConfigFile struct {
	HttpBindAddr   string `json:"http_bind_addr"`
	SessionKey     string `json:"session_key"`
	LdapServerAddr string `json:"ldap_server_addr"`
	LdapTLS        bool   `json:"ldap_tls"`
	UserFormat     string `json:"user_format"`
	GroupCanInvite string `json:"group_can_invite"`
	GroupCanAdmin  string `json:"group_can_admin"`
}

var configFlag = flag.String("config", "./config.json", "Configuration file path")

var config *ConfigFile

const SESSION_NAME = "guichet_session"

var store sessions.Store = nil

func readConfig() ConfigFile {
	key_bytes := make([]byte, 32)
	n, err := rand.Read(key_bytes)
	if err != nil || n != 32 {
		log.Fatal(err)
	}

	config_file := ConfigFile{
		HttpBindAddr:   ":9991",
		SessionKey:     base64.StdEncoding.EncodeToString(key_bytes),
		LdapServerAddr: "ldap://127.0.0.1:389",
		LdapTLS:        false,
		UserFormat:     "cn=%s,ou=users,dc=example,dc=com",
	}

	_, err = os.Stat(*configFlag)
	if os.IsNotExist(err) {
		// Generate default config file
		log.Printf("Generating default config file as %s", *configFlag)

		bytes, err := json.MarshalIndent(&config_file, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		err = ioutil.WriteFile(*configFlag, bytes, 0644)
		if err != nil {
			log.Fatal(err)
		}

		return config_file
	}

	if err != nil {
		log.Fatal(err)
	}

	bytes, err := ioutil.ReadFile(*configFlag)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(bytes, &config_file)
	if err != nil {
		log.Fatal(err)
	}

	return config_file
}

func main() {
	flag.Parse()

	config_file := readConfig()
	config = &config_file
	store = sessions.NewFilesystemStore("", []byte(config.SessionKey))

	http.HandleFunc("/", handleHome)
	http.HandleFunc("/logout", handleLogout)
	http.HandleFunc("/profile", handleProfile)
	http.HandleFunc("/passwd", handlePasswd)

	staticfiles := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", staticfiles))

	err := http.ListenAndServe(config.HttpBindAddr, logRequest(http.DefaultServeMux))
	if err != nil {
		log.Fatal("Cannot start http server: ", err)
	}
}

type LoginInfo struct {
	Username string
	DN       string
	Password string
}

type LoginStatus struct {
	Info      *LoginInfo
	conn      *ldap.Conn
	UserEntry *ldap.Entry
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func checkLogin(w http.ResponseWriter, r *http.Request) *LoginStatus {
	session, err := store.Get(r, SESSION_NAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	username, ok := session.Values["login_username"]
	password, ok2 := session.Values["login_password"]
	user_dn, ok3 := session.Values["login_dn"]

	var login_info *LoginInfo
	if !(ok && ok2 && ok3) {
		login_info = handleLogin(w, r)
		if login_info == nil {
			return nil
		}
	} else {
		login_info = &LoginInfo{
			DN:       user_dn.(string),
			Username: username.(string),
			Password: password.(string),
		}
	}

	l := ldapOpen(w)
	if l == nil {
		return nil
	}

	err = l.Bind(login_info.DN, login_info.Password)
	if err != nil {
		delete(session.Values, "login_username")
		delete(session.Values, "login_password")
		delete(session.Values, "login_dn")

		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil
		}
		return checkLogin(w, r)
	}

	searchRequest := ldap.NewSearchRequest(
		login_info.DN,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=organizationalPerson))"),
		[]string{"dn", "displayname", "givenname", "sn", "mail", "memberof"},
		nil)

	sr, err := l.Search(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	if len(sr.Entries) != 1 {
		http.Error(w, fmt.Sprintf("Multiple entries: %#v", sr.Entries), http.StatusInternalServerError)
		return nil
	}

	return &LoginStatus{
		Info:      login_info,
		conn:      l,
		UserEntry: sr.Entries[0],
	}
}

func ldapOpen(w http.ResponseWriter) *ldap.Conn {
	l, err := ldap.DialURL(config.LdapServerAddr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	if config.LdapTLS {
		err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil
		}
	}

	return l
}

// Templates ----

type LoginFormData struct {
	Username     string
	ErrorMessage string
}

// Page handlers ----

type HomePageData struct {
	Login     *LoginStatus
	CanAdmin  bool
	CanInvite bool
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	templateHome := template.Must(template.ParseFiles("templates/layout.html", "templates/home.html"))

	login := checkLogin(w, r)
	if login == nil {
		return
	}

	can_admin := false
	can_invite := false
	for _, group := range login.UserEntry.GetAttributeValues("memberof") {
		if config.GroupCanInvite != "" && group == config.GroupCanInvite {
			can_invite = true
		}
		if config.GroupCanAdmin != "" && group == config.GroupCanAdmin {
			can_admin = true
		}
	}

	templateHome.Execute(w, &HomePageData{
		Login:     login,
		CanAdmin:  can_admin,
		CanInvite: can_invite,
	})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, SESSION_NAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	delete(session.Values, "login_username")
	delete(session.Values, "login_password")
	delete(session.Values, "login_dn")

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func handleLogin(w http.ResponseWriter, r *http.Request) *LoginInfo {
	templateLogin := template.Must(template.ParseFiles("templates/layout.html", "templates/login.html"))

	if r.Method == "GET" {
		templateLogin.Execute(w, LoginFormData{})
		return nil
	} else if r.Method == "POST" {
		r.ParseForm()

		username := strings.Join(r.Form["username"], "")
		password := strings.Join(r.Form["password"], "")
		user_dn := strings.ReplaceAll(config.UserFormat, "%s", username)

		l := ldapOpen(w)
		if l == nil {
			return nil
		}

		err := l.Bind(user_dn, password)
		if err != nil {
			templateLogin.Execute(w, LoginFormData{
				Username:     username,
				ErrorMessage: err.Error(),
			})
			return nil
		}

		// Successfully logged in, save it to session
		session, err := store.Get(r, SESSION_NAME)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil
		}

		session.Values["login_username"] = username
		session.Values["login_password"] = password
		session.Values["login_dn"] = user_dn

		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil
		}

		return &LoginInfo{
			DN:       user_dn,
			Username: username,
			Password: password,
		}
	} else {
		http.Error(w, "Unsupported method", http.StatusBadRequest)
		return nil
	}
}

type ProfileTplData struct {
	Status       *LoginStatus
	ErrorMessage string
	Success      bool
	Mail         string
	DisplayName  string
	GivenName    string
	Surname      string
}

func handleProfile(w http.ResponseWriter, r *http.Request) {
	templateProfile := template.Must(template.ParseFiles("templates/layout.html", "templates/profile.html"))

	login := checkLogin(w, r)
	if login == nil {
		return
	}

	data := &ProfileTplData{
		Status:       login,
		ErrorMessage: "",
		Success:      false,
	}

	if r.Method == "POST" {
		r.ParseForm()

		data.Mail = strings.Join(r.Form["mail"], "")
		data.DisplayName = strings.Join(r.Form["display_name"], "")
		data.GivenName = strings.Join(r.Form["given_name"], "")
		data.Surname = strings.Join(r.Form["surname"], "")

		modify_request := ldap.NewModifyRequest(login.Info.DN, nil)
		modify_request.Replace("mail", []string{data.Mail})
		modify_request.Replace("displayname", []string{data.DisplayName})
		modify_request.Replace("givenname", []string{data.GivenName})
		modify_request.Replace("sn", []string{data.Surname})

		err := login.conn.Modify(modify_request)
		if err != nil {
			data.ErrorMessage = err.Error()
		} else {
			data.Success = true
		}
	} else {
		data.Mail = login.UserEntry.GetAttributeValue("mail")
		data.DisplayName = login.UserEntry.GetAttributeValue("displayname")
		data.GivenName = login.UserEntry.GetAttributeValue("givenname")
		data.Surname = login.UserEntry.GetAttributeValue("sn")
	}

	templateProfile.Execute(w, data)
}

type PasswdTplData struct {
	Status       *LoginStatus
	ErrorMessage string
	NoMatchError bool
	Success      bool
}

func handlePasswd(w http.ResponseWriter, r *http.Request) {
	templatePasswd := template.Must(template.ParseFiles("templates/layout.html", "templates/passwd.html"))

	login := checkLogin(w, r)
	if login == nil {
		return
	}

	data := &PasswdTplData{
		Status:       login,
		ErrorMessage: "",
		Success:      false,
	}

	if r.Method == "POST" {
		r.ParseForm()

		password := strings.Join(r.Form["password"], "")
		password2 := strings.Join(r.Form["password2"], "")

		if password2 != password {
			data.NoMatchError = true
		} else {
			modify_request := ldap.NewModifyRequest(login.Info.DN, nil)
			modify_request.Replace("userpassword", []string{SSHAEncode([]byte(password))})
			err := login.conn.Modify(modify_request)
			if err != nil {
				data.ErrorMessage = err.Error()
			} else {
				data.Success = true
			}
		}
	}

	templatePasswd.Execute(w, data)
}
