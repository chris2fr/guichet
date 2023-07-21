/*
config handles reading the config.json file at the root and processing the settings
*/
package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
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

var configFlag = flag.String("config", "./config.json", "Configuration file path")

var config *ConfigFile

func readConfig() ConfigFile {
	// Default configuration values for certain fields
	config_file := ConfigFile{
		HttpBindAddr:   ":9991",
		LdapServerAddr: "ldap://127.0.0.1:389",

		UserNameAttr:  "uid",
		GroupNameAttr: "gid",

		InvitationNameAttr: "cn",
		InvitedAutoGroups:  []string{},

		Org: "ResDigita",
	}

	_, err := os.Stat(*configFlag)
	if os.IsNotExist(err) {
		log.Fatalf("Could not find Guichet configuration file at %s. Please create this file, for exemple starting with config.json.exemple and customizing it for your deployment.", *configFlag)
	}

	if err != nil {
		log.Fatal(err)
	}

	bytes, err := ioutil.ReadFile(*configFlag)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(bytes, &config_file)
	if err != nil {
		log.Fatal(err)
	}

	return config_file
}
