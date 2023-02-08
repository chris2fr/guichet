package main

import (
	"net/http"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

type ProfileTplData struct {
	Status         *LoginStatus
	ErrorMessage   string
	Success        bool
	Mail           string
	DisplayName    string
	GivenName      string
	Surname        string
	Visibility     string
	Description    string
	ProfilePicture string
}

func handleProfile(w http.ResponseWriter, r *http.Request) {
	templateProfile := getTemplate("profile.html")

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
	data.Visibility = login.UserEntry.GetAttributeValue(FIELD_NAME_DIRECTORY_VISIBILITY)
	data.Description = login.UserEntry.GetAttributeValue("description")
	data.ProfilePicture = login.UserEntry.GetAttributeValue(FIELD_NAME_PROFILE_PICTURE)

	if r.Method == "POST" {
		//5MB maximum size files
		r.ParseMultipartForm(5 << 20)

		data.DisplayName = strings.TrimSpace(strings.Join(r.Form["display_name"], ""))
		data.GivenName = strings.TrimSpace(strings.Join(r.Form["given_name"], ""))
		data.Surname = strings.TrimSpace(strings.Join(r.Form["surname"], ""))
		data.Description = strings.Trim(strings.Join(r.Form["description"], ""), "")
		visible := strings.TrimSpace(strings.Join(r.Form["visibility"], ""))
		if visible != "" {
			visible = "on"
		}
		data.Visibility = visible

		profilePicture, err := uploadProfilePicture(w, r, login)
		if err != nil {
			data.ErrorMessage = err.Error()
		}

		if profilePicture != "" {
			data.ProfilePicture = profilePicture
		}

		modify_request := ldap.NewModifyRequest(login.Info.DN, nil)
		modify_request.Replace("displayname", []string{data.DisplayName})
		modify_request.Replace("givenname", []string{data.GivenName})
		modify_request.Replace("sn", []string{data.Surname})
		modify_request.Replace("description", []string{data.Description})
		modify_request.Replace(FIELD_NAME_DIRECTORY_VISIBILITY, []string{data.Visibility})
		if data.ProfilePicture != "" {
			modify_request.Replace(FIELD_NAME_PROFILE_PICTURE, []string{data.ProfilePicture})
		}

		err = login.conn.Modify(modify_request)
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
	templatePasswd := getTemplate("passwd.html")

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
			pw, err := SSHAEncode(password)
			if err == nil {
				modify_request.Replace("userpassword", []string{pw})
				err := login.conn.Modify(modify_request)
				if err != nil {
					data.ErrorMessage = err.Error()
				} else {
					data.Success = true
				}
			} else {
				data.ErrorMessage = err.Error()
			}
		}
	}

	templatePasswd.Execute(w, data)
}
