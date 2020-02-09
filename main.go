package main

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
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

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func checkLogin(w http.ResponseWriter, r *http.Request) *LoginInfo {
	session, err := store.Get(r, SESSION_NAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	username, ok := session.Values["login_username"]
	password, ok2 := session.Values["login_password"]
	user_dn, ok3 := session.Values["login_dn"]
	if !(ok && ok2 && ok3) {
		return handleLogin(w, r)
	}

	return &LoginInfo{
		DN:       user_dn.(string),
		Username: username.(string),
		Password: password.(string),
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

func handleHome(w http.ResponseWriter, r *http.Request) {
	templateHome := template.Must(template.ParseFiles("templates/layout.html", "templates/home.html"))

	login := checkLogin(w, r)
	if login == nil {
		return
	}

	templateHome.Execute(w, login)
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
