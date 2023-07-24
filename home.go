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
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	templateHome := getTemplate("home.html")

	login := checkLogin(w, r)
	if login == nil {
		return
	}

	data := &HomePageData{
		Login:    login,
		BaseDN:   config.BaseDN,
		Org:      config.Org,
		CanAdmin: login.CanAdmin,
	}

	templateHome.Execute(w, data)
}
