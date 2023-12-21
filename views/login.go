/*
login Handles login and current-user verification
*/

package views

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"guichet/models"

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

		newUserLdapConn, _ := models.OpenNewUserLdap(config)


		searchRequest := ldap.NewSearchRequest(
			config.UserBaseDN,
			ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
			fmt.Sprintf("(|(cn=%s)(uid=%s)(mail=%s))",username,username,username),
			[]string{
				"dn",
			},
			nil)
		//Transform the researh's result in a correct struct to send JSON
		searchRes, err := newUserLdapConn.Search(searchRequest)
		if err != nil {
			log.Printf("doLogin search : %v %v", err, newUserLdapConn)
			log.Printf("doLogin search : %v", searchRequest)
			log.Printf("doLogin search : %v", searchRes)
			// log.Printf("PasswordLost search: %v", user)
			return nil, err
		}
		if len(searchRes.Entries) == 0 {
			log.Printf("Il n'y a pas d'utilisateur qui correspond %v", searchRequest)
			// return errors.New("Il n'y a pas d'utilisateur qui correspond")
		}
		user_dn := searchRes.Entries[0].GetAttributeValue("dn")
		// user_dn := fmt.Sprintf("%s=%s,%s", config.UserNameAttr, username, config.UserBaseDN)
		// log.Printf("%v", user_dn)
		// log.Printf("%v", username)
		// if strings.EqualFold(username, config.AdminAccount) {
		////////////////////////////// TODO
		// if strings.EqualFold(username, config.AdminAccount) {
		// 	user_dn = username
		// }
		////////////////////////////// /TODO
		err = l.Bind(user_dn, password)
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
