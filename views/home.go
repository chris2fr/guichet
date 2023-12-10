/*
home show the home page
*/

package views

import (
	"net/http"
)

func HandleHome(w http.ResponseWriter, r *http.Request) {
	templateHome := getTemplate("home.html")

	login := checkLogin(w, r)
	if login == nil {
		status, _ := HandleLogin(w, r)
		if status == nil {
			return
		}
		login = checkLogin(w, r)
	}

	can_admin := false
	if login != nil {
		can_admin = login.Common.CanAdmin
	}

	data := HomePageData{
		Login: NestedLoginTplData{
			Login: login,
		},
		BaseDN: config.BaseDN,
		Org:    config.Org,
		Common: NestedCommonTplData{
			CanAdmin:  can_admin,
			CanInvite: true,
			LoggedIn:  true,
		},
	}
	execTemplate(w, templateHome, data.Common, data.Login, data)
	// templateHome.Execute(w, data)
		login.conn.Close()

}
