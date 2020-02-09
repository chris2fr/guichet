package main

import (
	"html/template"
	"net/http"
	"fmt"
	"sort"

	"github.com/go-ldap/ldap/v3"
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

type AdminUsersTplData struct {
	Login *LoginStatus
	UserNameAttr string
	Users []*ldap.Entry
}

func handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	templateAdminUsers := template.Must(template.ParseFiles("templates/layout.html", "templates/admin_users.html"))

	login := checkLogin(w, r)
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
		Login: login,
		UserNameAttr: config.UserNameAttr,
		Users: sr.Entries,
	}
	sort.Sort(data)

	templateAdminUsers.Execute(w, data)
}

func (d *AdminUsersTplData) Len() int {
	return len(d.Users)
}

func (d *AdminUsersTplData) Swap(i, j int) {
	d.Users[i], d.Users[j] = d.Users[j], d.Users[i]
}

func (d *AdminUsersTplData) Less(i, j int) bool {
	return d.Users[i].GetAttributeValue(config.UserNameAttr) <
		d.Users[j].GetAttributeValue(config.UserNameAttr)
}
