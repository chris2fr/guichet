package views

import (
	b64 "encoding/base64"
	"fmt"
	"guichet/models"
	"log"
	"net/http"
	"strings"

	// "github.com/go-ldap/ldap/v3"
	"github.com/gorilla/mux"
)

func HandleLostPassword(w http.ResponseWriter, r *http.Request) {
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
		data.SearchQuery = strings.TrimSpace(strings.Join(r.Form["searchquery"], ""))
		ldapNewConn, err := models.OpenNewUserLdap(&config)
		if err != nil {
			log.Printf(fmt.Sprintf("HandleLostPassword 99 : %v %v", err, ldapNewConn))
			data.Common.ErrorMessage = err.Error()
		}
		if err != nil {
			log.Printf(fmt.Sprintf("HandleLostPassword 104 : %v %v", err, ldapNewConn))
			data.Common.ErrorMessage = err.Error()
		} else {
			// err = ldapConn.Bind(config.NewUserDN, config.NewUserPassword)
			if err != nil {
				log.Printf(fmt.Sprintf("HandleLostPassword 109 : %v %v", err, ldapNewConn))
				data.Common.ErrorMessage = err.Error()
			} else {
				data.Common.Success = true
			}
		}
		err = models.PasswordLost(data.SearchQuery, &config, ldapNewConn)
		ldapNewConn.Close()
	}
	data.Common.CanAdmin = false
	// templateLostPasswordPage.Execute(w, data)
	execTemplate(w, templateLostPasswordPage, data.Common, NestedLoginTplData{}, data)
}

func HandleFoundPassword(w http.ResponseWriter, r *http.Request) {
	templateFoundPasswordPage := getTemplate("passwd.html")
	data := PasswdTplData{
		Common: NestedCommonTplData{
			CanAdmin: false,
			LoggedIn: false},
	}
	code := mux.Vars(r)["code"]
	// code = strings.TrimSpace(strings.Join([]string{code}, ""))
	newCode, _ := b64.URLEncoding.DecodeString(code)
	ldapNewConn, err := models.OpenNewUserLdap(&config)
	if err != nil {
		log.Printf("HandleFoundPassword OpenNewUserLdap(config) : %v", err)
		data.Common.ErrorMessage = err.Error()
	}
	codeArray := strings.Split(string(newCode), ";")
	user := models.User{
		UID:      codeArray[0],
		Password: codeArray[1],
		DN:       "uid=" + codeArray[0] + "," + config.InvitationBaseDN,
	}
	user.SeeAlso, err = models.PasswordFound(user, &config, ldapNewConn)
	if err != nil {
		log.Printf("PasswordFound(models.User, config, ldapConn) %v", err)
		log.Printf("PasswordFound(models.User, config, ldapConn) %v", user)
		log.Printf("PasswordFound(models.User, config, ldapConn) %v", ldapNewConn)
		data.Common.ErrorMessage = err.Error()
	} else {
		log.Printf("PasswordFound OK %v", user)
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
			err := models.PassWD(models.User{
				DN:       user.SeeAlso,
				Password: password,
			}, &config, ldapNewConn)
			if err != nil {
				data.Common.ErrorMessage = err.Error()
				log.Printf("PasswordFound KO %v", user.SeeAlso)
			} else {
				data.Common.Success = true
				log.Printf("PasswordFound OK %v", user.SeeAlso)
			}
		}
	}
	data.Common.CanAdmin = false
	// templateFoundPasswordPage.Execute(w, data)
	execTemplate(w, templateFoundPasswordPage, data.Common, data.Login, data)
	ldapNewConn.Close()
}

func HandlePasswd(w http.ResponseWriter, r *http.Request) {
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
			err := models.PassWD(models.User{
				DN:       login.Info.DN,
				Password: password,
			}, &config, login.conn)
			if err != nil {
				data.Common.ErrorMessage = err.Error()
				log.Printf("PasswordFound KO %v", login.Info.DN)
			} else {
				data.Common.Success = true
				log.Printf("PasswordFound OK %v", login.Info.DN)
			}
		}
	}
	data.Common.CanAdmin = false
	// templatePasswd.Execute(w, data)
	execTemplate(w, templatePasswd, data.Common, data.Login, data)
	login.conn.Close()
}
