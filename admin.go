package main

import (
	"strings"
	"fmt"
	"html/template"
	"net/http"
	"sort"

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
	Login        *LoginStatus
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
		Login:        login,
		GroupNameAttr: config.GroupNameAttr,
		Groups:        EntryList(sr.Entries),
	}
	sort.Sort(data.Groups)

	templateAdminGroups.Execute(w, data)
}

type AdminLDAPTplData struct {
	DN string
	Members []string
	Groups []string
	Props map[string]*PropValues
	Children []Child
	Path []PathItem
	AddError string
}

type Child struct {
	DN string
	Identifier string
	DisplayName string
}

type PathItem struct {
	DN string
	Identifier string
	Active bool
}

type PropValues struct {
	Values []string
	Editable bool
	ModifySuccess bool
	ModifyError string
}

func handleAdminLDAP(w http.ResponseWriter, r *http.Request) {
	templateAdminLDAP := template.Must(template.ParseFiles("templates/layout.html", "templates/admin_ldap.html"))

	login := checkAdminLogin(w, r)
	if login == nil {
		return
	}

	dn := mux.Vars(r)["dn"]

	modifyAttr := ""
	modifyError := ""
	modifySuccess := false
	addError := ""

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

			modifyAttr = attr
			if len(values_filtered) == 0 {
				modifyError = "Refusing to delete attribute."
			} else {
				modify_request := ldap.NewModifyRequest(dn, nil)
				modify_request.Replace(attr, values_filtered)

				err := login.conn.Modify(modify_request)
				if err != nil {
					modifyError = err.Error()
				} else {
					modifySuccess = true
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
			modifyAttr = attr
			if err != nil {
				addError = err.Error()
			}
		} else if action == "delete" {
			attr := strings.Join(r.Form["attr"], "")

			modify_request := ldap.NewModifyRequest(dn, nil)
			modify_request.Replace(attr, []string{})

			err := login.conn.Modify(modify_request)
			if err != nil {
				modifyError = err.Error()
			}
		}
	}

	// Build path
	path := []PathItem{
		PathItem{
			DN: config.BaseDN,
			Identifier: config.BaseDN,
			Active: dn == config.BaseDN,
		},
	}

	len_base_dn := len(strings.Split(config.BaseDN, ","))
	dn_split := strings.Split(dn, ",")
	dn_last_attr := strings.Split(dn_split[0], "=")[0]
	for i := len_base_dn + 1; i <= len(dn_split); i++ {
		path = append(path, PathItem{
			DN: strings.Join(dn_split[len(dn_split)-i:len(dn_split)], ","),
			Identifier: dn_split[len(dn_split)-i],
			Active: i == len(dn_split),
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
		if attr.Name != dn_last_attr {
			if existing, ok := props[attr.Name]; ok {
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
				pv := &PropValues{
					Values: attr.Values,
					Editable: editable,
				}
				if attr.Name == modifyAttr {
					if modifySuccess {
						pv.ModifySuccess = true
					} else if modifyError != "" {
						pv.ModifyError = modifyError
					}
				}
				props[attr.Name] = pv
			}
		}
	}

	members := []string{}
	if mp, ok := props["member"]; ok {
		members = mp.Values
		delete(props, "member")
	}
	groups := []string{}
	if gp, ok := props["memberof"]; ok {
		groups = gp.Values
		delete(props, "memberof")
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
			DN: item.DN,
			Identifier: strings.Split(item.DN, ",")[0],
			DisplayName: item.GetAttributeValue("displayname"),
		})
	}

	templateAdminLDAP.Execute(w, &AdminLDAPTplData{
		DN: dn,
		Members: members,
		Groups: groups,
		Props: props,
		Children: children,
		Path: path,
		AddError: addError,
	})
}
