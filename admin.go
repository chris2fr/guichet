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
		[]string{config.UserNameAttr, "dn", "displayname", "givenname", "sn", "mail"},
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

type AdminLDAPTplData struct {
	DN string

	Path        []PathItem
	ChildrenOU  []Child
	ChildrenOther    []Child
	CanAddChild bool
	Props       map[string]*PropValues
	CanDelete   bool

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
		if strings.EqualFold(oc, "organizationalperson") || strings.EqualFold(oc, "person") {
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

	// Parse group list and prepare form section
	groups_dn := []string{}
	if gp, ok := props["memberof"]; ok {
		groups_dn = gp.Values
		delete(props, "memberof")
	}

	groups := []EntryName{}
	possibleNewGroups := []EntryName{}
	if len(groups_dn) > 0 || hasGroups {
		// Lookup all existing groups in the server
		// to know the DN -> display name correspondance
		searchRequest = ldap.NewSearchRequest(
			config.GroupBaseDN,
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			fmt.Sprintf("(objectClass=groupOfNames)"),
			[]string{"dn", "description"},
			nil)
		sr, err = login.conn.Search(searchRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		groupMap := make(map[string]string)
		for _, ent := range sr.Entries {
			groupMap[ent.DN] = ent.GetAttributeValue("displayname")
			if groupMap[ent.DN] == "" {
				groupMap[ent.DN] = ent.GetAttributeValue("description")
			}
		}

		// Calculate list of current groups
		for _, grpdn := range groups_dn {
			groups = append(groups, EntryName{
				DN:   grpdn,
				Name: groupMap[grpdn],
			})
			delete(groupMap, grpdn)
		}

		// Calculate list of possible new groups
		for dn, name := range groupMap {
			entry := EntryName{
				DN:   dn,
				Name: name,
			}
			if entry.Name == "" {
				entry.Name = entry.DN
			}
			possibleNewGroups = append(possibleNewGroups, entry)
		}
	}

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

		Path:        path,
		ChildrenOU:    childrenOU,
		ChildrenOther:    childrenOther,
		Props:       props,
		CanAddChild: dn_last_attr == "ou" || isOrganization,
		CanDelete:   dn != config.BaseDN && len(childrenOU) == 0 && len(childrenOther) == 0,

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
	Description           string
	StructuralObjectClass string
	ObjectClass           string

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
	} else if template == "group" {
		data.IdType = config.UserNameAttr
		data.StructuralObjectClass = "groupOfNames"
		data.ObjectClass = "groupOfNames\ntop"
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
		data.Description = strings.TrimSpace(strings.Join(r.Form["description"], ""))

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
			dn := data.IdType + "=" + data.IdValue + "," + super_dn
			req := ldap.NewAddRequest(dn, nil)
			req.Attribute("objectclass", object_class)
			if data.StructuralObjectClass != "" {
				req.Attribute("structuralobjectclass", []string{data.StructuralObjectClass})
			}
			if data.DisplayName != "" {
				req.Attribute("displayname", []string{data.DisplayName})
			}
			if data.Description != "" {
				req.Attribute("description", []string{data.Description})
			}

			err := login.conn.Add(req)
			if err != nil {
				data.Error = err.Error()
			} else {
				http.Redirect(w, r, "/admin/ldap/"+dn, http.StatusFound)
			}

		}
	}

	templateAdminCreate.Execute(w, data)
}
