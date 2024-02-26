/*
Routes the requests to the app Guichet
*/
package controllers

import (
	"guichet/models"
	"guichet/views"
	"net/http"

	"github.com/dchest/captcha"
	"github.com/gorilla/mux"
	// "log"
)

// const (
// 	// Default number of digits in captcha solution.
// 	DefaultLen = 6
// 	// The number of captchas created that triggers garbage collection used
// 	// by default store.
// 	CollectNum = 100
// 	// Expiration time of captchas used by default store.
// 	Expiration = 10 * time.Minute
// )

const (
	// Standard width and height of a captcha image.
	StdWidth  = 240
	StdHeight = 80
)


var staticPath = "./static"
var config = models.ReadConfig()

/*
Create the different routes
*/
func MakeGVRouter() (*mux.Router, error) {


	r := mux.NewRouter()
	r.HandleFunc("/", views.HandleHome)

	r.HandleFunc("/session/logout", views.HandleLogout)

	r.HandleFunc("/user", views.HandleUser)
	r.HandleFunc("/user/new", views.HandleInviteNewAccount)
	r.HandleFunc("/user/new/", views.HandleInviteNewAccount)
	r.HandleFunc("/user/wait", views.HandleUserWait)
	r.HandleFunc("/user/mail", views.HandleUserMail)

	r.HandleFunc("/picture/{name}", views.HandleDownloadPicture)

	r.HandleFunc("/passwd", views.HandlePasswd)
	r.HandleFunc("/passwd/lost", views.HandleLostPassword)
	r.HandleFunc("/passwd/lost/{code}", views.HandleFoundPassword)

	r.HandleFunc("/admin", views.HandleHome)
	r.HandleFunc("/admin/activate", views.HandleAdminActivateUsers)
	r.HandleFunc("/admin/unactivate/{cn}", views.HandleAdminUnactivateUser)
	r.HandleFunc("/admin/activate/{cn}", views.HandleAdminActivateUser)
	r.HandleFunc("/admin/users", views.HandleAdminUsers)
	r.HandleFunc("/admin/groups", views.HandleAdminGroups)
	r.HandleFunc("/admin/ldap/{dn}", views.HandleAdminLDAP)
	r.HandleFunc("/admin/create/{template}/{super_dn}", views.HandleAdminCreate)

	r.Handle("/captcha/", captcha.Server(StdWidth, StdHeight))
	// 	//api for create captcha
	// 	r.HandleFunc("/api/getCaptcha", views.GenerateCaptchaHandler)

	// 	//api for verify captcha
	// 	r.HandleFunc("/api/verifyCaptcha", views.CaptchaVerifyHandle)
	// // if err := http.ListenAndServe("localhost:8666", nil); err != nil {
	// // 	log.Fatal(err)
	// // }

	// r.HandleFunc("/directory/search", views.HandleDirectorySearch)
	// r.HandleFunc("/directory", views.HandleDirectory)
	// r.HandleFunc("/garage/key", views.HandleGarageKey)
	// r.HandleFunc("/garage/website", views.HandleGarageWebsiteList)
	// r.HandleFunc("/garage/website/new", views.HandleGarageWebsiteNew)
	// r.HandleFunc("/garage/website/b/{bucket}", views.HandleGarageWebsiteInspect)

	// r.HandleFunc("/user/send_code", views.HandleInviteSendCode)

	// r.HandleFunc("/invitation/{code}", views.HandleInvitationCode)

	// r.HandleFunc("/admin-mailing", views.HandleAdminMailing)
	// r.HandleFunc("/admin/mailing/{id}", views.HandleAdminMailingList)

	staticFiles := http.FileServer(http.Dir(staticPath))
	r.Handle("/static/{file:.*}", http.StripPrefix("/static/", staticFiles))


	// log.Printf("Starting HTTP server on %s", config.HttpBindAddr)
	err := http.ListenAndServe(config.HttpBindAddr, views.LogRequest(r))

	return r, err
}


