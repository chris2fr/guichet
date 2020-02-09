package main

import (
	"os"
	"strings"
	"flag"
	"log"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"encoding/base64"
	"crypto/rand"
	"crypto/tls"
	"html/template"

	"github.com/gorilla/sessions"
	"github.com/go-ldap/ldap/v3"
)

type ConfigFile struct {
	HttpBindAddr   string `json:"http_bind_addr"`
	SessionKey     string `json:"session_key"`
	LdapServerAddr string `json:"ldap_server_addr"`
	LdapTLS bool `json:"ldap_tls"`
	UserFormat string `json:"user_format"`
}

var configFlag = flag.String("config", "./config.json", "Configuration file path")

var config *ConfigFile

const SESSION_NAME = "guichet_session"
var store *sessions.CookieStore = nil

func readConfig() ConfigFile{
	key_bytes := make([]byte, 32)
	n, err := rand.Read(key_bytes)
	if err!= nil || n != 32 {
		log.Fatal(err)
	}

	config_file := ConfigFile{
		HttpBindAddr:   ":9991",
		SessionKey:     base64.StdEncoding.EncodeToString(key_bytes),
		LdapServerAddr: "ldap://127.0.0.1:389",
		LdapTLS: false,
		UserFormat: "cn=%s,ou=users,dc=example,dc=com",
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
	store = sessions.NewCookieStore([]byte(config.SessionKey))

	http.HandleFunc("/", handleHome)

	staticfiles := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", staticfiles))

	err := http.ListenAndServe(config.HttpBindAddr, logRequest(http.DefaultServeMux))
	if err != nil {
		log.Fatal("Cannot start http server: ", err)
	}
}

type LoginInfo struct {
	Username string
	DN string
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

	login_info, has_login_info := session.Values["login_info"]
	if !has_login_info {
		return handleLogin(w, r)
	}

	return login_info.(*LoginInfo)
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
	Username string
	ErrorMessage string
}

var (
	templateLogin = template.Must(template.ParseFiles("templates/layout.html", "templates/login.html"))
	templateHome = template.Must(template.ParseFiles("templates/layout.html", "templates/home.html"))
)

// Page handlers ----

func handleHome(w http.ResponseWriter, r *http.Request) {
	login := checkLogin(w, r)
	if login == nil {
		return
	}

	templateHome.Execute(w, login)
}

func handleLogin(w http.ResponseWriter, r *http.Request) *LoginInfo {
	if r.Method == "GET" {
		templateLogin.Execute(w, LoginFormData{})
		return nil
	} else if r.Method == "POST" {
		r.ParseForm()

		username := strings.Join(r.Form["username"], "")
		user_dn := strings.ReplaceAll(config.UserFormat, "%s", username)

		login_info := &LoginInfo{
			DN: user_dn,
			Username: username,
			Password: strings.Join(r.Form["password"], ""),
		}

		l := ldapOpen(w)
		if l == nil {
			return nil
		}

		err := l.Bind(user_dn, login_info.Password)
		if err != nil {
			templateLogin.Execute(w, LoginFormData{
				Username: username,
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

		session.Values["login_info"] = login_info
		return login_info
	} else {
		http.Error(w, "Unsupported method", http.StatusBadRequest)
		return nil
	}
}
