/*
gpas is GVoisin password reset
*/

package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/go-ldap/ldap/v3"
	// "strings"
)

// var EMAIL_REGEXP := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func passwordLost(user User, config *ConfigFile, ldapConn *ldap.Conn) error {
	if user.CN == "" && user.Mail == "" && user.OtherMailbox == "" {
		return errors.New("Il n'y a pas de quoi identifier l'utilisateur")
	}
	searchFilter := "(|"
	if user.CN == "" {
		searchFilter += "(cn=" + user.CN + ")"
	}
	if user.Mail == "" {
		searchFilter += "(mail=" + user.Mail + ")"
	}
	if user.OtherMailbox == "" {
		searchFilter += "(carLicense=" + user.OtherMailbox + ")"
	}
	searchReq := ldap.NewSearchRequest(config.UserBaseDN, ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false, "(|()()())", []string{"cn", "uid", "mail", "carLicense"}, nil)
	searchRes, err := ldapConn.Search(searchReq)
	if err != nil {
		log.Printf(fmt.Sprintf("passwordLost : %v %v", err, ldapConn))
	}
	if len(searchRes.Entries) == 0 {
		return errors.New("Il n'y a pas d'utilisateur qui correspond")
	}
	// Préparation du courriel à envoyer
	code := "GPas"
	templateMail := template.Must(template.ParseFiles(templatePath + "/invite_mail.txt"))
	buf := bytes.NewBuffer([]byte{})
	templateMail.Execute(buf, &CodeMailFields{
		To:             user.OtherMailbox,
		From:           config.MailFrom,
		InviteFrom:     "GPas",
		Code:           code,
		WebBaseAddress: config.WebAddress,
	})
	log.Printf("Sending mail to: %s", user.OtherMailbox)
	var auth sasl.Client = nil
	if config.SMTPUsername != "" {
		auth = sasl.NewPlainClient("", config.SMTPUsername, config.SMTPPassword)
	}
	err = smtp.SendMail(config.SMTPServer, auth, config.MailFrom, []string{user.OtherMailbox}, buf)
	if err != nil {
		return err
	}
	log.Printf("Mail sent.")
	return nil
}