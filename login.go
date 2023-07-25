/*
login handles login and current-user verification
*/

package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

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

func (login *LoginStatus) WelcomeName() string {
	ret := login.UserEntry.GetAttributeValue("givenName")
	if ret == "" {
		ret = login.UserEntry.GetAttributeValue("displayName")
	}
	if ret == "" {
		ret = login.Info.Username
	}
	return ret
}

func handleLogout(w http.ResponseWriter, r *http.Request) {

	err := logout(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

type LoginFormData struct {
	Username     string
	WrongUser    bool
	WrongPass    bool
	ErrorMessage string
	LoggedIn     bool
	CanAdmin     bool
}

func handleLogin(w http.ResponseWriter, r *http.Request) *LoginInfo {
	templateLogin := getTemplate("login.html")

	loginStatus := checkLogin(w, r)
	if loginStatus != nil {
		handleHome(w, r)
		return loginStatus.Info
	}

	if r.Method == "GET" {
		templateLogin.Execute(w, LoginFormData{CanAdmin: false})
		return nil
	} else if r.Method == "POST" {
		// log.Printf("%v", "Parsing Form handleLogin")
		r.ParseForm()

		username := strings.Join(r.Form["username"], "")
		password := strings.Join(r.Form["password"], "")
		user_dn := fmt.Sprintf("%s=%s,%s", config.UserNameAttr, username, config.UserBaseDN)

		// log.Printf("%v", user_dn)
		// log.Printf("%v", username)

		if strings.EqualFold(username, config.AdminAccount) {
			user_dn = username
		}
		loginInfo, err := doLogin(w, r, username, user_dn, password)
		// log.Printf("%v", loginInfo)
		if err != nil {
			data := &LoginFormData{
				Username: username,
				LoggedIn: false,
				CanAdmin: false,
			}
			if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
				data.WrongPass = true
			} else if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
				data.WrongUser = true
			} else {
				log.Printf("%v", err)
				log.Printf("%v", user_dn)
				log.Printf("%v", username)
				data.ErrorMessage = err.Error()
			}
			templateLogin.Execute(w, data)
		}
		return loginInfo

	} else {
		http.Error(w, "Unsupported method", http.StatusBadRequest)
		return nil
	}
}

func doLogin(w http.ResponseWriter, r *http.Request, username string, user_dn string, password string) (*LoginInfo, error) {
	l, _ := ldapOpen(w)

	err := l.Bind(user_dn, password)
	if err != nil {
		log.Printf("doLogin : %v", err)
		log.Printf("doLogin : %v", user_dn)
		return nil, err
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
		log.Printf("doLogin Session Save: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}

	LoginInfo := LoginInfo{
		DN:       user_dn,
		Username: username,
		Password: password,
	}

	return &LoginInfo, nil
}
