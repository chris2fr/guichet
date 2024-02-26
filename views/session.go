/*
Handles session login and lougout with HTTP stuff
*/
package views

import (
	"fmt"
	"guichet/models"
	"log"
	"net/http"

	"bytes"
	"encoding/json"
	"io"
)

func PocketLogin (w http.ResponseWriter, r *http.Request) {


	type PocketRecord struct {
		avatar string
		email string
		username string
		name string
		collectionId string
		collectionName string
		emailVisibility bool 
		created string
		id string 
		updated string 
		verified bool
	}

	type PocketUser struct {
		record PocketRecord
		token string 
		code float64 
		message string 
		data map[string]interface{}
	}

	var postBody []byte
	var err error
	var responseBody *bytes.Buffer
	var resp *http.Response
	var body []byte
	// var pocketUserData PocketUser
	var jsonData map[string]interface{}
	// var jsonRecordData map[string]interface{}

	// log.Printf(apis.ContextAuthRecordKey)

	postBody, _ = json.Marshal(map[string]string{
		"identity": r.PostFormValue("identity"),
		"password": r.PostFormValue("password"),
	})

	// log.Printf(string(postBody))

	responseBody = bytes.NewBuffer(postBody)
	resp, err = http.Post(config.PocketbaseServer + "/api/collections/users/auth-with-password", "application/json", responseBody)
	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	// jsonData := PocketUser{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		panic(err)	
	}

	// fmt.Println(jsonData)

	session, err := GuichetSessionStore.Get(r, SESSION_NAME)
	if err != nil {
		session, _ = GuichetSessionStore.New(r, SESSION_NAME)
		log.Println("Supposed to be a new session")
		// return err
	}
	token, loggedin := jsonData["token"]
	if loggedin {
		session.Values["pocketbase_token"] = token
		session.Values["loggedin"] = loggedin
		// session.Values["loggedin"] = loggedin && jsonData["verified"].bool
		session.Values["can_admin"] = loggedin && r.PostFormValue("identity") == "chris@lesgrandsvoisins.com"
		// _ = json.Unmarshal([]byte(jsonData["record"]), &jsonRecordData)
		// session.Values["record"] = jsonData["record"].(map[string]interface{})
		jsonRecord := jsonData["record"].(map[string]interface{})
		session.Values["email"] = jsonRecord["email"].(string)
		session.Values["emailVisibility"] = jsonRecord["emailVisibility"].(bool)
		session.Values["username"] = jsonRecord["username"].(string)
		session.Values["avatar"] = jsonRecord["avatar"].(string)
		session.Values["name"] = jsonRecord["name"].(string)
		session.Values["created"] = jsonRecord["created"].(string)
		session.Values["updated"] = jsonRecord["updated"].(string)
		session.Values["verified"] = jsonRecord["verified"].(bool)
		// fmt.Println(session.Values["loggedin"])
		err = session.Save(r, w)
		if err != nil {
			log.Println("here")
			fmt.Println(err)
		}
	} else {
		session.Values["loggedin"] = false
		delete(session.Values, "pocketbase_token")
		delete(session.Values, "loggedin")
		delete(session.Values, "can_admin")
		delete(session.Values, "info")
		_ = session.Save(r, w)
	}

	// sb := string(body)
	// log.Printf(sb)
	// log.Printf(apis.ContextAuthRecordKey)
	http.Redirect(w, r, "/", http.StatusFound)

}

func PocketLoginCheck (w http.ResponseWriter, r *http.Request) (bool, bool, *LoginInfo) {
	
	session, err := GuichetSessionStore.Get(r, SESSION_NAME)
	if err != nil {
		// session, _ = GuichetSessionStore.New(r, SESSION_NAME)
		// return err
		fmt.Println("Not Supposed to be a new session")
		log.Printf("checkLogin ldapOpen : %v", err)
		log.Printf("checkLogin ldapOpen : %v", session)
		log.Printf("checkLogin ldapOpen : %v", session.Values)
	}
	
	// fmt.Println(session.Values["loggedin"])

	if session.Values["loggedin"] != nil {
			
		username, _ := session.Values["username"]
		email, _ := session.Values["email"]
		name, _ := session.Values["name"]
		can_admin, _ := session.Values["can_admin"]
		info := LoginInfo {
			Username: username.(string),
			Email: email.(string),
			Name: name.(string),
			CanAdmin: can_admin.(bool),
		}
		// info.Username = session.Values["record"].(map[string]string)["username"]
		// info.Email = session.Values["record"].(map[string]string)["email"]
		// info.Avatar = session.Values["record"].(map[string]string)["avatar"]
		// info.DN = session.Values["record"].(map[string]string)["dn"]
		// fmt.Println(info)
		return true, session.Values["can_admin"].(bool), &info
	} else {
		return false, false, nil
	}
}

func PocketLogout (w http.ResponseWriter, r *http.Request) {
	session, err := GuichetSessionStore.Get(r, SESSION_NAME)
	if err != nil {
		session, _ = GuichetSessionStore.New(r, SESSION_NAME)
		// return err
	} else {
		session.Values["loggedin"] = false
		delete(session.Values, "pocketbase_token")
		delete(session.Values, "loggedin")
		delete(session.Values, "can_admin")
		delete(session.Values, "info")
		_ = session.Save(r, w)
	}
	// http.Redirect(w,r,"/",http.StatusTemporaryRedirect)
	// http.Redirect(w, r, "/", http.StatusFound)
	loggedin, canadmin, info := PocketLoginCheck(w,r)
	if loggedin || canadmin || info != nil {
		log.Println("not logged out when should be logged out")
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func checkLogin(w http.ResponseWriter, r *http.Request) *LoginStatus {
	var login_info *LoginInfo
	l, err := ldapOpen(w)
	if l == nil {
		return nil
	}
	session, err := GuichetSessionStore.Get(r, SESSION_NAME)
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
		err = models.Bind(models.User{
			DN:       login_info.DN,
			Password: login_info.Password,
		}, &config, l)
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
		ldapUser, err := models.GetUser(models.User{
			DN: login_info.DN,
			CN: login_info.Username,
		}, &config, l)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil
		}
		userEntry := ldapUser.UserEntry
		loginStatus :=LoginStatus{
			Info:      login_info,
			conn:      l,
			UserEntry: userEntry,
			Common: NestedCommonTplData{
				CanAdmin:  ldapUser.CanAdmin,
				CanInvite: ldapUser.CanInvite,
			},
		}
		return &loginStatus
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
	session, err := GuichetSessionStore.Get(r, SESSION_NAME)
	if err != nil {
		session, _ = GuichetSessionStore.New(r, SESSION_NAME)
		// return err
	} else {
		delete(session.Values, "login_username")
		delete(session.Values, "login_password")
		delete(session.Values, "login_dn")

		err = session.Save(r, w)
	}

	// return err
	return nil
}
