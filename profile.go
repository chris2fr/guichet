package main

import (
	"net/http"
	"strings"
)

type ProfileTplData struct {
	Status       *LoginStatus
	ErrorMessage string
	Success      bool
	Mail         string
	DisplayName  string
	GivenName    string
	Surname      string
	Description  string
}

//ProfilePicture string
//Visibility     string

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
				data.ErrorMessage = "handleProfile : " + err.Error()
			} else {
				data.Success = true
			}
		}
		findUser, err := get(user, config, login.conn)
		if err != nil {
			data.ErrorMessage = "handleProfile : " + err.Error()
		}
		data.DisplayName = findUser.DisplayName
		data.GivenName = findUser.GivenName
		data.Surname = findUser.SN
		data.Description = findUser.Description
		data.Mail = findUser.Mail

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
						data.ErrorMessage = err.Error()
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
		// 	data.ErrorMessage = err.Error()
		// } else {
		// 	data.Success = true
		// }

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
			err := passwd(User{
				DN:       login.Info.DN,
				Password: password,
			}, config, login.conn)
			if err != nil {
				data.ErrorMessage = err.Error()
			} else {
				data.Success = true
			}
		}
	}

	templatePasswd.Execute(w, data)
}
