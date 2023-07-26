package main

import (
	b64 "encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"

	// "github.com/go-ldap/ldap/v3"
	"github.com/gorilla/mux"
)

func handleLostPassword(w http.ResponseWriter, r *http.Request) {
	templateLostPasswordPage := getTemplate("passwd/lost.html")
	if checkLogin(w, r) != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

	data := PasswordLostData{
		Common: NestedCommonTplData{
			CanAdmin: false,
			LoggedIn: false},
	}

	if r.Method == "POST" {
		r.ParseForm()
		data.Username = strings.TrimSpace(strings.Join(r.Form["username"], ""))
		data.Mail = strings.TrimSpace(strings.Join(r.Form["mail"], ""))
		data.OtherMailbox = strings.TrimSpace(strings.Join(r.Form["othermailbox"], ""))
		user := User{
			CN:           strings.TrimSpace(strings.Join(r.Form["username"], "")),
			UID:          strings.TrimSpace(strings.Join(r.Form["username"], "")),
			Mail:         strings.TrimSpace(strings.Join(r.Form["mail"], "")),
			OtherMailbox: strings.TrimSpace(strings.Join(r.Form["othermailbox"], "")),
		}
		ldapNewConn, err := openNewUserLdap(config)
		if err != nil {
			log.Printf(fmt.Sprintf("handleLostPassword 99 : %v %v", err, ldapNewConn))
			data.Common.ErrorMessage = err.Error()
		}
		if err != nil {
			log.Printf(fmt.Sprintf("handleLostPassword 104 : %v %v", err, ldapNewConn))
			data.Common.ErrorMessage = err.Error()
		} else {
			// err = ldapConn.Bind(config.NewUserDN, config.NewUserPassword)
			if err != nil {
				log.Printf(fmt.Sprintf("handleLostPassword 109 : %v %v", err, ldapNewConn))
				data.Common.ErrorMessage = err.Error()
			} else {
				data.Common.Success = true
			}
		}
		err = passwordLost(user, config, ldapNewConn)
	}
	data.Common.CanAdmin = false
	templateLostPasswordPage.Execute(w, data)
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
	ldapNewConn, err := openNewUserLdap(config)
	if err != nil {
		log.Printf("handleFoundPassword openNewUserLdap(config) : %v", err)
		data.Common.ErrorMessage = err.Error()
	}
	codeArray := strings.Split(string(newCode), ";")
	user := User{
		UID:      codeArray[0],
		Password: codeArray[1],
		DN:       "uid=" + codeArray[0] + "," + config.InvitationBaseDN,
	}
	user.SeeAlso, err = passwordFound(user, config, ldapNewConn)
	if err != nil {
		log.Printf("passwordFound(user, config, ldapConn) %v", err)
		log.Printf("passwordFound(user, config, ldapConn) %v", user)
		log.Printf("passwordFound(user, config, ldapConn) %v", ldapNewConn)
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
			}, config, ldapNewConn)
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
		http.Redirect(w, r, "/", http.StatusFound)
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
