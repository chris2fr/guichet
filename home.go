/*
home show the home page
*/

package main

import "net/http"

type HomePageData struct {
	Login    *LoginStatus
	BaseDN   string
	Org      string
	CanAdmin bool
	LoggedIn bool
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	templateHome := getTemplate("home.html")

	login := checkLogin(w, r)
	if login == nil {
		status := handleLogin(w, r)
		if status == nil {
			return
		}
		login = checkLogin(w, r)
	}

	can_admin := false
	if login != nil {
		can_admin = login.CanAdmin
	}

	data := HomePageData{
		Login:    login,
		BaseDN:   config.BaseDN,
		Org:      config.Org,
		CanAdmin: can_admin,
		LoggedIn: true,
	}
	templateHome.Execute(w, data)

}
