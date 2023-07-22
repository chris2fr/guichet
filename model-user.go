/*
Model-User handles everything having to do with the user.
*/
package main

import (
	"fmt"
	"log"

	"github.com/go-ldap/ldap/v3"
)

/*
Represents a user
*/
type User struct {
	DN          string
	CN          string
	GivenName   string
	DisplayName string
	Mail        string
	SN          string
	UID         string
	Description string
	Password    string
}

func add(user User, config *ConfigFile, ldapConn *ldap.Conn) error {
	log.Printf(fmt.Sprint("Adding New User"))

	dn := user.DN
	req := ldap.NewAddRequest(dn, nil)
	req.Attribute("objectClass", []string{"top", "inetOrgPerson"})
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
		log.Printf(fmt.Sprintf("71: %v", err))
		log.Printf(fmt.Sprintf("72: %v", req))
		log.Printf(fmt.Sprintf("73: %v", user))
		return err
	} else {
		passwordModifyRequest := ldap.NewPasswordModifyRequest(user.DN, "", user.Password)
		_, err = ldapConn.PasswordModify(passwordModifyRequest)
		if err != nil {
			return err
		}
		return nil
	}
}

func modify(user User, config *ConfigFile, ldapConn *ldap.Conn) error {
	modify_request := ldap.NewModifyRequest(user.DN, nil)
	replaceIfContent(modify_request, "displayName", user.DisplayName)
	replaceIfContent(modify_request, "givenName", user.GivenName)
	replaceIfContent(modify_request, "sn", user.SN)
	replaceIfContent(modify_request, "description", user.Description)
	err := ldapConn.Modify(modify_request)
	if err != nil {
		return err
	}
	return nil
}

func replaceIfContent(modifReq *ldap.ModifyRequest, key string, value string) error {
	if value != "" {
		modifReq.Replace(key, []string{value})
	}
	return nil
}

// func encodePassword(inPassword string) (string, error) {
// 	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
// 	return utf16.NewEncoder().String("\"" + inPassword + "\"")
// 	// if err != nil {
// 	// 	log.Printf("Error encoding password:  %s", err)
// 	// 	return err
// 	// }

// }
