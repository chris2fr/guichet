/*
home show the home page
*/

package views

import (
	"net/http"
)

func HandleHome(w http.ResponseWriter, r *http.Request) {
	templateHome := getTemplate("home.html")
	loggedin, can_admin, info := PocketLoginCheck(w,r)
	if !loggedin {
		templateLogin := getTemplate("login.html")
		execTemplate(w, templateLogin, NestedCommonTplData{
			CanAdmin:  false,
			CanInvite: true,
			LoggedIn:  false}, NestedLoginTplData{}, LoginFormData{
			Common: NestedCommonTplData{
				CanAdmin:  false,
				CanInvite: true,
				LoggedIn:  false}})
		// templateLogin.Execute(w, LoginFormData{
		// 	Common: NestedCommonTplData{
		// 		CanAdmin:  false,
		// 		CanInvite: true,
		// 		LoggedIn:  false}})
		return
	}
	// if ! loggedin {
	// 	status, _ := HandleLogin(w, r)
	// 	if status == nil {
	// 		return
	// 	}
	// 	loggedin, can_admin, info = PocketLoginCheck(w,r)
	// }
	// login := checkLogin(w, r)
	// if login == nil {
	// 	status, _ := HandleLogin(w, r)
	// 	if status == nil {
	// 		return
	// 	}
	// 	login = checkLogin(w, r)
	// }

	// can_admin := false
	// if login != nil {
	// 	can_admin = login.Common.CanAdmin
	// }
	

	data := HomePageData{
		Login: NestedLoginTplData{
			// Login: login,
			Login: &LoginStatus{
				Info: info,
			},
		},
		BaseDN: config.BaseDN,
		Org:    config.Org,
		Common: NestedCommonTplData{
			CanAdmin:  can_admin,
			CanInvite: loggedin,
			LoggedIn:  loggedin,
		},
	}
	execTemplate(w, templateHome, data.Common, data.Login, data)
	// templateHome.Execute(w, data)
		// login.conn.Close()

}
