package main

import (
	b64 "encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

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
		data.Common.LoggedIn = false
		http.Redirect(w, r, "/", http.StatusFound)

		// templatePasswd.Execute(w, data)
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
