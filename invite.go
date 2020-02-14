package main

import (
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

func checkInviterLogin(w http.ResponseWriter, r *http.Request) *LoginStatus {
	login := checkLogin(w, r)
	if login == nil {
		return nil
	}

	if !login.CanInvite {
		http.Error(w, "Not authorized to invite new users.", http.StatusUnauthorized)
		return nil
	}

	return login
}

type NewAccountData struct {
	Username    string
	DisplayName string
	GivenName   string
	Surname     string

	ErrorUsernameTaken    bool
	ErrorInvalidUsername  bool
	ErrorPasswordTooShort bool
	ErrorPasswordMismatch bool
	ErrorMessage          string
	WarningMessage        string
	Success               bool
}

func handleInviteNewAccount(w http.ResponseWriter, r *http.Request) {
	templateInviteNewAccount := template.Must(template.ParseFiles("templates/layout.html", "templates/invite_new_account.html"))

	login := checkInviterLogin(w, r)
	if login == nil {
		return
	}

	data := &NewAccountData{}

	if r.Method == "POST" {
		r.ParseForm()

		data.Username = strings.TrimSpace(strings.Join(r.Form["username"], ""))
		data.DisplayName = strings.TrimSpace(strings.Join(r.Form["displayname"], ""))
		data.GivenName = strings.TrimSpace(strings.Join(r.Form["givenname"], ""))
		data.Surname = strings.TrimSpace(strings.Join(r.Form["surname"], ""))

		password1 := strings.Join(r.Form["password"], "")
		password2 := strings.Join(r.Form["password2"], "")

		tryCreateAccount(login.conn, data, password1, password2)
	}

	templateInviteNewAccount.Execute(w, data)
}

func tryCreateAccount(l *ldap.Conn, data *NewAccountData, pass1 string, pass2 string) {
	// Check if username is correct
	if match, err := regexp.MatchString("^[a-zA-Z0-9._-]+$", data.Username); !(err == nil && match) {
		data.ErrorInvalidUsername = true
	}

	// Check if user exists
	userDn := config.UserNameAttr + "=" + data.Username + "," + config.UserBaseDN
	searchRq := ldap.NewSearchRequest(
		userDn,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		"(objectclass=*)",
		[]string{"dn"},
		nil)

	sr, err := l.Search(searchRq)
	if err != nil {
		data.ErrorMessage = err.Error()
		return
	}

	if len(sr.Entries) > 0 {
		data.ErrorUsernameTaken = true
		return
	}

	// Check that password is long enough
	if len(pass1) < 8 {
		data.ErrorPasswordTooShort = true
		return
	}

	if pass1 != pass2 {
		data.ErrorPasswordMismatch = true
		return
	}

	// Actually create user
	req := ldap.NewAddRequest(userDn, nil)
	req.Attribute("objectclass", []string{"inetOrgPerson", "organizationalPerson", "person", "top"})
	req.Attribute("structuralobjectclass", []string{"inetOrgPerson"})
	req.Attribute("userpassword", []string{SSHAEncode([]byte(pass1))})
	if len(data.DisplayName) > 0 {
		req.Attribute("displayname", []string{data.DisplayName})
	}
	if len(data.GivenName) > 0 {
		req.Attribute("givenname", []string{data.GivenName})
	}
	if len(data.Surname) > 0 {
		req.Attribute("sn", []string{data.Surname})
	}
	if len(config.InvitedMailFormat) > 0 {
		email := strings.ReplaceAll(config.InvitedMailFormat, "{}", data.Username)
		req.Attribute("mail", []string{email})
	}

	err = l.Add(req)
	if err != nil {
		data.ErrorMessage = err.Error()
		return
	}

	for _, group := range config.InvitedAutoGroups {
		req := ldap.NewModifyRequest(group, nil)
		req.Add("member", []string{userDn})
		err = l.Modify(req)
		if err != nil {
			data.WarningMessage += fmt.Sprintf("Cannot add to %s: %s\n", group, err.Error())
		}
	}

	data.Success = true
}

func handleInviteSendCode(w http.ResponseWriter, r *http.Request) {
}
