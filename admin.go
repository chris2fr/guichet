package main

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/gorilla/mux"
)

func checkAdminLogin(w http.ResponseWriter, r *http.Request) *LoginStatus {
	login := checkLogin(w, r)
	if login == nil {
		return nil
	}

	if !login.CanAdmin {
		http.Error(w, "Not authorized to perform administrative operations.", http.StatusUnauthorized)
		return nil
	}
	return login
}

type EntryList []*ldap.Entry

func (d EntryList) Len() int {
	return len(d)
}

func (d EntryList) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d EntryList) Less(i, j int) bool {
	return d[i].DN < d[j].DN
}

type AdminUsersTplData struct {
	Login        *LoginStatus
	UserNameAttr string
	UserBaseDN   string
	Users        EntryList
}

func handleAdminActivateUsers(w http.ResponseWriter, r *http.Request) {
	templateAdminActivateUsers := getTemplate("admin_activate.html")
	login := checkAdminLogin(w, r)
	if login == nil {
		return
	}

	searchRequest := ldap.NewSearchRequest(
		config.InvitationBaseDN,
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=organizationalPerson))"),
		[]string{config.UserNameAttr, "displayName", "givenName", "sn", "mail", "uid", "cn"},
		nil)

	sr, err := login.conn.Search(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := &AdminUsersTplData{
		Login:        login,
		UserNameAttr: config.UserNameAttr,
		UserBaseDN:   config.UserBaseDN,
		Users:        EntryList(sr.Entries),
	}
	templateAdminActivateUsers.Execute(w, data)

}

func handleAdminActivateUser(w http.ResponseWriter, r *http.Request) {
	cn := mux.Vars(r)["cn"]
	login := checkAdminLogin(w, r)
	if login == nil {
		return
	}
	modifyRequest := *ldap.NewModifyDNRequest("cn="+cn+","+config.InvitationBaseDN, "cn="+cn, true, config.UserBaseDN)
	err := login.conn.ModifyDN(&modifyRequest)
	if err != nil {
		return
	}
	http.Redirect(w, r, "/admin/activate", http.StatusFound)
}

func handleAdminUnactivateUser(w http.ResponseWriter, r *http.Request) {
	cn := mux.Vars(r)["cn"]
	login := checkAdminLogin(w, r)
	if login == nil {
		return
	}
	modifyRequest := *ldap.NewModifyDNRequest("cn="+cn+","+config.UserBaseDN, "cn="+cn, true, config.InvitationBaseDN)
	err := login.conn.ModifyDN(&modifyRequest)
	if err != nil {
		return
	}
	http.Redirect(w, r, "/admin/users", http.StatusFound)
}

func handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	templateAdminUsers := getTemplate("admin_users.html")

	login := checkAdminLogin(w, r)
	if login == nil {
		return
	}

	searchRequest := ldap.NewSearchRequest(
		config.UserBaseDN,
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=organizationalPerson))"),
		[]string{config.UserNameAttr, "dn", "displayName", "givenName", "sn", "mail", "uid", "cn"},
		nil)

	sr, err := login.conn.Search(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := &AdminUsersTplData{
		Login:        login,
		UserNameAttr: config.UserNameAttr,
		UserBaseDN:   config.UserBaseDN,
		Users:        EntryList(sr.Entries),
	}
	sort.Sort(data.Users)

	// addNewUser(NewUser{
	// 	DN:          "cn=newuser@lesgv.com,ou=newusers,dc=resdigita,dc=org",
	// 	CN:          "newuser@lesgv.com",
	// 	GivenName:   "New",
	// 	SN:          "User",
	// 	DisplayName: "New User",
	// 	Mail:        "newuser@lesgv.com",
	// }, config, login)

	templateAdminUsers.Execute(w, data)
}

type AdminGroupsTplData struct {
	Login         *LoginStatus
	GroupNameAttr string
	GroupBaseDN   string
	Groups        EntryList
}

func handleAdminGroups(w http.ResponseWriter, r *http.Request) {
	templateAdminGroups := getTemplate("admin_groups.html")

	login := checkAdminLogin(w, r)
	if login == nil {
		return
	}

	searchRequest := ldap.NewSearchRequest(
		config.GroupBaseDN,
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=groupOfNames))"),
		[]string{config.GroupNameAttr, "dn", "description"},
		nil)

	sr, err := login.conn.Search(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := &AdminGroupsTplData{
		Login:         login,
		GroupNameAttr: config.GroupNameAttr,
		GroupBaseDN:   config.GroupBaseDN,
		Groups:        EntryList(sr.Entries),
	}
	sort.Sort(data.Groups)

	templateAdminGroups.Execute(w, data)
}

type AdminMailingTplData struct {
	Login           *LoginStatus
	MailingNameAttr string
	MailingBaseDN   string
	MailingLists    EntryList
}

func handleAdminMailing(w http.ResponseWriter, r *http.Request) {
	templateAdminMailing := getTemplate("admin_mailing.html")

	login := checkAdminLogin(w, r)
	if login == nil {
		return
	}

	searchRequest := ldap.NewSearchRequest(
		config.MailingBaseDN,
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=groupOfNames))"),
		[]string{config.MailingNameAttr, "dn", "description"},
		nil)

	sr, err := login.conn.Search(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := &AdminMailingTplData{
		Login:           login,
		MailingNameAttr: config.MailingNameAttr,
		MailingBaseDN:   config.MailingBaseDN,
		MailingLists:    EntryList(sr.Entries),
	}
	sort.Sort(data.MailingLists)

	templateAdminMailing.Execute(w, data)
}

type AdminMailingListTplData struct {
	Login           *LoginStatus
	MailingNameAttr string
	MailingBaseDN   string

	MailingList        *ldap.Entry
	Members            EntryList
	PossibleNewMembers EntryList
	AllowGuest         bool

	Error   string
	Success bool
}

func handleAdminMailingList(w http.ResponseWriter, r *http.Request) {
	templateAdminMailingList := getTemplate("admin_mailing_list.html")

	login := checkAdminLogin(w, r)
	if login == nil {
		return
	}

	id := mux.Vars(r)["id"]
	dn := fmt.Sprintf("%s=%s,%s", config.MailingNameAttr, id, config.MailingBaseDN)

	// handle modifications
	dError := ""
	dSuccess := false

	if r.Method == "POST" {
		r.ParseForm()
		action := strings.Join(r.Form["action"], "")
		if action == "add-member" {
			member := strings.Join(r.Form["member"], "")
			modify_request := ldap.NewModifyRequest(dn, nil)
			modify_request.Add("member", []string{member})

			err := login.conn.Modify(modify_request)
			// log.Printf(fmt.Sprintf("198: %v",modify_request))
			if err != nil {
				dError = err.Error()
			} else {
				dSuccess = true
			}
		} else if action == "add-external" {
			mail := strings.Join(r.Form["mail"], "")
			sn := strings.Join(r.Form["sn"], "")
			givenname := strings.Join(r.Form["givenname"], "")
			member := strings.Join(r.Form["member"], "")
			displayname := strings.Join(r.Form["displayname"], "")

			searchRequest := ldap.NewSearchRequest(
				config.UserBaseDN,
				ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
				fmt.Sprintf("(&(objectClass=organizationalPerson)(mail=%s))", mail),
				[]string{"dn", "displayname", "mail"},
				nil)
			sr, err := login.conn.Search(searchRequest)
			if err != nil {
				dError = err.Error()
			} else {
				if len(sr.Entries) == 0 {
					if config.MailingGuestsBaseDN != "" {
						guestDn := fmt.Sprintf("%s=%s,%s", config.UserNameAttr, mail, config.MailingGuestsBaseDN)
						req := ldap.NewAddRequest(guestDn, nil)
						//req.Attribute("objectclass", []string{"inetOrgPerson", "organizationalPerson", "person", "top"})
						req.Attribute("objectclass", []string{"inetOrgPerson"})
						req.Attribute("mail", []string{fmt.Sprintf("%s", mail)})
						if givenname != "" {
							req.Attribute("givenname", []string{givenname})
						}
						if member != "" {
							req.Attribute("member", []string{member})
						}
						if displayname != "" {
							req.Attribute("displayname", []string{displayname})
						}
						if sn != "" {
							req.Attribute("sn", []string{sn})
						}
						// log.Printf(fmt.Sprintf("226: %v",req))
						err := login.conn.Add(req)
						if err != nil {
							dError = err.Error()
						} else {
							modify_request := ldap.NewModifyRequest(dn, nil)
							modify_request.Add("member", []string{guestDn})

							err := login.conn.Modify(modify_request)
							// log.Printf(fmt.Sprintf("249: %v",modify_request))
							if err != nil {
								dError = err.Error()
							} else {
								dSuccess = true
							}
						}
					} else {
						dError = "Adding guest users not supported, the user must already have an LDAP account."
					}
				} else if len(sr.Entries) == 1 {
					modify_request := ldap.NewModifyRequest(dn, nil)
					modify_request.Add("member", []string{sr.Entries[0].DN})

					err := login.conn.Modify(modify_request)
					// log.Printf(fmt.Sprintf("264: %v",modify_request))
					if err != nil {
						dError = err.Error()
					} else {
						dSuccess = true
					}
				} else {
					dError = fmt.Sprintf("Multiple users exist with email address %s", mail)
				}
			}
		} else if action == "delete-member" {
			member := strings.Join(r.Form["member"], "")
			modify_request := ldap.NewModifyRequest(dn, nil)
			modify_request.Delete("member", []string{member})

			err := login.conn.Modify(modify_request)
			// log.Printf(fmt.Sprintf("280: %v",modify_request))
			if err != nil {
				dError = err.Error()
			} else {
				dSuccess = true
			}
		}
	}

	// Retrieve mailing list
	searchRequest := ldap.NewSearchRequest(
		dn,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(objectclass=groupOfNames)"),
		[]string{"dn", config.MailingNameAttr, "member", "description"},
		nil)

	sr, err := login.conn.Search(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(sr.Entries) != 1 {
		http.Error(w, fmt.Sprintf("Object not found: %s", dn), http.StatusNotFound)
		return
	}

	ml := sr.Entries[0]

	memberDns := make(map[string]bool)
	for _, attr := range ml.Attributes {
		if attr.Name == "member" {
			for _, v := range attr.Values {
				memberDns[v] = true
			}
		}
	}

	// Retrieve list of current and possible new members
	members := []*ldap.Entry{}
	possibleNewMembers := []*ldap.Entry{}

	searchRequest = ldap.NewSearchRequest(
		config.UserBaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(objectClass=organizationalPerson)"),
		[]string{"dn", "displayname", "mail"},
		nil)
	sr, err = login.conn.Search(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, ent := range sr.Entries {
		if _, ok := memberDns[ent.DN]; ok {
			members = append(members, ent)
		} else {
			possibleNewMembers = append(possibleNewMembers, ent)
		}
	}

	data := &AdminMailingListTplData{
		Login:           login,
		MailingNameAttr: config.MailingNameAttr,
		MailingBaseDN:   config.MailingBaseDN,

		MailingList:        ml,
		Members:            members,
		PossibleNewMembers: possibleNewMembers,
		AllowGuest:         config.MailingGuestsBaseDN != "",

		Error:   dError,
		Success: dSuccess,
	}
	sort.Sort(data.Members)
	sort.Sort(data.PossibleNewMembers)

	templateAdminMailingList.Execute(w, data)
}

// ===================================================
// 					   LDAP EXPLORER
// ===================================================

type AdminLDAPTplData struct {
	DN string

	Path          []PathItem
	ChildrenOU    []Child
	ChildrenOther []Child
	CanAddChild   bool
	Props         map[string]*PropValues
	CanDelete     bool

	HasMembers         bool
	Members            []EntryName
	PossibleNewMembers []EntryName
	HasGroups          bool
	Groups             []EntryName
	PossibleNewGroups  []EntryName

	ListMemGro map[string]string

	Error   string
	Success bool
}

type EntryName struct {
	DN   string
	Name string
}

type Child struct {
	DN         string
	Identifier string
	Name       string
}

type PathItem struct {
	DN         string
	Identifier string
	Active     bool
}

type PropValues struct {
	Name      string
	Values    []string
	Editable  bool
	Deletable bool
}

func handleAdminLDAP(w http.ResponseWriter, r *http.Request) {
	templateAdminLDAP := getTemplate("admin_ldap.html")

	login := checkAdminLogin(w, r)
	if login == nil {
		return
	}

	dn := mux.Vars(r)["dn"]

	dError := ""
	dSuccess := false

	// Build path
	path := []PathItem{
		PathItem{
			DN:         config.BaseDN,
			Identifier: config.BaseDN,
			Active:     dn == config.BaseDN,
		},
	}
	// log.Printf(fmt.Sprintf("434: %v",path))

	len_base_dn := len(strings.Split(config.BaseDN, ","))
	dn_split := strings.Split(dn, ",")
	dn_last_attr := strings.Split(dn_split[0], "=")[0]
	for i := len_base_dn + 1; i <= len(dn_split); i++ {
		path = append(path, PathItem{
			DN:         strings.Join(dn_split[len(dn_split)-i:len(dn_split)], ","),
			Identifier: dn_split[len(dn_split)-i],
			Active:     i == len(dn_split),
		})
	}
	// log.Printf(fmt.Sprintf("446: %v",path))

	// Handle modification operation
	if r.Method == "POST" {
		r.ParseForm()
		action := strings.Join(r.Form["action"], "")
		if action == "modify" {
			attr := strings.Join(r.Form["attr"], "")
			values := strings.Split(strings.Join(r.Form["values"], ""), "\n")
			values_filtered := []string{}
			for _, v := range values {
				v2 := strings.TrimSpace(v)
				if v2 != "" {
					values_filtered = append(values_filtered, v2)
				}
			}

			if len(values_filtered) == 0 {
				dError = "Refusing to delete attribute."
			} else {
				modify_request := ldap.NewModifyRequest(dn, nil)
				modify_request.Replace(attr, values_filtered)

				err := login.conn.Modify(modify_request)
				// log.Printf(fmt.Sprintf("468: %v",modify_request))
				if err != nil {
					dError = err.Error()
				} else {
					dSuccess = true
				}
			}
		} else if action == "add" {
			attr := strings.Join(r.Form["attr"], "")
			values := strings.Split(strings.Join(r.Form["values"], ""), "\n")
			values_filtered := []string{}
			for _, v := range values {
				v2 := strings.TrimSpace(v)
				if v2 != "" {
					values_filtered = append(values_filtered, v2)
				}
			}

			modify_request := ldap.NewModifyRequest(dn, nil)
			modify_request.Add(attr, values_filtered)

			err := login.conn.Modify(modify_request)
			// log.Printf(fmt.Sprintf("490: %v",modify_request))
			if err != nil {
				dError = err.Error()
			} else {
				dSuccess = true
			}
		} else if action == "delete" {
			attr := strings.Join(r.Form["attr"], "")

			modify_request := ldap.NewModifyRequest(dn, nil)
			modify_request.Replace(attr, []string{})

			err := login.conn.Modify(modify_request)
			// log.Printf(fmt.Sprintf("503: %v",modify_request))
			if err != nil {
				dError = err.Error()
			} else {
				dSuccess = true
			}
		} else if action == "delete-from-group" {
			group := strings.Join(r.Form["group"], "")
			modify_request := ldap.NewModifyRequest(group, nil)
			modify_request.Delete("member", []string{dn})

			err := login.conn.Modify(modify_request)
			// log.Printf(fmt.Sprintf("515: %v",modify_request))
			if err != nil {
				dError = err.Error()
			} else {
				dSuccess = true
			}
		} else if action == "add-to-group" {
			group := strings.Join(r.Form["group"], "")
			modify_request := ldap.NewModifyRequest(group, nil)
			modify_request.Add("member", []string{dn})

			err := login.conn.Modify(modify_request)
			// log.Printf(fmt.Sprintf("527: %v",modify_request))
			if err != nil {
				dError = err.Error()
			} else {
				dSuccess = true
			}
		} else if action == "delete-member" {
			member := strings.Join(r.Form["member"], "")
			modify_request := ldap.NewModifyRequest(dn, nil)
			modify_request.Delete("member", []string{member})

			err := login.conn.Modify(modify_request)
			// log.Printf(fmt.Sprintf("539: %v",modify_request))
			if err != nil {
				dError = err.Error()
			} else {
				dSuccess = true
			}
		} else if action == "delete-object" {
			del_request := ldap.NewDelRequest(dn, nil)
			err := login.conn.Del(del_request)
			if err != nil {
				dError = err.Error()
			} else {
				http.Redirect(w, r, "/admin/ldap/"+strings.Join(dn_split[1:], ","), http.StatusFound)
				return
			}
		}
	}

	// Get object and parse it
	searchRequest := ldap.NewSearchRequest(
		dn,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(objectclass=*)"),
		[]string{},
		nil)

	sr, err := login.conn.Search(searchRequest)
	// log.Printf(fmt.Sprintf("569: %v",searchRequest))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(sr.Entries) != 1 {
		http.Error(w, fmt.Sprintf("Object not found: %s", dn), http.StatusNotFound)
		return
	}

	object := sr.Entries[0]

	// Read object properties and prepare appropriate form fields
	props := make(map[string]*PropValues)
	for _, attr := range object.Attributes {
		name_lower := strings.ToLower(attr.Name)
		if name_lower != dn_last_attr {
			if existing, ok := props[name_lower]; ok {
				existing.Values = append(existing.Values, attr.Values...)
			} else {
				editable := true
				for _, restricted := range []string{
					"creatorsname", "modifiersname", "createtimestamp",
					"modifytimestamp", "entryuuid",
				} {
					if strings.EqualFold(attr.Name, restricted) {
						editable = false
						break
					}
				}
				deletable := true
				for _, restricted := range []string{"objectclass", "structuralobjectclass"} {
					if strings.EqualFold(attr.Name, restricted) {
						deletable = false
						break
					}
				}
				props[name_lower] = &PropValues{
					Name:      attr.Name,
					Values:    attr.Values,
					Editable:  editable,
					Deletable: deletable,
				}
			}
		}
	}

	// Check objectclass to determine object type
	objectClass := []string{}
	if val, ok := props["objectclass"]; ok {
		objectClass = val.Values
	}
	hasMembers, hasGroups, isOrganization := false, false, false
	for _, oc := range objectClass {
		if strings.EqualFold(oc, "organizationalPerson") || strings.EqualFold(oc, "person") || strings.EqualFold(oc, "inetOrgPerson") {
			hasGroups = true
		}
		if strings.EqualFold(oc, "groupOfNames") {
			hasMembers = true
		}
		if strings.EqualFold(oc, "organization") {
			isOrganization = true
		}
	}

	// Parse member list and prepare form section
	members_dn := []string{}
	if mp, ok := props["member"]; ok {
		members_dn = mp.Values
		delete(props, "member")
	}

	members := []EntryName{}
	possibleNewMembers := []EntryName{}
	if len(members_dn) > 0 || hasMembers {
		// Lookup all existing users in the server
		// to know the DN -> display name correspondance
		searchRequest = ldap.NewSearchRequest(
			config.UserBaseDN,
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			fmt.Sprintf("(objectClass=organizationalPerson)"),
			[]string{"dn", "displayname", "description"},
			nil)
		sr, err = login.conn.Search(searchRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userMap := make(map[string]string)
		for _, ent := range sr.Entries {
			userMap[ent.DN] = ent.GetAttributeValue("displayname")
			if userMap[ent.DN] == "" {
				userMap[ent.DN] = ent.GetAttributeValue("description")
			}
		}

		// Select members with their name and remove them from map
		for _, memdn := range members_dn {
			members = append(members, EntryName{
				DN:   memdn,
				Name: userMap[memdn],
			})
			delete(userMap, memdn)
		}

		// Create list of members that can be added
		for dn, name := range userMap {
			entry := EntryName{
				DN:   dn,
				Name: name,
			}
			if entry.Name == "" {
				entry.Name = entry.DN
			}
			possibleNewMembers = append(possibleNewMembers, entry)
		}
	}

	// // Parse group list and prepare form section
	// groups_dn := []string{}
	// if gp, ok := props["memberof"]; ok {
	// 	groups_dn = gp.Values
	// 	delete(props, "memberof")
	// }

	groups := []EntryName{}
	possibleNewGroups := []EntryName{}
	searchRequest = ldap.NewSearchRequest(
		config.GroupBaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=groupOfNames)(member=%s))", dn),
		[]string{"dn", "displayName", "cn", "description"},
		nil)
	// log.Printf(fmt.Sprintf("708: %v",searchRequest))
	sr, err = login.conn.Search(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// log.Printf(fmt.Sprintf("714: %v",sr.Entries))
	for _, ent := range sr.Entries {
		groups = append(groups, EntryName{
			DN:   ent.DN,
			Name: ent.GetAttributeValue("cn"),
		})
	}
	searchRequest = ldap.NewSearchRequest(
		config.GroupBaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=groupOfNames)(!(member=%s)))", dn),
		[]string{"dn", "displayName", "cn", "description"},
		nil)
	// log.Printf(fmt.Sprintf("724: %v",searchRequest))
	sr, err = login.conn.Search(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// log.Printf(fmt.Sprintf("714: %v",sr.Entries))
	for _, ent := range sr.Entries {
		possibleNewGroups = append(possibleNewGroups, EntryName{
			DN:   ent.DN,
			Name: ent.GetAttributeValue("cn"),
		})
	}

	// possibleNewGroup.DN = ent.GetAttributeValue("dn")
	// possibleNewGroup.Name = ent.GetAttributeValue("cn")
	// // log.Printf(fmt.Sprintf("725: %v %v",dn, ent.GetAttributeValue("member")))
	// for _, member := range ent   .GetAttributeValue("member") {
	// // 	// log.Printf(fmt.Sprintf("725: %v %v",dn, member))
	// if ent.GetAttributeValue("member") == dn {
	// 	groups = append(groups,possibleNewGroup,)
	// 	possibleNewGroup.DN = ""
	// 	possibleNewGroup.Name = ""
	// }
	// // }
	// if possibleNewGroup.DN != "" {
	//   possibleNewGroups = append(possibleNewGroups,possibleNewGroup,)
	// 	possibleNewGroup = EntryName{}
	// }

	// groupMap[.DN] = ent.GetAttributeValue("displayName")
	// if groupMap[.DN] == "" {
	// 	groupMap[.DN] = ent.GetAttributeValue("cn")
	// }
	// if groupMap[.DN] == "" {
	// 	groupMap[.DN] = ent.GetAttributeValue("description")
	// }
	// }

	// // Calculate list of current groups
	// // log.Printf(fmt.Sprintf("%v",groups_dn))
	// for _, grpdn := range groups_dn {
	// 	// log.Printf(fmt.Sprintf("%v",grpdn))
	// 	groups = append(groups, EntryName{
	// 		DN:   grpdn,
	// 		Name: groupMap[grpdn],
	// 	})
	// 	delete(groupMap, grpdn)
	// }

	// // Calculate list of possible new groups
	// for dn, name := range groupMap {
	// 	entry := EntryName{
	// 		DN:   dn,
	// 		Name: name,
	// 	}
	// 	if entry.Name == "" {
	// 		entry.Name = entry.DN
	// 	}
	// 	possibleNewGroups = append(possibleNewGroups, entry)
	// }
	// }

	// Get children
	searchRequest = ldap.NewSearchRequest(
		dn,
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(objectclass=*)"),
		[]string{"dn", "displayname", "description"},
		nil)

	sr, err = login.conn.Search(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sort.Sort(EntryList(sr.Entries))

	childrenOU := []Child{}
	childrenOther := []Child{}
	for _, item := range sr.Entries {
		name := item.GetAttributeValue("displayname")
		if name == "" {
			name = item.GetAttributeValue("description")
		}
		child := Child{
			DN:         item.DN,
			Identifier: strings.Split(item.DN, ",")[0],
			Name:       name,
		}
		if strings.HasPrefix(item.DN, "ou=") {
			childrenOU = append(childrenOU, child)
		} else {
			childrenOther = append(childrenOther, child)
		}
	}

	// Run template, finally!
	templateAdminLDAP.Execute(w, &AdminLDAPTplData{
		DN: dn,

		Path:          path,
		ChildrenOU:    childrenOU,
		ChildrenOther: childrenOther,
		Props:         props,
		CanAddChild:   dn_last_attr == "ou" || isOrganization,
		CanDelete:     dn != config.BaseDN && len(childrenOU) == 0 && len(childrenOther) == 0,

		HasMembers:         len(members) > 0 || hasMembers,
		Members:            members,
		PossibleNewMembers: possibleNewMembers,
		HasGroups:          len(groups) > 0 || hasGroups,
		Groups:             groups,
		PossibleNewGroups:  possibleNewGroups,

		Error:   dError,
		Success: dSuccess,
	})
}

type CreateData struct {
	SuperDN  string
	Path     []PathItem
	Template string

	IdType                string
	IdValue               string
	DisplayName           string
	GivenName             string
	Member                string
	Mail                  string
	Description           string
	StructuralObjectClass string
	ObjectClass           string
	SN                    string

	Error string
}

func handleAdminCreate(w http.ResponseWriter, r *http.Request) {
	templateAdminCreate := getTemplate("admin_create.html")

	login := checkAdminLogin(w, r)
	if login == nil {
		return
	}

	template := mux.Vars(r)["template"]
	super_dn := mux.Vars(r)["super_dn"]

	// Check that base DN exists
	searchRequest := ldap.NewSearchRequest(
		super_dn,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(objectclass=*)"),
		[]string{},
		nil)

	sr, err := login.conn.Search(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(sr.Entries) != 1 {
		http.Error(w, fmt.Sprintf("Parent object %s does not exist", super_dn), http.StatusNotFound)
		return
	}

	// Build path
	path := []PathItem{
		PathItem{
			DN:         config.BaseDN,
			Identifier: config.BaseDN,
		},
	}

	len_base_dn := len(strings.Split(config.BaseDN, ","))
	dn_split := strings.Split(super_dn, ",")
	for i := len_base_dn + 1; i <= len(dn_split); i++ {
		path = append(path, PathItem{
			DN:         strings.Join(dn_split[len(dn_split)-i:len(dn_split)], ","),
			Identifier: dn_split[len(dn_split)-i],
		})
	}

	// Handle data
	data := &CreateData{
		SuperDN: super_dn,
		Path:    path,
	}
	data.Template = template
	if template == "user" {
		data.IdType = config.UserNameAttr
		data.StructuralObjectClass = "inetOrgPerson"
		data.ObjectClass = "inetOrgPerson\norganizationalPerson\nperson\ntop"
	} else if template == "group" || template == "ml" {
		data.IdType = config.UserNameAttr
		data.StructuralObjectClass = "groupOfNames"
		data.ObjectClass = "groupOfNames\ntop"
		data.Member = "cn=sogo@resdigita.org,ou=users,dc=resdigita,dc=org"
	} else if template == "ou" {
		data.IdType = "ou"
		data.StructuralObjectClass = "organizationalUnit"
		data.ObjectClass = "organizationalUnit\ntop"
	} else {
		data.IdType = "cn"
		data.ObjectClass = "top"
		data.Template = ""
	}

	if r.Method == "POST" {
		r.ParseForm()
		if data.Template == "" {
			data.IdType = strings.TrimSpace(strings.Join(r.Form["idtype"], ""))
			data.StructuralObjectClass = strings.TrimSpace(strings.Join(r.Form["soc"], ""))
			data.ObjectClass = strings.Join(r.Form["oc"], "")
		}
		data.IdValue = strings.TrimSpace(strings.Join(r.Form["idvalue"], ""))
		data.DisplayName = strings.TrimSpace(strings.Join(r.Form["displayname"], ""))
		data.GivenName = strings.TrimSpace(strings.Join(r.Form["givenname"], ""))
		data.Mail = strings.TrimSpace(strings.Join(r.Form["mail"], ""))
		data.Member = strings.TrimSpace(strings.Join(r.Form["member"], ""))
		data.Description = strings.TrimSpace(strings.Join(r.Form["description"], ""))
		data.SN = strings.TrimSpace(strings.Join(r.Form["sn"], ""))

		object_class := []string{}
		for _, oc := range strings.Split(data.ObjectClass, "\n") {
			x := strings.TrimSpace(oc)
			if x != "" {
				object_class = append(object_class, x)
			}
		}

		if len(object_class) == 0 {
			data.Error = "No object class specified"
		} else if match, err := regexp.MatchString("^[a-z]+$", data.IdType); err != nil || !match {
			data.Error = "Invalid identifier type"
		} else if len(data.IdValue) == 0 {
			data.Error = "No identifier specified"
		} else {
			newUser := NewUser{
				DN: data.IdType + "=" + data.IdValue + "," + super_dn,
			}
			// dn := data.IdType + "=" + data.IdValue + "," + super_dn
			// req := ldap.NewAddRequest(dn, nil)
			// req.Attribute("objectclass", object_class)
			// req.Attribute("mail", []string{data.IdValue})
			/*
				if data.StructuralObjectClass != "" {
					req.Attribute("structuralobjectclass", []string{data.StructuralObjectClass})
				}
			*/
			if data.Mail != "" {
				newUser.Mail = data.Mail
				// req.Attribute("mail", []string{data.Mail})
			}
			if data.IdType == "cn" {
				newUser.CN = data.IdValue
			} else if data.IdType == "mail" {
				newUser.Mail = data.IdValue
			} else if data.IdType == "uid" {
				newUser.UID = data.IdValue
			}

			if data.DisplayName != "" {
				newUser.DisplayName = data.DisplayName
				// req.Attribute("displayname", []string{data.DisplayName})
			}
			if data.GivenName != "" {
				newUser.GivenName = data.GivenName
				// req.Attribute("givenname", []string{data.GivenName})
			}

			// if data.Member != "" {
			// 	req.Attribute("member", []string{data.Member})
			// }
			if data.SN != "" {
				newUser.SN = data.SN
				// req.Attribute("sn", []string{data.SN})
			}
			if data.Description != "" {
				newUser.Description = data.Description
				// req.Attribute("description", []string{data.Description})
			}

			addNewUser(newUser, config)

			// err := login.conn.Add(req)
			// // log.Printf(fmt.Sprintf("899: %v",err))
			// // log.Printf(fmt.Sprintf("899: %v",req))
			// // log.Printf(fmt.Sprintf("899: %v",data))
			// if err != nil {
			// 	data.Error = err.Error()
			// } else {
			if template == "ml" {
				http.Redirect(w, r, "/admin/mailing/"+data.IdValue, http.StatusFound)
			} else {
				http.Redirect(w, r, "/admin/ldap/"+newUser.DN, http.StatusFound)
			}
			// }
		}
	}

	templateAdminCreate.Execute(w, data)
}
