/*
login handles login and current-user verification
*/

package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

type LoginInfo struct {
	Username string
	DN       string
	Password string
}

type LoginStatus struct {
	Info      *LoginInfo
	conn      *ldap.Conn
	UserEntry *ldap.Entry
	CanAdmin  bool
	CanInvite bool
}

func (login *LoginStatus) WelcomeName() string {
	ret := login.UserEntry.GetAttributeValue("givenName")
	if ret == "" {
		ret = login.UserEntry.GetAttributeValue("displayName")
	}
	if ret == "" {
		ret = login.Info.Username
	}
	return ret
}

func checkLogin(w http.ResponseWriter, r *http.Request) *LoginStatus {
	var login_info *LoginInfo

	session, err := store.Get(r, SESSION_NAME)
	if err == nil {
		username, ok := session.Values["login_username"]
		password, ok2 := session.Values["login_password"]
		user_dn, ok3 := session.Values["login_dn"]

		if ok && ok2 && ok3 {
			login_info = &LoginInfo{
				DN:       user_dn.(string),
				Username: username.(string),
				Password: password.(string),
			}
		}
	}

	if login_info == nil {
		login_info = handleLogin(w, r)
		if login_info == nil {
			return nil
		}
	}

	l := ldapOpen(w)
	if l == nil {
		return nil
	}

	err = l.Bind(login_info.DN, login_info.Password)
	if err != nil {
		delete(session.Values, "login_username")
		delete(session.Values, "login_password")
		delete(session.Values, "login_dn")

		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil
		}
		return checkLogin(w, r)
	}

	loginStatus := &LoginStatus{
		Info: login_info,
		conn: l,
	}

	requestKind := "(objectClass=organizationalPerson)"
	if strings.EqualFold(login_info.DN, config.AdminAccount) {
		requestKind = "(objectclass=*)"
	}
	searchRequest := ldap.NewSearchRequest(
		login_info.DN,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		requestKind,
		[]string{
			"dn",
			"displayname",
			"givenname",
			"sn",
			"mail",
			"cn",
			"memberof",
			"description",
			"garage_s3_access_key",
		},
		nil)
	//			FIELD_NAME_DIRECTORY_VISIBILITY,
	//			FIELD_NAME_PROFILE_PICTURE,

	sr, err := l.Search(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	if len(sr.Entries) != 1 {
		http.Error(w, fmt.Sprintf("Unable to find entry for %s", login_info.DN), http.StatusInternalServerError)
		return nil
	}

	loginStatus.UserEntry = sr.Entries[0]

	loginStatus.CanAdmin = strings.EqualFold(loginStatus.Info.DN, config.AdminAccount)
	loginStatus.CanInvite = false

	groups := []EntryName{}
	searchRequest = ldap.NewSearchRequest(
		config.GroupBaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=groupOfNames)(member=%s))", login_info.DN),
		[]string{"dn", "displayName", "cn", "description"},
		nil)
	// // log.Printf(fmt.Sprintf("708: %v",searchRequest))
	sr, err = l.Search(searchRequest)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	//// log.Printf(fmt.Sprintf("303: %v",sr.Entries))
	for _, ent := range sr.Entries {
		// log.Printf(fmt.Sprintf("305: %v",ent.DN))
		groups = append(groups, EntryName{
			DN:   ent.DN,
			Name: ent.GetAttributeValue("cn"),
		})
		// log.Printf(fmt.Sprintf("310: %v",config.GroupCanInvite))
		if config.GroupCanInvite != "" && strings.EqualFold(ent.DN, config.GroupCanInvite) {
			loginStatus.CanInvite = true
		}
		// log.Printf(fmt.Sprintf("314: %v",config.GroupCanAdmin))
		if config.GroupCanAdmin != "" && strings.EqualFold(ent.DN, config.GroupCanAdmin) {
			loginStatus.CanAdmin = true
		}
	}

	// for _, attr := range loginStatus.UserEntry.Attributes {
	// 	if strings.EqualFold(attr.Name, "memberof") {
	// 		for _, group := range attr.Values {
	// 			if config.GroupCanInvite != "" && strings.EqualFold(group, config.GroupCanInvite) {
	// 				loginStatus.CanInvite = true
	// 			}
	// 			if config.GroupCanAdmin != "" && strings.EqualFold(group, config.GroupCanAdmin) {
	// 				loginStatus.CanAdmin = true
	// 			}
	// 		}
	// 	}
	// }

	return loginStatus
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, SESSION_NAME)
	if err != nil {
		session, _ = store.New(r, SESSION_NAME)
	}

	delete(session.Values, "login_username")
	delete(session.Values, "login_password")
	delete(session.Values, "login_dn")

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

type LoginFormData struct {
	Username     string
	WrongUser    bool
	WrongPass    bool
	ErrorMessage string
}

func handleLogin(w http.ResponseWriter, r *http.Request) *LoginInfo {
	templateLogin := getTemplate("login.html")

	if r.Method == "GET" {
		templateLogin.Execute(w, LoginFormData{})
		return nil
	} else if r.Method == "POST" {
		r.ParseForm()

		username := strings.Join(r.Form["username"], "")
		password := strings.Join(r.Form["password"], "")
		user_dn := fmt.Sprintf("%s=%s,%s", config.UserNameAttr, username, config.UserBaseDN)

		if strings.EqualFold(username, config.AdminAccount) {
			user_dn = username
		}
		loginInfo := doLogin(w, r, username, user_dn, password)
		return &loginInfo

	} else {
		http.Error(w, "Unsupported method", http.StatusBadRequest)
		return nil
	}
}

func doLogin(w http.ResponseWriter, r *http.Request, username string, user_dn string, password string) LoginInfo {
	l := ldapOpen(w)
	if l == nil {
		return nil
	}

	err := l.Bind(user_dn, password)
	if err == nil {
		templateLogin := getTemplate("login.html")
		data := &LoginFormData{
			Username: username,
		}
		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			data.WrongPass = true
		} else if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
			data.WrongUser = true
		} else {
			data.ErrorMessage = err.Error()
		}
		templateLogin.Execute(w, data)
		return nil
	}

	// Successfully logged in, save it to session
	session, err := store.Get(r, SESSION_NAME)
	if err != nil {
		session, _ = store.New(r, SESSION_NAME)
	}

	session.Values["login_username"] = username
	session.Values["login_password"] = password
	session.Values["login_dn"] = user_dn

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	return LoginInfo{
		DN:       user_dn,
		Username: username,
		Password: password,
	}

}
