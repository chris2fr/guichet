package main

import (
	"crypto/rand"
	"crypto/tls"
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
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type ConfigFile struct {
	HttpBindAddr   string `json:"http_bind_addr"`
	LdapServerAddr string `json:"ldap_server_addr"`
	LdapTLS        bool   `json:"ldap_tls"`

	BaseDN        string `json:"base_dn"`
	UserBaseDN    string `json:"user_base_dn"`
	UserNameAttr  string `json:"user_name_attr"`
	GroupBaseDN   string `json:"group_base_dn"`
	GroupNameAttr string `json:"group_name_attr"`

	AdminAccount   string `json:"admin_account"`
	GroupCanInvite string `json:"group_can_invite"`
	GroupCanAdmin  string `json:"group_can_admin"`
}

var configFlag = flag.String("config", "./config.json", "Configuration file path")

var config *ConfigFile

const SESSION_NAME = "guichet_session"

var store sessions.Store = nil

func readConfig() ConfigFile {
	config_file := ConfigFile{
		HttpBindAddr:   ":9991",
		LdapServerAddr: "ldap://127.0.0.1:389",
		LdapTLS:        false,
		BaseDN:         "dc=example,dc=com",
		UserBaseDN:     "ou=users,dc=example,dc=com",
		UserNameAttr:   "uid",
		GroupBaseDN:    "ou=groups,dc=example,dc=com",
		GroupNameAttr:  "gid",
		AdminAccount:   "uid=admin,dc=example,dc=com",
		GroupCanInvite: "",
		GroupCanAdmin:  "gid=admin,ou=groups,dc=example,dc=com",
	}

	_, err := os.Stat(*configFlag)
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

	session_key := make([]byte, 32)
	n, err := rand.Read(session_key)
	if err != nil || n != 32 {
		log.Fatal(err)
	}
	store = sessions.NewCookieStore(session_key)

	r := mux.NewRouter()
	r.HandleFunc("/", handleHome)
	r.HandleFunc("/logout", handleLogout)
	r.HandleFunc("/profile", handleProfile)
	r.HandleFunc("/passwd", handlePasswd)

	r.HandleFunc("/admin/users", handleAdminUsers)
	r.HandleFunc("/admin/groups", handleAdminGroups)
	r.HandleFunc("/admin/ldap/{dn}", handleAdminLDAP)
	r.HandleFunc("/admin/create/{template}/{super_dn}", handleAdminCreate)

	staticfiles := http.FileServer(http.Dir("static"))
	r.Handle("/static/{file:.*}", http.StripPrefix("/static/", staticfiles))

	log.Printf("Starting HTTP server on %s", config.HttpBindAddr)
	err = http.ListenAndServe(config.HttpBindAddr, logRequest(r))
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
	CanAdmin  bool
	CanInvite bool
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func checkLogin(w http.ResponseWriter, r *http.Request) *LoginStatus {
	var login_info *LoginInfo

	session, err := store.Get(r, SESSION_NAME)
	if err == nil {
		username, ok := session.Values["login_username"]
		password, ok2 := session.Values["login_password"]
		user_dn, ok3 := session.Values["login_dn"]

		if ok && ok2 && ok3 {
			login_info = &LoginInfo{
				DN:       user_dn.(string),
				Username: username.(string),
				Password: password.(string),
			}
		}
	}

	if login_info == nil {
		login_info = handleLogin(w, r)
		if login_info == nil {
			return nil
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

	loginStatus := &LoginStatus{
		Info: login_info,
		conn: l,
	}

	requestKind := "(objectClass=organizationalPerson)"
	if strings.EqualFold(login_info.DN, config.AdminAccount) {
		requestKind = "(objectclass=*)"
	}
	searchRequest := ldap.NewSearchRequest(
		login_info.DN,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		requestKind,
		[]string{"dn", "displayname", "givenname", "sn", "mail", "memberof"},
		nil)

	sr, err := l.Search(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	if len(sr.Entries) != 1 {
		http.Error(w, fmt.Sprintf("Unable to find entry for %s", login_info.DN), http.StatusInternalServerError)
		return nil
	}

	loginStatus.UserEntry = sr.Entries[0]

	loginStatus.CanAdmin = strings.EqualFold(loginStatus.Info.DN, config.AdminAccount)
	loginStatus.CanInvite = false
	for _, attr := range loginStatus.UserEntry.Attributes {
		if strings.EqualFold(attr.Name, "memberof") {
			for _, group := range attr.Values {
				if config.GroupCanInvite != "" && strings.EqualFold(group, config.GroupCanInvite) {
					loginStatus.CanInvite = true
				}
				if config.GroupCanAdmin != "" && strings.EqualFold(group, config.GroupCanAdmin) {
					loginStatus.CanAdmin = true
				}
			}
		}
	}

	return loginStatus
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

// Page handlers ----

type HomePageData struct {
	Login       *LoginStatus
	WelcomeName string
	BaseDN      string
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	templateHome := template.Must(template.ParseFiles("templates/layout.html", "templates/home.html"))

	login := checkLogin(w, r)
	if login == nil {
		return
	}

	data := &HomePageData{
		Login:       login,
		BaseDN:      config.BaseDN,
		WelcomeName: login.UserEntry.GetAttributeValue("givenname"),
	}
	if data.WelcomeName == "" {
		data.WelcomeName = login.UserEntry.GetAttributeValue("displayname")
	}
	if data.WelcomeName == "" {
		data.WelcomeName = login.Info.Username
	}

	templateHome.Execute(w, data)
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

type LoginFormData struct {
	Username     string
	ErrorMessage string
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
		user_dn := fmt.Sprintf("%s=%s,%s", config.UserNameAttr, username, config.UserBaseDN)
		if strings.EqualFold(username, config.AdminAccount) {
			user_dn = username
		}

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
			session, _ = store.New(r, SESSION_NAME)
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
