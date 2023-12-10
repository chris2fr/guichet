/*
login Handles login and current-user verification
*/

package views

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-ldap/ldap/v3"
)




func HandleLogout(w http.ResponseWriter, r *http.Request) {

	err := logout(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func HandleLogin(w http.ResponseWriter, r *http.Request) (*LoginInfo, error) {
	templateLogin := getTemplate("login.html")

	if r.Method == "POST" {
		// log.Printf("%v", "Parsing Form HandleLogin")
		r.ParseForm()

		username := strings.Join(r.Form["username"], "")
		password := strings.Join(r.Form["password"], "")
		l, _ := ldapOpen(w)

		user_dn := fmt.Sprintf("%s=%s,%s", config.UserNameAttr, username, config.UserBaseDN)

		// log.Printf("%v", user_dn)
		// log.Printf("%v", username)
	
		if strings.EqualFold(username, config.AdminAccount) {
			user_dn = username
		}
		
	
		err := l.Bind(user_dn, password)
		if err != nil {
			log.Printf("DoLogin : %v", err)
			log.Printf("DoLogin : %v", user_dn)
			l.Close()
			return nil, err
		}
	

	
	
	
	
	
	// func encodePassword(inPassword string) (string, error) {
	// 	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	// 	return utf16.NewEncoder().String("\"" + inPassword + "\"")
	// 	// if err != nil {
	// 	// 	log.Printf("Error encoding password:  %s", err)
	// 	// 	return err
	// 	// }
	
	// }
	

		// log.Printf("%v", LoginInfo)
		if err != nil {
			data := &LoginFormData{
				Username: username,
				Common: NestedCommonTplData{
					CanAdmin:  false,
					CanInvite: true,
					LoggedIn:  false,
				},
			}
			if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
				data.WrongPass = true
			} else if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
				data.WrongUser = true
			} else {
				log.Printf("%v", err)
				log.Printf("%v", user_dn)
				log.Printf("%v", username)
				data.Common.ErrorMessage = err.Error()
			}
		}
		// Successfully logged in, save it to session
		session, err := GuichetSessionStore.Get(r, SESSION_NAME)
		if err != nil {
			session, _ = GuichetSessionStore.New(r, SESSION_NAME)
		}
		session.Values["login_username"] = username
		session.Values["login_password"] = password
		session.Values["login_dn"] = user_dn
	
		err = session.Save(r, w)
		if err != nil {
			log.Printf("DoLogin Session Save: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			l.Close()
			return nil, err
		}

		
		// // templateLogin.Execute(w, data)
		// execTemplate(w, templateLogin, data.Common, NestedLoginTplData{}, *config, data)

		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)

	} else if r.Method == "GET" {
		execTemplate(w, templateLogin, NestedCommonTplData{
			CanAdmin:  false,
			CanInvite: true,
			LoggedIn:  false}, NestedLoginTplData{}, LoginFormData{
			Common: NestedCommonTplData{
				CanAdmin:  false,
				CanInvite: true,
				LoggedIn:  false}})
		// templateLogin.Execute(w, LoginFormData{
		// 	Common: NestedCommonTplData{
		// 		CanAdmin:  false,
		// 		CanInvite: true,
		// 		LoggedIn:  false}})
		return nil, nil
	} else {
		http.Error(w, "Unsupported method", http.StatusBadRequest)
		return nil, nil
	}
	// execTemplate(w, templateLogin, data.Common, NestedLoginTplData{}, *config, data)
	return nil, nil
}

// func NotDoLogin(w http.ResponseWriter, r *http.Request, username string, user_dn string, password string) (*LoginInfo, error) {
	

// }
