package main

import (
	"bytes"
	"crypto/tls"
	"log"
	"net"
	"net/smtp"

	"math/rand"

	"html/template"

	"github.com/go-ldap/ldap/v3"
	// "golang.org/x/text/encoding/unicode"
)

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

func suggestPassword() string {
	password := ""
	chars := "abcdfghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%&*+_-="
	for i := 0; i < 12; i++ {
		password += string([]rune(chars)[rand.Intn(len(chars))])
	}
	return password
}

// Sends an email according to the enclosed information
func sendMail(sendMailTplData SendMailTplData) error {
	log.Printf("sendMail")
	templateMail := template.Must(template.ParseFiles(templatePath + "/" + sendMailTplData.RelTemplatePath))
	buf := bytes.NewBuffer([]byte{})
	err := templateMail.Execute(buf, sendMailTplData)
	message := buf.Bytes()
	auth := smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.SMTPServer)
	log.Printf("auth: %v", auth)
	err = smtp.SendMail(config.SMTPServer+":587", auth, config.SMTPUsername, []string{user.OtherMailbox}, message)
	if err != nil {
		log.Printf("email send error %v", err)
		return err
	}
	log.Printf("Mail sent.")
	return err
}
