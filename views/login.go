/*
login Handles login and current-user verification
*/

package views

import (
	"fmt"
	"guichet/models"
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

func DetermineLDAPUserDN (identity string) (string, error) {
	newUserLdapConn, err := models.OpenNewUserLdap(&config)
		if err != nil {
			log.Printf("doLogin search : %v %v", err, newUserLdapConn)
			return "", err
		}

		searchRequest := ldap.NewSearchRequest(
			config.UserBaseDN,
			ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
			fmt.Sprintf("(|(cn=%s)(uid=%s))",identity,identity),
			[]string{
				"objectClass",
			},
			nil)
		//Transform the researh's result in a correct struct to send JSON
		searchRes, err := newUserLdapConn.Search(searchRequest)
		if err != nil {
			log.Printf("doLogin search : %v %v", err, newUserLdapConn)
			log.Printf("doLogin search : %v", searchRequest)
			log.Printf("doLogin search : %v", searchRes)
			// log.Printf("PasswordLost search: %v", user)
			return "", err
		}
		if len(searchRes.Entries) == 0 {
			log.Printf("Il n'y a pas d'utilisateur qui correspond %v", searchRequest)
			return "", err // errors.New("Il n'y a pas d'utilisateur qui correspond")
		}
		newUserLdapConn.Close()
		return searchRes.Entries[0].DN, nil
}

func BindAsLDAPUser (user_dn string, password string) (error) {
	ldapConn, _ := ldapSimpleOpen()
	err := ldapConn.Bind(user_dn, password)
	if err != nil {
		log.Printf("DoLogin : %v", err)
		log.Printf("DoLogin : %v", user_dn)
		ldapConn.Close()
		// data.Common.ErrorMessage = err.Error()
	  return err
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
	ldapConn.Close()
	return nil
}

func HandleLDAPLogin(w http.ResponseWriter, r *http.Request) (*LoginInfo, error) {
	// templateLogin := getTemplate("login.html")
	if r.Method == "POST" {
		// log.Printf("%v", "Parsing Form HandleLogin")
		r.ParseForm()

		user_dn, err := DetermineLDAPUserDN(strings.Join(r.Form["identity"], ""))
		if err != nil {
			return nil, err
		}
		err = BindAsLDAPUser(user_dn, strings.Join(r.Form["password"], ""))
		if err != nil {
			return nil, err
		}
		// if err != nil {
		// 	data := &LoginFormData{
		// 		Username: strings.Join(r.Form["identity"], ""),
		// 		Common: NestedCommonTplData{
		// 			CanAdmin:  false,
		// 			CanInvite: true,
		// 			LoggedIn:  false,
		// 		},
		// 	}
		// 	if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
		// 		data.WrongPass = true
		// 	} else if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
		// 		data.WrongUser = true
		// 	} else {
		// 		log.Printf("%v", err)
		// 		log.Printf("%v", user_dn)
		// 		log.Printf("%v", strings.Join(r.Form["identity"], ""))
		// 		data.Common.ErrorMessage = err.Error()
		// 	}
		// }
	
		
		// user_dn := fmt.Sprintf("%s=%s,%s", config.UserNameAttr, username, config.UserBaseDN)
		// log.Printf("%v", user_dn)
		// log.Printf("%v", username)
		// if strings.EqualFold(username, config.AdminAccount) {
		////////////////////////////// TODO
		// if strings.EqualFold(username, config.AdminAccount) {
		// 	user_dn = username
		// }
		////////////////////////////// /TODO
		// password := strings.Join(r.Form["password"], "")
		



		// Successfully logged in, save it to session
		session, err := GuichetSessionStore.Get(r, SESSION_NAME)
		if err != nil {
			session, _ = GuichetSessionStore.New(r, SESSION_NAME)
		}
		// Todo, better define login_username
		session.Values["login_username"] = strings.Join(r.Form["identity"], "")
		// Todo This is not great at all
		session.Values["login_password"] = strings.Join(r.Form["password"], "")
		session.Values["login_dn"] = user_dn
	
		err = session.Save(r, w)
		if err != nil {
			log.Printf("DoLogin Session Save: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, err
		}

		
		// // templateLogin.Execute(w, data)
		// execTemplate(w, templateLogin, data.Common, NestedLoginTplData{}, *config, data)

		// http.Redirect(w, r, "/", http.StatusTemporaryRedirect)

	// } else if r.Method == "GET" {
	// 	execTemplate(w, templateLogin, NestedCommonTplData{
	// 		CanAdmin:  false,
	// 		CanInvite: true,
	// 		LoggedIn:  false}, NestedLoginTplData{}, LoginFormData{
	// 		Common: NestedCommonTplData{
	// 			CanAdmin:  false,
	// 			CanInvite: true,
	// 			LoggedIn:  false}})
	// 	// templateLogin.Execute(w, LoginFormData{
	// 	// 	Common: NestedCommonTplData{
	// 	// 		CanAdmin:  false,
	// 	// 		CanInvite: true,
	// 	// 		LoggedIn:  false}})
	// 	return nil, nil
	// } else {
	// 	http.Error(w, "Unsupported method", http.StatusBadRequest)
	// 	return nil, nil
	}
	// execTemplate(w, templateLogin, data.Common, NestedLoginTplData{}, *config, data)
	log.Printf("DoLogin: %v", "LDAP Logged In")
	return nil, nil
}

// func NotDoLogin(w http.ResponseWriter, r *http.Request, username string, user_dn string, password string) (*LoginInfo, error) {
	

// }
