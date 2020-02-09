package main

import (
	"fmt"
	"html/template"
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

	can_admin := false
	for _, group := range login.UserEntry.GetAttributeValues("memberof") {
		if config.GroupCanAdmin != "" && group == config.GroupCanAdmin {
			can_admin = true
		}
	}

	if !can_admin {
		http.Redirect(w, r, "/", http.StatusFound)
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
	Users        EntryList
}

func handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	templateAdminUsers := template.Must(template.ParseFiles("templates/layout.html", "templates/admin_users.html"))

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
		Users:        EntryList(sr.Entries),
	}
	sort.Sort(data.Users)

	templateAdminUsers.Execute(w, data)
}

type AdminGroupsTplData struct {
	Login         *LoginStatus
	GroupNameAttr string
	Groups        EntryList
}

func handleAdminGroups(w http.ResponseWriter, r *http.Request) {
	templateAdminGroups := template.Must(template.ParseFiles("templates/layout.html", "templates/admin_groups.html"))

	login := checkAdminLogin(w, r)
	if login == nil {
		return
	}

	searchRequest := ldap.NewSearchRequest(
		config.GroupBaseDN,
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=groupOfNames))"),
		[]string{config.GroupNameAttr, "dn", "displayname"},
		nil)

	sr, err := login.conn.Search(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := &AdminGroupsTplData{
		Login:         login,
		GroupNameAttr: config.GroupNameAttr,
		Groups:        EntryList(sr.Entries),
	}
	sort.Sort(data.Groups)

	templateAdminGroups.Execute(w, data)
}

type AdminLDAPTplData struct {
	DN string

	Path     []PathItem
	Children []Child
	CanAddChild bool
	Props    map[string]*PropValues

	HasMembers bool
	Members    []EntryName
	HasGroups  bool
	Groups     []EntryName

	Error   string
	Success bool
}

type EntryName struct {
	DN          string
	DisplayName string
}

type Child struct {
	DN          string
	Identifier  string
	DisplayName string
}

type PathItem struct {
	DN         string
	Identifier string
	Active     bool
}

type PropValues struct {
	Name     string
	Values   []string
	Editable bool
}

func handleAdminLDAP(w http.ResponseWriter, r *http.Request) {
	templateAdminLDAP := template.Must(template.ParseFiles("templates/layout.html", "templates/admin_ldap.html"))

	login := checkAdminLogin(w, r)
	if login == nil {
		return
	}

	dn := mux.Vars(r)["dn"]

	dError := ""
	dSuccess := false

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
		}
	}

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
		http.Error(w, fmt.Sprintf("%d objects found", len(sr.Entries)), http.StatusInternalServerError)
		return
	}

	object := sr.Entries[0]

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
				props[name_lower] = &PropValues{
					Name:     attr.Name,
					Values:   attr.Values,
					Editable: editable,
				}
			}
		}
	}

	members_dn := []string{}
	if mp, ok := props["member"]; ok {
		members_dn = mp.Values
		delete(props, "member")
	}

	members := []EntryName{}
	if len(members_dn) > 0 {
		mapDnToName := make(map[string]string)
		searchRequest = ldap.NewSearchRequest(
			config.UserBaseDN,
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			fmt.Sprintf("(objectClass=organizationalPerson)"),
			[]string{"dn", "displayname"},
			nil)
		sr, err := login.conn.Search(searchRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, ent := range sr.Entries {
			mapDnToName[ent.DN] = ent.GetAttributeValue("displayname")
		}
		for _, memdn := range members_dn {
			members = append(members, EntryName{
				DN:          memdn,
				DisplayName: mapDnToName[memdn],
			})
		}
	}

	groups_dn := []string{}
	if gp, ok := props["memberof"]; ok {
		groups_dn = gp.Values
		delete(props, "memberof")
	}

	groups := []EntryName{}
	if len(groups_dn) > 0 {
		mapDnToName := make(map[string]string)
		searchRequest = ldap.NewSearchRequest(
			config.GroupBaseDN,
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			fmt.Sprintf("(objectClass=groupOfNames)"),
			[]string{"dn", "displayname"},
			nil)
		sr, err := login.conn.Search(searchRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, ent := range sr.Entries {
			mapDnToName[ent.DN] = ent.GetAttributeValue("displayname")
		}
		for _, grpdn := range groups_dn {
			groups = append(groups, EntryName{
				DN:          grpdn,
				DisplayName: mapDnToName[grpdn],
			})
		}
	}

	// Get children
	searchRequest = ldap.NewSearchRequest(
		dn,
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(objectclass=*)"),
		[]string{"dn", "displayname"},
		nil)

	sr, err = login.conn.Search(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sort.Sort(EntryList(sr.Entries))

	children := []Child{}
	for _, item := range sr.Entries {
		children = append(children, Child{
			DN:          item.DN,
			Identifier:  strings.Split(item.DN, ",")[0],
			DisplayName: item.GetAttributeValue("displayname"),
		})
	}

	// Checkup objectclass
	objectClass := []string{}
	if val, ok := props["objectclass"]; ok {
		objectClass = val.Values
	}
	hasMembers, hasGroups := false, false
	for _, oc := range objectClass {
		if strings.EqualFold(oc, "organizationalperson") || strings.EqualFold(oc, "person") {
			hasGroups = true
		}
		if strings.EqualFold(oc, "groupOfNames") {
			hasMembers = true
		}
	}

	templateAdminLDAP.Execute(w, &AdminLDAPTplData{
		DN: dn,

		Path:     path,
		Children: children,
		Props:    props,
		CanAddChild: dn_last_attr == "ou",

		HasMembers: len(members) > 0 || hasMembers,
		Members:    members,
		HasGroups:  len(groups) > 0 || hasGroups,
		Groups:     groups,

		Error:   dError,
		Success: dSuccess,
	})
}

type CreateData struct {
	SuperDN string
	Path     []PathItem

	IdType                string
	IdValue               string
	DisplayName           string
	StructuralObjectClass string
	ObjectClass           string

	Error string
}

func handleAdminCreate(w http.ResponseWriter, r *http.Request) {
	templateAdminCreate := template.Must(template.ParseFiles("templates/layout.html", "templates/admin_create.html"))

	login := checkAdminLogin(w, r)
	if login == nil {
		return
	}

	template := mux.Vars(r)["template"]
	super_dn := mux.Vars(r)["super_dn"]

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
		Path: path,
	}
	if template == "user" {
		data.IdType = config.UserNameAttr
		data.StructuralObjectClass = "inetOrgPerson"
		data.ObjectClass = "inetOrgPerson\norganizationalPerson\nperson\ntop"
	} else if template == "group" {
		data.IdType = config.UserNameAttr
		data.StructuralObjectClass = "groupOfNames"
		data.ObjectClass = "groupOfNames\ntop"
	} else {
		data.IdType = "cn"
		data.ObjectClass = "top"
	}

	if r.Method == "POST" {
		r.ParseForm()
		data.IdType = strings.Join(r.Form["idtype"], "")
		data.IdValue = strings.Join(r.Form["idvalue"], "")
		data.DisplayName = strings.Join(r.Form["displayname"], "")
		data.StructuralObjectClass = strings.Join(r.Form["soc"], "")
		data.ObjectClass = strings.Join(r.Form["oc"], "")

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
		} else if match, err := regexp.MatchString("^[\\d\\w_-]+$", data.IdValue); err != nil || !match {
			data.Error = "Invalid identifier"
		} else {
			dn := data.IdType + "=" + data.IdValue + "," + super_dn
			req := ldap.NewAddRequest(dn, nil)
			req.Attribute("objectClass", object_class)
			req.Attribute("structuralObjectClass",
				[]string{data.StructuralObjectClass})
			req.Attribute("displayname", []string{data.DisplayName})
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
