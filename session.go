/*
Handles session login and lougout with HTTP stuff
*/
package main

import (
	"log"
	"net/http"
)

func checkLogin(w http.ResponseWriter, r *http.Request) *LoginStatus {
	var login_info *LoginInfo
	l, err := ldapOpen(w)
	if l == nil {
		return nil
	}
	session, err := store.Get(r, SESSION_NAME)
	if err != nil {
		log.Printf("checkLogin ldapOpen : %v", err)
		log.Printf("checkLogin ldapOpen : %v", session)
		log.Printf("checkLogin ldapOpen : %v", session.Values)
		return nil
	}
	username, ok := session.Values["login_username"]
	password, ok2 := session.Values["login_password"]
	user_dn, ok3 := session.Values["login_dn"]

	if ok && ok2 && ok3 {
		login_info = &LoginInfo{
			DN:       user_dn.(string),
			Username: username.(string),
			Password: password.(string),
		}
		err = bind(User{
			DN:       login_info.DN,
			Password: login_info.Password,
		}, config, l)
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
		ldapUser, err := get(User{
			DN: login_info.DN,
			CN: login_info.Username,
		}, config, l)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil
		}
		userEntry := ldapUser.UserEntry
		loginStatus := &LoginStatus{
			Info:      login_info,
			conn:      l,
			UserEntry: userEntry,
			CanAdmin:  ldapUser.CanAdmin,
			CanInvite: ldapUser.CanInvite,
		}
		return loginStatus
	} else {
		return nil
	}
}

/*

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
*/

func logout(w http.ResponseWriter, r *http.Request) error {
	session, err := store.Get(r, SESSION_NAME)
	if err != nil {
		session, _ = store.New(r, SESSION_NAME)
		return err
	}

	delete(session.Values, "login_username")
	delete(session.Values, "login_password")
	delete(session.Values, "login_dn")

	err = session.Save(r, w)
	return err
}
