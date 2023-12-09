package models

import (
	"bytes"
	"crypto/tls"
	"log"
	"net"
	"net/smtp"

	"html/template"

	"github.com/go-ldap/ldap/v3"
	// "golang.org/x/text/encoding/unicode"
	"encoding/json"

	"io/ioutil"

	"os"

	"flag"
)

//

func ReadConfig() ConfigFile {
	// Default configuration values for certain fields
	flag.Parse()
	var configFlag = flag.String("config", "./config.json", "Configuration file path")

	config_file := ConfigFile{
		HttpBindAddr:   ":9992",
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



type EmailContentVarsTplData struct {
	Code        string
	SendAddress string
	InviteFrom  string
}

// Data to be passed to an email for sending
type SendMailTplData struct {
	// Sender of the email
	To string
	// Receiver of the email
	From string
	// Relative path (without leading /) to the email template in the templates folder
	// usually ending in .txt
	RelTemplatePath string
	// Variables to be included in the template of the email
	EmailContentVars EmailContentVarsTplData
}



func openLdap(config *ConfigFile) (*ldap.Conn, error) {
	var ldapConn *ldap.Conn
	var err error
	if config.LdapTLS {
		tlsConf := &tls.Config{
			ServerName:         config.LdapServerAddr,
			InsecureSkipVerify: true,
		}
		ldapConn, err = ldap.DialTLS("tcp", net.JoinHostPort(config.LdapServerAddr, "636"), tlsConf)
	} else {
		ldapConn, err = ldap.DialURL("ldap://" + config.LdapServerAddr)
	}
	if err != nil {
		log.Printf("openLDAP %v", err)
		log.Printf("openLDAP %v", config.LdapServerAddr)
	}
	return ldapConn, err

	// l, err := ldap.DialURL(config.LdapServerAddr)
	// if err != nil {
	// 	log.Printf(fmt.Sprint("Erreur connect LDAP %v", err))
	// 	log.Printf(fmt.Sprint("Erreur connect LDAP %v", config.LdapServerAddr))
	// 	return nil
	// } else {
	// 	return l
	// }
}



// Sends an email according to the enclosed information
func sendMail(sendMailTplData SendMailTplData) error {
	log.Printf("sendMail")
	templateMail := template.Must(template.ParseFiles( "./templates" + sendMailTplData.RelTemplatePath))
	buf := bytes.NewBuffer([]byte{})
	err := templateMail.Execute(buf, sendMailTplData)
	message := buf.Bytes()
	auth := smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.SMTPServer)
	log.Printf("auth: %v", auth)
	err = smtp.SendMail(config.SMTPServer+":587", auth, config.SMTPUsername, []string{sendMailTplData.To}, message)
	if err != nil {
		log.Printf("sendMail smtp.SendMail %v", err)
		log.Printf("sendMail smtp.SendMail %v", sendMailTplData)
		return err
	}
	log.Printf("Mail sent.")
	return err
}
