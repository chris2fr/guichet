/*
Routes the requests to the app
*/
package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

type ConfigFile struct {
	HttpBindAddr   string `json:"http_bind_addr"`
	LdapServerAddr string `json:"ldap_server_addr"`
	LdapTLS        bool   `json:"ldap_tls"`

	BaseDN        string `json:"base_dn"`
	UserBaseDN    string `json:"user_base_dn"`
	UserNameAttr  string `json:"user_name_attr"`
	GroupBaseDN   string `json:"group_base_dn"`
	GroupNameAttr string `json:"group_name_attr"`

	MailingBaseDN       string `json:"mailing_list_base_dn"`
	MailingNameAttr     string `json:"mailing_list_name_attr"`
	MailingGuestsBaseDN string `json:"mailing_list_guest_user_base_dn"`

	InvitationBaseDN   string   `json:"invitation_base_dn"`
	InvitationNameAttr string   `json:"invitation_name_attr"`
	InvitedMailFormat  string   `json:"invited_mail_format"`
	InvitedAutoGroups  []string `json:"invited_auto_groups"`

	WebAddress   string `json:"web_address"`
	MailFrom     string `json:"mail_from"`
	SMTPServer   string `json:"smtp_server"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`

	AdminAccount   string `json:"admin_account"`
	GroupCanInvite string `json:"group_can_invite"`
	GroupCanAdmin  string `json:"group_can_admin"`

	S3AdminEndpoint string `json:"s3_admin_endpoint"`
	S3AdminToken    string `json:"s3_admin_token"`

	S3Endpoint  string `json:"s3_endpoint"`
	S3AccessKey string `json:"s3_access_key"`
	S3SecretKey string `json:"s3_secret_key"`
	S3Region    string `json:"s3_region"`
	S3Bucket    string `json:"s3_bucket"`

	Org             string `json:"org"`
	DomainName      string `json:"domain_name"`
	NewUserDN       string `json:"new_user_dn"`
	NewUserPassword string `json:"new_user_password"`
}

var staticPath = "./static"

/*
Create the different routes
*/
func makeGVRouter() (*mux.Router, error) {
	r := mux.NewRouter()
	r.HandleFunc("/", handleHome)

	r.HandleFunc("/session/logout", handleLogout)

	r.HandleFunc("/user", handleProfile)
	r.HandleFunc("/user/new", handleInviteNewAccount)

	r.HandleFunc("/picture/{name}", handleDownloadPicture)

	r.HandleFunc("/passwd", handlePasswd)
	r.HandleFunc("/passwd/lost", handleLostPassword)
	r.HandleFunc("/passwd/lost/{code}", handleFoundPassword)

	r.HandleFunc("/admin", handleHome)
	r.HandleFunc("/admin/activate", handleAdminActivateUsers)
	r.HandleFunc("/admin/unactivate/{cn}", handleAdminUnactivateUser)
	r.HandleFunc("/admin/activate/{cn}", handleAdminActivateUser)
	r.HandleFunc("/admin/users", handleAdminUsers)
	r.HandleFunc("/admin/groups", handleAdminGroups)
	r.HandleFunc("/admin/ldap/{dn}", handleAdminLDAP)
	r.HandleFunc("/admin/create/{template}/{super_dn}", handleAdminCreate)

	// r.HandleFunc("/directory/search", handleDirectorySearch)
	// r.HandleFunc("/directory", handleDirectory)
	// r.HandleFunc("/garage/key", handleGarageKey)
	// r.HandleFunc("/garage/website", handleGarageWebsiteList)
	// r.HandleFunc("/garage/website/new", handleGarageWebsiteNew)
	// r.HandleFunc("/garage/website/b/{bucket}", handleGarageWebsiteInspect)

	// r.HandleFunc("/user/send_code", handleInviteSendCode)

	// r.HandleFunc("/invitation/{code}", handleInvitationCode)

	// r.HandleFunc("/admin-mailing", handleAdminMailing)
	// r.HandleFunc("/admin/mailing/{id}", handleAdminMailingList)

	staticFiles := http.FileServer(http.Dir(staticPath))
	r.Handle("/static/{file:.*}", http.StripPrefix("/static/", staticFiles))

	// log.Printf("Starting HTTP server on %s", config.HttpBindAddr)
	err := http.ListenAndServe(config.HttpBindAddr, logRequest(r))

	return r, err
}
