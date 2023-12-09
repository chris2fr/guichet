package views

import (
	// b64 "encoding/base64"
	"fmt"
	// "log"
	"guichet/models"
	"log"
	"net/http"
	"strings"

	"github.com/go-ldap/ldap/v3"
	// "github.com/gorilla/mux"
)

func HandleUserWait(w http.ResponseWriter, r *http.Request) {
	templateUser := getTemplate("user/wait.html")
	templateUser.Execute(w, HomePageData{
		Common: NestedCommonTplData{
			CanAdmin: false,
			LoggedIn: false,
		},
	})
}

func HandleUserMail(w http.ResponseWriter, r *http.Request) {
	login := checkLogin(w, r)
	if login == nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	email := r.FormValue("email")
	action := r.FormValue("action")
	var err error
	if action == "Add" {
		// Add the new mail value to the entry
		modifyRequest := ldap.NewModifyRequest(login.Info.DN, nil)
		modifyRequest.Add("mail", []string{email})

		err = login.conn.Modify(modifyRequest)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error adding the email: %v", modifyRequest), http.StatusInternalServerError)
			return
		}
	} else if action == "Delete" {
		modifyRequest := ldap.NewModifyRequest(login.Info.DN, nil)
		modifyRequest.Delete("mail", []string{email})

		log.Printf("HandleUserMail %v", modifyRequest)
		err = login.conn.Modify(modifyRequest)
		if err != nil {
			log.Printf("HandleUserMail DeleteMail %v", err)
			http.Error(w, fmt.Sprintf("Error deleting the email: %s", err), http.StatusInternalServerError)
			return
		}
	}

	message := fmt.Sprintf("Mail value updated successfully to: %s", email)
	http.Redirect(w, r, "/user?message="+message, http.StatusSeeOther)

}

func toInteger(index string) {
	panic("unimplemented")
}

func HandleUser(w http.ResponseWriter, r *http.Request) {
	templateUser := getTemplate("user.html")

	login := checkLogin(w, r)
	if login == nil {
		http.Redirect(w, r, "/", http.StatusFound)
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
	data.OtherMailbox = login.UserEntry.GetAttributeValue("carLicense")
	data.MailValues = login.UserEntry.GetAttributeValues("mail")
	//	data.Visibility = login.UserEntry.GetAttributeValue(FIELD_NAME_DIRECTORY_VISIBILITY)
	data.Description = login.UserEntry.GetAttributeValue("description")
	//data.ProfilePicture = login.UserEntry.GetAttributeValue(FIELD_NAME_PROFILE_PICTURE)

	if r.Method == "POST" {
		//5MB maximum size files
		r.ParseMultipartForm(5 << 20)
		user := models.User{
			DN:           login.Info.DN,
			GivenName:    strings.TrimSpace(strings.Join(r.Form["given_name"], "")),
			DisplayName:  strings.TrimSpace(strings.Join(r.Form["display_name"], "")),
			Mail:         strings.TrimSpace(strings.Join(r.Form["mail"], "")),
			SN:           strings.TrimSpace(strings.Join(r.Form["surname"], "")),
			OtherMailbox: strings.TrimSpace(strings.Join(r.Form["othermailbox"], "")),
			Description:  strings.TrimSpace(strings.Join(r.Form["description"], "")),
			// Password: ,
			//UID: ,
			// CN: ,
		}

		if user.DisplayName != "" {
			err := models.ModifyUser(user, &config, login.conn)
			if err != nil {
				data.Common.ErrorMessage = "HandleUser : " + err.Error()
			} else {
				data.Common.Success = true
			}
		}
		findUser, err := models.GetUser(user, &config, login.conn)
		if err != nil {
			data.Common.ErrorMessage = "HandleUser : " + err.Error()
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

	log.Printf("HandleUser : %v", data)

	// templateUser.Execute(w, data)
	execTemplate(w, templateUser, data.Common, data.Login, data)
}
