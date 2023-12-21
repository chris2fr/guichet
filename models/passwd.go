/*
gpas is LesGrandsVoisins password reset
*/

package models

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math/rand"

	// "github.com/emersion/go-sasl"
	// "github.com/emersion/go-smtp"
	"net/smtp"

	"github.com/go-ldap/ldap/v3"

	// "strings"
	b64 "encoding/base64"
)

var templatePath = "./templates"

type CodeMailFields struct {
	From           string
	To             string
	Code           string
	InviteFrom     string
	WebBaseAddress string
	Common         NestedCommonTplData
}
type NestedCommonTplData struct {
	Error          string
	ErrorMessage   string
	CanAdmin       bool
	CanInvite      bool
	LoggedIn       bool
	Success        bool
	WarningMessage string
	WebsiteName    string
	WebsiteURL     string
}
// type InvitationAccount struct {
// 	UID string
// 	Password string
// 	BaseDN string
// }

// Suggesting a 12 char password with some excentrics
func SuggestPassword() string {
	password := ""
	chars := "abcdfghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%&*+_-="
	for i := 0; i < 12; i++ {
		password += string([]rune(chars)[rand.Intn(len(chars))])
	}
	return password
}

// var EMAIL_REGEXP := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func getInvitationBaseDN(config *ConfigFile) string {
	return config.InvitationBaseDN
}

func PasswordLost(searchQuery string, config *ConfigFile, ldapConn *ldap.Conn) error {
	searchFilter := fmt.Sprintf("(|(uid=%v)(cn=%v)(mail=%v))",searchQuery,searchQuery,searchQuery)
	searchReq := ldap.NewSearchRequest(config.UserBaseDN, ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false, searchFilter, []string{"cn", "uid", "mail", "carLicense", "sn", "displayName", "givenName"}, nil)
	searchRes, err := ldapConn.Search(searchReq)
	if err != nil {
		log.Printf("PasswordLost search : %v %v", err, ldapConn)
		log.Printf("PasswordLost search : %v", searchReq)
		log.Printf("PasswordLost search : %v", searchRes)
		log.Printf("PasswordLost search: %v", searchQuery)
		return err
	}
	if len(searchRes.Entries) == 0 {
		log.Printf("Il n'y a pas d'utilisateur qui correspond %v", searchReq)
		return errors.New("Il n'y a pas d'utilisateur qui correspond")
	}
	// log.Printf("PasswordLost 58 : %v", searchQuery)
	// log.Printf("PasswordLost 59 : %v", searchRes.Entries[0])
	// log.Printf("PasswordLost 60 : %v", searchRes.Entries[0].GetAttributeValue("cn"))
	// log.Printf("PasswordLost 61 : %v", searchRes.Entries[0].GetAttributeValue("uid"))
	// log.Printf("PasswordLost 62 : %v", searchRes.Entries[0].GetAttributeValue("mail"))
	// log.Printf("PasswordLost 63 : %v", searchRes.Entries[0].GetAttributeValue("carLicense"))
	// Préparation du courriel à envoyer

	delReq := ldap.NewDelRequest("uid="+searchRes.Entries[0].GetAttributeValue("uid")+","+config.InvitationBaseDN, nil)
	err = ldapConn.Del(delReq)
  user := User{
		Password: SuggestPassword(),
		DN: "uid=" + searchRes.Entries[0].GetAttributeValue("uid") + "," + config.InvitationBaseDN,
		UID: searchRes.Entries[0].GetAttributeValue("uid"),
		CN: searchRes.Entries[0].GetAttributeValue("cn"),
		Mail: searchRes.Entries[0].GetAttributeValue("mail"),
		OtherMailbox: searchRes.Entries[0].GetAttributeValue("carLicense"),
		SeeAlso: searchRes.Entries[0].DN,
	}

	code := b64.URLEncoding.EncodeToString([]byte(user.UID + ";" + user.Password))
	/* Check for outstanding invitation */
	searchReq = ldap.NewSearchRequest(config.InvitationBaseDN, ldap.ScopeSingleLevel,
		ldap.NeverDerefAliases, 0, 0, false, "(uid="+user.UID+")", []string{"seeAlso"}, nil)
	searchRes, err = ldapConn.Search(searchReq)
	if err != nil {
		log.Printf(fmt.Sprintf("PasswordLost (Check existing invitation) : %v", err))
		log.Printf(fmt.Sprintf("PasswordLost (Check existing invitation) : %v", user))
		return err
	}
	// if len(searchRes.Entries) == 0 {
	/* Add the invitation */
	addReq := ldap.NewAddRequest(
		"uid="+user.UID+","+config.InvitationBaseDN,
		nil)
	addReq.Attribute("objectClass", []string{"top", "account", "simpleSecurityObject"})
	addReq.Attribute("uid", []string{user.UID})
	addReq.Attribute("userPassword", []string{user.Password})
	addReq.Attribute("seeAlso", []string{user.SeeAlso})
	//addReq.Attribute("seeAlso", []string{config.UserNameAttr + "=" + user.CN + "," + config.UserBaseDN})
	// Password invitation may already exist
	//
	err = ldapConn.Add(addReq)
	if err != nil {
		log.Printf("PasswordLost 83 : %v", err)
		log.Printf("PasswordLost 84 : %v", user)

		log.Printf("PasswordLost 84 : %v", addReq)
		// // log.Printf("PasswordLost 85 : %v", searchRes.Entries[0]))
		// // For some reason I get here even if the entry exists already
		// return err
	}
	// }
	ldapNewConn, err := OpenNewUserLdap(config)
	if err != nil {
		log.Printf("PasswordLost OpenNewUserLdap : %v", err)
	}
	err = PassWD(user, config, ldapNewConn)
	if err != nil {
		log.Printf("PasswordLost PassWD : %v", err)
		log.Printf("PasswordLost PassWD : %v", user)
		log.Printf("PasswordLost PassWD : %v", searchRes.Entries[0])
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

func PasswordFound(user User, config *ConfigFile, ldapConn *ldap.Conn) (string, error) {
	l, err := openLdap(config)
	if err != nil {
		log.Printf("PasswordFound openLdap %v", err)
		// log.Printf("PasswordFound openLdap Config : %v", config)
		l.Close()
		return "", err
	}
	if user.DN == "" && user.UID != "" {
		user.DN = "uid=" + user.UID + "," + config.InvitationBaseDN
	}
	err = l.Bind(user.DN, user.Password)
	if err != nil {
		log.Printf("PasswordFound l.Bind %v", err)
		log.Printf("PasswordFound l.Bind %v", user.DN)
		log.Printf("PasswordFound l.Bind %v", user.UID)
		l.Close()
		return "", err
	} else {
		log.Printf("models.PasswordFound Bind OK  %v", user.UID)
	}
	searchReq := ldap.NewSearchRequest(user.DN, ldap.ScopeBaseObject,
		ldap.NeverDerefAliases, 0, 0, false, "(uid="+user.UID+")", []string{"seeAlso"}, nil)
	var searchRes *ldap.SearchResult
	searchRes, err = ldapConn.Search(searchReq)
	if err != nil {
		log.Printf("PasswordFound %v", err)
		log.Printf("PasswordFound %v", searchReq)
		log.Printf("PasswordFound %v", ldapConn)
		log.Printf("PasswordFound %v", searchRes)
		l.Close()
		return "", err
	}  else {
		log.Printf("models.PasswordFound searchRes OK  %v", user.UID)
	}
	if len(searchRes.Entries) == 0 {
		log.Printf("PasswordFound %v", err)
		log.Printf("PasswordFound %v", searchReq)
		log.Printf("PasswordFound %v", ldapConn)
		log.Printf("PasswordFound %v", searchRes)
		l.Close()
		return "", err
	}  else {
		log.Printf("models.PasswordFound searchRes > 0  %v", user.UID)
	}
	delReq := ldap.NewDelRequest("uid="+user.CN+","+config.InvitationBaseDN, nil)
	ldapConn.Del(delReq)
	l.Close()
	return searchRes.Entries[0].GetAttributeValue("seeAlso"), err
}
