/*
Guichet provides a user-management system around an LDAP Directory

Oriniated with deuxfleurs.fr and advanced by resdigita.com
*/
package main

import (
	"crypto/rand"
	"crypto/tls"

	// "encoding/json"
	"flag"
	// "fmt"
	"html/template"
	// "io/ioutil"
	"log"
	"net/http"

	// "os"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

const SESSION_NAME = "guichet_session"

var staticPath = "./static"
var templatePath = "./templates"

var store sessions.Store = nil

func getTemplate(name string) *template.Template {
	return template.Must(template.New("layout.html").Funcs(template.FuncMap{
		"contains": strings.Contains,
	}).ParseFiles(
		templatePath+"/layout.html",
		templatePath+"/"+name,
	))
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
	r.HandleFunc("/picture/{name}", handleDownloadPicture)

	r.HandleFunc("/directory/search", handleDirectorySearch)
	r.HandleFunc("/directory", handleDirectory)

	r.HandleFunc("/garage/key", handleGarageKey)
	r.HandleFunc("/garage/website", handleGarageWebsiteList)
	r.HandleFunc("/garage/website/new", handleGarageWebsiteNew)
	r.HandleFunc("/garage/website/b/{bucket}", handleGarageWebsiteInspect)

	r.HandleFunc("/invite/new_account", handleInviteNewAccount)
	r.HandleFunc("/invite/send_code", handleInviteSendCode)
	r.HandleFunc("/invitation/{code}", handleInvitationCode)

	r.HandleFunc("/admin/users", handleAdminUsers)
	r.HandleFunc("/admin/groups", handleAdminGroups)
	r.HandleFunc("/admin/mailing", handleAdminMailing)
	r.HandleFunc("/admin/mailing/{id}", handleAdminMailingList)
	r.HandleFunc("/admin/ldap/{dn}", handleAdminLDAP)
	r.HandleFunc("/admin/create/{template}/{super_dn}", handleAdminCreate)

	staticfiles := http.FileServer(http.Dir(staticPath))
	r.Handle("/static/{file:.*}", http.StripPrefix("/static/", staticfiles))

	// log.Printf("Starting HTTP server on %s", config.HttpBindAddr)
	err = http.ListenAndServe(config.HttpBindAddr, logRequest(r))
	if err != nil {
		log.Fatal("Cannot start http server: ", err)
	}
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
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
	Login  *LoginStatus
	BaseDN string
	Org    string
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	templateHome := getTemplate("home.html")

	login := checkLogin(w, r)
	if login == nil {
		return
	}

	data := &HomePageData{
		Login:  login,
		BaseDN: config.BaseDN,
		Org:    config.Org,
	}

	templateHome.Execute(w, data)
}
