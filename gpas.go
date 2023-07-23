/*
gpas is GVoisin password reset
*/

package main

import (
	"errors"
	"fmt"
	"log"

	// "github.com/emersion/go-sasl"
	// "github.com/emersion/go-smtp"
	"net/smtp"

	"github.com/go-ldap/ldap/v3"
	// "strings"
)

// var EMAIL_REGEXP := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func passwordLost(user User, config *ConfigFile, ldapConn *ldap.Conn) error {
	if user.CN == "" && user.Mail == "" && user.OtherMailbox == "" {
		return errors.New("Il n'y a pas de quoi identifier l'utilisateur")
	}
	searchFilter := "(|"
	if user.CN != "" {
		searchFilter += "(cn=" + user.CN + ")"
	}
	if user.Mail != "" {
		searchFilter += "(mail=" + user.Mail + ")"
	}
	if user.OtherMailbox != "" {
		searchFilter += "(carLicense=" + user.OtherMailbox + ")"
	}
	searchFilter += ")"
	searchReq := ldap.NewSearchRequest(config.UserBaseDN, ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false, searchFilter, []string{"cn", "uid", "mail", "carLicense"}, nil)
	searchRes, err := ldapConn.Search(searchReq)
	if err != nil {
		log.Printf(fmt.Sprintf("passwordLost : %v %v", err, ldapConn))
		log.Printf(fmt.Sprintf("passwordLost : %v", searchReq))
		log.Printf(fmt.Sprintf("passwordLost : %v", user))
		return errors.New("LDAP chose")
	}
	if len(searchRes.Entries) == 0 {
		return errors.New("Il n'y a pas d'utilisateur qui correspond")
	}
	// Préparation du courriel à envoyer
	// code := "GPas"
	// templateMail := template.Must(template.ParseFiles(templatePath + "/invite_mail.txt"))
	// buf := bytes.NewBuffer([]byte{})
	// templateMail.Execute(buf, &CodeMailFields{
	// 	To:             user.OtherMailbox,
	// 	From:           config.MailFrom,
	// 	InviteFrom:     "GPas",
	// 	Code:           code,
	// 	WebBaseAddress: config.WebAddress,
	// })
	message := []byte("Hi " + user.OtherMailbox)
	log.Printf("Sending mail to: %s", user.OtherMailbox)
	// var auth sasl.Client = nil
	// if config.SMTPUsername != "" {
	// 	auth = sasl.NewPlainClient("", config.SMTPUsername, config.SMTPPassword)
	// }
	auth := smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.SMTPServer)
	log.Printf("auth: %v", auth)
	err = smtp.SendMail(config.SMTPServer+":587", auth, config.SMTPUsername, []string{user.OtherMailbox}, message)
	if err != nil {
		return err
	}
	log.Printf("Mail sent.")
	return nil
}
