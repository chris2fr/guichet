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

	// "github.com/emersion/go-sasl"
	// "github.com/emersion/go-smtp"
	"net/smtp"

	"github.com/go-ldap/ldap/v3"
	// "strings"
	b64 "encoding/base64"
)

// type InvitationAccount struct {
// 	UID string
// 	Password string
// 	BaseDN string
// }

// var EMAIL_REGEXP := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func passwordLost(user User, config *ConfigFile, ldapConn *ldap.Conn) error {
	if user.CN == "" && user.Mail == "" && user.OtherMailbox == "" {
		return errors.New("Il n'y a pas de quoi identifier l'utilisateur")
	}
	searchFilter := "(|"
	if user.CN != "" {
		searchFilter += "(cn=" + user.UID + ")"
	}
	if user.Mail != "" {
		searchFilter += "(mail=" + user.Mail + ")"
	}
	if user.OtherMailbox != "" {
		searchFilter += "(carLicense=" + user.OtherMailbox + ")"
	}
	searchFilter += ")"
	searchReq := ldap.NewSearchRequest(config.UserBaseDN, ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false, searchFilter, []string{"cn", "uid", "mail", "carLicense", "sn", "displayName", "givenName"}, nil)
	searchRes, err := ldapConn.Search(searchReq)
	if err != nil {
		log.Printf("passwordLost search : %v %v", err, ldapConn)
		log.Printf("passwordLost search : %v", searchReq)
		log.Printf("passwordLost search : %v", searchRes)
		log.Printf("passwordLost search: %v", user)
		return err
	}
	if len(searchRes.Entries) == 0 {
		log.Printf("Il n'y a pas d'utilisateur qui correspond %v", searchReq)
		return errors.New("Il n'y a pas d'utilisateur qui correspond")
	}
	// log.Printf("passwordLost 58 : %v", user)
	// log.Printf("passwordLost 59 : %v", searchRes.Entries[0])
	// log.Printf("passwordLost 60 : %v", searchRes.Entries[0].GetAttributeValue("cn"))
	// log.Printf("passwordLost 61 : %v", searchRes.Entries[0].GetAttributeValue("uid"))
	// log.Printf("passwordLost 62 : %v", searchRes.Entries[0].GetAttributeValue("mail"))
	// log.Printf("passwordLost 63 : %v", searchRes.Entries[0].GetAttributeValue("carLicense"))
	// Préparation du courriel à envoyer
	user.Password = suggestPassword()
	code := b64.URLEncoding.EncodeToString([]byte(user.UID + ";" + user.Password))
	user.DN = "uid=" + searchRes.Entries[0].GetAttributeValue("cn") + "," + config.InvitationBaseDN
	user.UID = searchRes.Entries[0].GetAttributeValue("cn")
	user.CN = searchRes.Entries[0].GetAttributeValue("cn")
	user.Mail = searchRes.Entries[0].GetAttributeValue("mail")
	user.OtherMailbox = searchRes.Entries[0].GetAttributeValue("carLicense")
	/* Check for outstanding invitation */
	searchReq = ldap.NewSearchRequest(config.InvitationBaseDN, ldap.ScopeSingleLevel,
		ldap.NeverDerefAliases, 0, 0, false, "(uid="+user.UID+")", []string{"seeAlso"}, nil)
	searchRes, err = ldapConn.Search(searchReq)
	if err != nil {
		log.Printf(fmt.Sprintf("passwordLost (Check existing invitation) : %v", err))
		log.Printf(fmt.Sprintf("passwordLost (Check existing invitation) : %v", user))
		return err
	}
	// if len(searchRes.Entries) == 0 {
	/* Add the invitation */
	addReq := ldap.NewAddRequest(
		"uid="+user.UID+","+config.InvitationBaseDN,
		nil)
	addReq.Attribute("objectClass", []string{"top", "account", "simpleSecurityObject"})
	addReq.Attribute("uid", []string{user.UID})
	addReq.Attribute("userPassword", []string{suggestPassword()})
	addReq.Attribute("seeAlso", []string{config.UserNameAttr + "=" + user.UID + "," + config.InvitationBaseDN})
	err = ldapConn.Add(addReq)
	if err != nil {
		log.Printf("passwordLost 83 : %v", err)
		log.Printf("passwordLost 84 : %v", user)

		log.Printf("passwordLost 84 : %v", addReq)
		// // log.Printf("passwordLost 85 : %v", searchRes.Entries[0]))
		// // For some reason I get here even if the entry exists already
		// return err
	}
	// }
	ldapNewConn, err := openNewUserLdap(config)
	if err != nil {
		log.Printf("passwordLost openNewUserLdap : %v", err)
	}
	err = passwd(user, config, ldapNewConn)
	if err != nil {
		log.Printf("passwordLost passwd : %v", err)
		log.Printf("passwordLost passwd : %v", user)
		log.Printf("passwordLost passwd : %v", searchRes.Entries[0])
		return err
	}
	templateMail := template.Must(template.ParseFiles(templatePath + "/passwd/lost_password_email.txt"))
	buf := bytes.NewBuffer([]byte{})
	templateMail.Execute(buf, &CodeMailFields{
		To:             user.OtherMailbox,
		From:           config.MailFrom,
		InviteFrom:     user.UID,
		Code:           code,
		WebBaseAddress: config.WebAddress,
	})
	// message := []byte("Hi " + user.OtherMailbox)
	log.Printf("Sending mail to: %s", user.OtherMailbox)
	// var auth sasl.Client = nil
	// if config.SMTPUsername != "" {
	// 	auth = sasl.NewPlainClient("", config.SMTPUsername, config.SMTPPassword)
	// }
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

func passwordFound(user User, config *ConfigFile, ldapConn *ldap.Conn) (string, error) {
	l, err := openLdap(config)
	if err != nil {
		log.Printf("passwordFound openLdap %v", err)
		// log.Printf("passwordFound openLdap Config : %v", config)
		return "", err
	}
	if user.DN == "" && user.UID != "" {
		user.DN = "uid=" + user.UID + "," + config.InvitationBaseDN
	}
	err = l.Bind(user.DN, user.Password)
	if err != nil {
		log.Printf("passwordFound l.Bind %v", err)
		log.Printf("passwordFound l.Bind %v", user.DN)
		log.Printf("passwordFound l.Bind %v", user.UID)
		return "", err
	}
	searchReq := ldap.NewSearchRequest(user.DN, ldap.ScopeBaseObject,
		ldap.NeverDerefAliases, 0, 0, false, "(uid="+user.UID+")", []string{"seeAlso"}, nil)
	var searchRes *ldap.SearchResult
	searchRes, err = ldapConn.Search(searchReq)
	if err != nil {
		log.Printf("passwordFound %v", err)
		log.Printf("passwordFound %v", searchReq)
		log.Printf("passwordFound %v", ldapConn)
		log.Printf("passwordFound %v", searchRes)
		return "", err
	}
	if len(searchRes.Entries) == 0 {
		log.Printf("passwordFound %v", err)
		log.Printf("passwordFound %v", searchReq)
		log.Printf("passwordFound %v", ldapConn)
		log.Printf("passwordFound %v", searchRes)
		return "", err
	}
	delReq := ldap.NewDelRequest("uid="+user.CN+","+config.InvitationBaseDN, nil)
	ldapConn.Del(delReq)
	return searchRes.Entries[0].GetAttributeValue("seeAlso"), err
}
