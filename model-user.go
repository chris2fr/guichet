/*
Model-User handles everything having to do with the user.
*/
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

func get(user User, config *ConfigFile, ldapConn *ldap.Conn) (*User, error) {
	searchReq := ldap.NewSearchRequest(
		user.DN,
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		"(objectClass=*)",
		[]string{
			"cn",
			"givenName",
			"displayName",
			"uid",
			"sn",
			"mail",
			"description",
			"carLicense",
		},
		nil)
	searchRes, err := ldapConn.Search(searchReq)
	if err != nil {
		log.Printf("get User : %v", err)
		log.Printf("get User : %v", searchReq)
		log.Printf("get User : %v", searchRes)
		return nil, err
	}
	userEntry := searchRes.Entries[0]
	resUser := User{
		DN:           user.DN,
		GivenName:    searchRes.Entries[0].GetAttributeValue("givenName"),
		DisplayName:  searchRes.Entries[0].GetAttributeValue("displayName"),
		Description:  searchRes.Entries[0].GetAttributeValue("description"),
		SN:           searchRes.Entries[0].GetAttributeValue("sn"),
		UID:          searchRes.Entries[0].GetAttributeValue("uid"),
		CN:           searchRes.Entries[0].GetAttributeValue("cn"),
		Mail:         searchRes.Entries[0].GetAttributeValue("mail"),
		OtherMailbox: searchRes.Entries[0].GetAttributeValue("carLicense"),
		CanAdmin:     strings.EqualFold(user.DN, config.AdminAccount),
		CanInvite:    true,
		UserEntry:    userEntry,
	}
	searchReq.BaseDN = config.GroupCanAdmin
	searchReq.Filter = "(member=" + user.DN + ")"
	searchRes, err = ldapConn.Search(searchReq)
	if err == nil {
		if len(searchRes.Entries) > 0 {
			resUser.CanAdmin = true
		}
	}
	return &resUser, nil
}

func add(user User, config *ConfigFile, ldapConn *ldap.Conn) error {
	log.Printf(fmt.Sprint("Adding New User"))

	dn := user.DN
	req := ldap.NewAddRequest(dn, nil)
	req.Attribute("objectClass", []string{"top", "person", "organizationalPerson", "inetOrgPerson"})
	if user.DisplayName != "" {
		req.Attribute("displayName", []string{user.DisplayName})
	}
	if user.GivenName != "" {
		req.Attribute("givenName", []string{user.GivenName})
	}
	if user.Mail != "" {
		req.Attribute("mail", []string{user.Mail})
	}
	if user.UID != "" {
		req.Attribute("uid", []string{user.UID})
	}
	// if user.Member != "" {
	// 	req.Attribute("member", []string{user.Member})
	// }
	if user.SN != "" {
		req.Attribute("sn", []string{user.SN})
	}
	if user.OtherMailbox != "" {
		req.Attribute("carLicense", []string{user.OtherMailbox})
	}
	if user.Description != "" {
		req.Attribute("description", []string{user.Description})
	}
	// if user.Password != "" {
	// 	pwdEncoded, _ := encodePassword(user.Password)
	// 	// if err != nil {
	// 	// 	log.Printf("Error encoding password:  %s", err)
	// 	// 	return err
	// 	// }
	// 	req.Attribute("userPassword", []string{pwdEncoded})
	// }

	// conn :=

	err := ldapConn.Add(req)
	if err != nil {
		log.Printf(fmt.Sprintf("add(User) ldapconn.Add: %v", err))
		log.Printf(fmt.Sprintf("add(User) ldapconn.Add: %v", req))
		log.Printf(fmt.Sprintf("add(User) ldapconn.Add: %v", user))
		//return err
	}
	passwordModifyRequest := ldap.NewPasswordModifyRequest(user.DN, "", user.Password)
	_, err = ldapConn.PasswordModify(passwordModifyRequest)
	if err != nil {
		return err
	}

	// Send the email
	sendMailTplData := SendMailTplData{
		From:            "alice@resdigita.org",
		To:              user.OtherMailbox,
		RelTemplatePath: "user/new.email.txt",
		EmailContentVars: EmailContentVarsTplData{
			InviteFrom:  "alice@resdigita.org",
			SendAddress: "https://www.gvoisins.org",
			Code:        "...",
		},
	}
	err = sendMail(sendMailTplData)
	if err != nil {
		log.Printf("add(user) sendMail: %v", err)
		log.Printf("add(user) sendMail: %v", user)
		log.Printf("add(user) sendMail: %v", sendMailTplData)
	}
	return err
}

func modify(user User, config *ConfigFile, ldapConn *ldap.Conn) error {
	modify_request := ldap.NewModifyRequest(user.DN, nil)
	previousUser, err := get(user, config, ldapConn)
	if err != nil {
		return err
	}
	replaceIfContent(modify_request, "displayName", user.DisplayName, previousUser.DisplayName)
	replaceIfContent(modify_request, "givenName", user.GivenName, previousUser.GivenName)
	replaceIfContent(modify_request, "sn", user.SN, previousUser.SN)
	replaceIfContent(modify_request, "description", user.Description, previousUser.Description)
	err = ldapConn.Modify(modify_request)
	if err != nil {
		log.Printf(fmt.Sprintf("71: %v", err))
		log.Printf(fmt.Sprintf("72: %v", modify_request))
		log.Printf(fmt.Sprintf("73: %v", user))
		return err
	}
	return nil
}

func passwd(user User, config *ConfigFile, ldapConn *ldap.Conn) error {
	passwordModifyRequest := ldap.NewPasswordModifyRequest(user.DN, "", user.Password)
	_, err := ldapConn.PasswordModify(passwordModifyRequest)
	log.Printf(fmt.Sprintf("passwd 162 : %v %v", err, ldapConn))
	log.Printf(fmt.Sprintf("passwd 163 : %v", user))
	return err
}

func bind(user User, config *ConfigFile, ldapConn *ldap.Conn) error {
	return ldapConn.Bind(user.DN, user.Password)
}

// func encodePassword(inPassword string) (string, error) {
// 	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
// 	return utf16.NewEncoder().String("\"" + inPassword + "\"")
// 	// if err != nil {
// 	// 	log.Printf("Error encoding password:  %s", err)
// 	// 	return err
// 	// }

// }
