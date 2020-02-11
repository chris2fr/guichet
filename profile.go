package main

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

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

	data.Mail = login.UserEntry.GetAttributeValue("mail")
	data.DisplayName = login.UserEntry.GetAttributeValue("displayname")
	data.GivenName = login.UserEntry.GetAttributeValue("givenname")
	data.Surname = login.UserEntry.GetAttributeValue("sn")

	if r.Method == "POST" {
		r.ParseForm()

		data.DisplayName = strings.TrimSpace(strings.Join(r.Form["display_name"], ""))
		data.GivenName = strings.TrimSpace(strings.Join(r.Form["given_name"], ""))
		data.Surname = strings.TrimSpace(strings.Join(r.Form["surname"], ""))

		modify_request := ldap.NewModifyRequest(login.Info.DN, nil)
		modify_request.Replace("displayname", []string{data.DisplayName})
		modify_request.Replace("givenname", []string{data.GivenName})
		modify_request.Replace("sn", []string{data.Surname})

		err := login.conn.Modify(modify_request)
		if err != nil {
			data.ErrorMessage = err.Error()
		} else {
			data.Success = true
		}
	}

	templateProfile.Execute(w, data)
}

type PasswdTplData struct {
	Status        *LoginStatus
	ErrorMessage  string
	TooShortError bool
	NoMatchError  bool
	Success       bool
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

		if len(password) < 8 {
			data.TooShortError = true
		} else if password2 != password {
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
