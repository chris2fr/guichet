package main

import (
	b64 "encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func handleProfile(w http.ResponseWriter, r *http.Request) {
	templateProfile := getTemplate("profile.html")

	login := checkLogin(w, r)
	if login == nil {
		templatePasswd := getTemplate("passwd.html")
		templatePasswd.Execute(w, PasswdTplData{

			Common: NestedCommonTplData{
				CanAdmin: false,
				LoggedIn: false},
		})
		return
	}

	data := &ProfileTplData{
		Login: NestedLoginTplData{
			Status: login,
			Login:  login,
		},
		Common: NestedCommonTplData{
			CanAdmin:     login.Common.CanAdmin,
			LoggedIn:     true,
			ErrorMessage: "",
			Success:      false,
		},
	}

	data.Mail = login.UserEntry.GetAttributeValue("mail")
	data.DisplayName = login.UserEntry.GetAttributeValue("displayName")
	data.GivenName = login.UserEntry.GetAttributeValue("givenName")
	data.Surname = login.UserEntry.GetAttributeValue("sn")
	//	data.Visibility = login.UserEntry.GetAttributeValue(FIELD_NAME_DIRECTORY_VISIBILITY)
	data.Description = login.UserEntry.GetAttributeValue("description")
	//data.ProfilePicture = login.UserEntry.GetAttributeValue(FIELD_NAME_PROFILE_PICTURE)

	if r.Method == "POST" {
		//5MB maximum size files
		r.ParseMultipartForm(5 << 20)
		user := User{
			DN: login.Info.DN,
			// CN: ,
			GivenName:   strings.TrimSpace(strings.Join(r.Form["given_name"], "")),
			DisplayName: strings.TrimSpace(strings.Join(r.Form["display_name"], "")),
			Mail:        strings.TrimSpace(strings.Join(r.Form["mail"], "")),
			SN:          strings.TrimSpace(strings.Join(r.Form["surname"], "")),
			//UID: ,
			Description: strings.TrimSpace(strings.Join(r.Form["description"], "")),
			// Password: ,
		}

		if user.DisplayName != "" {
			err := modify(user, config, login.conn)
			if err != nil {
				data.Common.ErrorMessage = "handleProfile : " + err.Error()
			} else {
				data.Common.Success = true
			}
		}
		findUser, err := get(user, config, login.conn)
		if err != nil {
			data.Common.ErrorMessage = "handleProfile : " + err.Error()
		}
		data.DisplayName = findUser.DisplayName
		data.GivenName = findUser.GivenName
		data.Surname = findUser.SN
		data.Description = findUser.Description
		data.Mail = findUser.Mail
		data.Common.LoggedIn = false

		/*
					visible := strings.TrimSpace(strings.Join(r.Form["visibility"], ""))
					if visible != "" {
						visible = "on"
					} else {
			      visible = "off"
			    }
					data.Visibility = visible
		*/
		/*
					profilePicture, err := uploadProfilePicture(w, r, login)
					if err != nil {
						data.Common.ErrorMessage = err.Error()
					}
			    if profilePicture != "" {
						data.ProfilePicture = profilePicture
					}
		*/

		//modify_request.Replace(FIELD_NAME_DIRECTORY_VISIBILITY, []string{data.Visibility})
		//modify_request.Replace(FIELD_NAME_DIRECTORY_VISIBILITY, []string{"on"})
		//if data.ProfilePicture != "" {
		//		modify_request.Replace(FIELD_NAME_PROFILE_PICTURE, []string{data.ProfilePicture})
		//	}

		// err := login.conn.Modify(modify_request)
		// log.Printf(fmt.Sprintf("Profile:079: %v",modify_request))
		// log.Printf(fmt.Sprintf("Profile:079: %v",err))
		// log.Printf(fmt.Sprintf("Profile:079: %v",data))
		// if err != nil {
		// 	data.Common.ErrorMessage = err.Error()
		// } else {
		// 	data.Common.Success = true
		// }

	}

	templateProfile.Execute(w, data)
}

func handleFoundPassword(w http.ResponseWriter, r *http.Request) {
	templateFoundPasswordPage := getTemplate("passwd.html")
	data := PasswdTplData{
		Common: NestedCommonTplData{
			CanAdmin: false,
			LoggedIn: false},
	}
	code := mux.Vars(r)["code"]
	// code = strings.TrimSpace(strings.Join([]string{code}, ""))
	newCode, _ := b64.URLEncoding.DecodeString(code)
	ldapConn, err := openNewUserLdap(config)
	if err != nil {
		log.Printf(fmt.Sprint("handleFoundPassword / openNewUserLdap / %v", err))
		data.Common.ErrorMessage = err.Error()
	}
	codeArray := strings.Split(string(newCode), ";")
	user := User{
		UID:      codeArray[0],
		Password: codeArray[1],
		DN:       "uid=" + codeArray[0] + ",ou=invitations,dc=resdigita,dc=org",
	}
	user.SeeAlso, err = passwordFound(user, config, ldapConn)
	if err != nil {
		log.Printf("handleFoundPassword / passwordFound %v", err)
		log.Printf("handleFoundPassword / passwordFound %v", err)
		data.Common.ErrorMessage = err.Error()
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
			err := passwd(User{
				DN:       user.SeeAlso,
				Password: password,
			}, config, ldapConn)
			if err != nil {
				data.Common.ErrorMessage = err.Error()
			} else {
				data.Common.Success = true
			}
		}
	}
	data.Common.CanAdmin = false
	templateFoundPasswordPage.Execute(w, data)
}

func handlePasswd(w http.ResponseWriter, r *http.Request) {
	templatePasswd := getTemplate("passwd.html")
	data := &PasswdTplData{
		Common: NestedCommonTplData{
			CanAdmin:     false,
			LoggedIn:     true,
			ErrorMessage: "",
			Success:      false,
		},
	}

	login := checkLogin(w, r)
	if login == nil {
		templatePasswd.Execute(w, data)
		return
	}
	data.Login.Status = login
	data.Common.CanAdmin = login.Common.CanAdmin

	if r.Method == "POST" {
		r.ParseForm()

		password := strings.Join(r.Form["password"], "")
		password2 := strings.Join(r.Form["password2"], "")

		if len(password) < 8 {
			data.TooShortError = true
		} else if password2 != password {
			data.NoMatchError = true
		} else {
			err := passwd(User{
				DN:       login.Info.DN,
				Password: password,
			}, config, login.conn)
			if err != nil {
				data.Common.ErrorMessage = err.Error()
			} else {
				data.Common.Success = true
			}
		}
	}
	data.Common.CanAdmin = false
	templatePasswd.Execute(w, data)
}
