package main

import (
	"fmt"
	"log"

	"math/rand"

	"github.com/go-ldap/ldap/v3"
	"golang.org/x/text/encoding/unicode"
)

type NewUser struct {
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

func openLdap(config ConfigFile) *ldap.Conn {
	l, err := ldap.DialURL(config.LdapServerAddr)
	if err != nil {
		log.Printf(fmt.Sprint("Erreur connect LDAP %v", err))
		return nil
	} else {
		return l
	}
}

func suggestPassword() string {
	password := ""
	chars := "abcdfghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%&*+_-="
	for i := 0; i < 12; i++ {
		password += string([]rune(chars)[rand.Intn(len(chars))])
	}
	return password
}

func addNewUser(newUser NewUser, config *ConfigFile, ldapConn *ldap.Conn) bool {
	log.Printf(fmt.Sprint("Adding New User"))
	// l, _ := ldap.DialURL(config.LdapServerAddr)
	// l.Bind(config.NewUserDN, config.NewUserPassword)
	// err := l.StartTLS(&tls.Config{InsecureSkipVerify: true})
	// if err != nil {
	// 	log.Printf(fmt.Sprintf("86: %v", err))
	// }

	// l.Bind(config.)
	dn := newUser.DN
	req := ldap.NewAddRequest(dn, nil)
	req.Attribute("objectClass", []string{"top", "inetOrgPerson"})
	if newUser.DisplayName != "" {
		req.Attribute("displayName", []string{newUser.DisplayName})
	}
	if newUser.GivenName != "" {
		req.Attribute("givenName", []string{newUser.GivenName})
	}
	if newUser.Mail != "" {
		req.Attribute("mail", []string{newUser.Mail})
	}
	if newUser.UID != "" {
		req.Attribute("uid", []string{newUser.UID})
	}
	// if newUser.Member != "" {
	// 	req.Attribute("member", []string{newUser.Member})
	// }
	if newUser.SN != "" {
		req.Attribute("sn", []string{newUser.SN})
	}
	if newUser.Description != "" {
		req.Attribute("description", []string{newUser.Description})
	}
	if newUser.Password != "" {
		pwdEncoded, _ := encodePassword(newUser.Password)
		// if err != nil {
		// 	log.Printf("Error encoding password:  %s", err)
		// 	return err
		// }
		req.Attribute("userPassword", []string{pwdEncoded})
	}

	// conn :=

	err := ldapConn.Add(req)
	log.Printf(fmt.Sprintf("71: %v", err))
	log.Printf(fmt.Sprintf("72: %v", req))
	log.Printf(fmt.Sprintf("73: %v", newUser))
	if err != nil {
		log.Printf(fmt.Sprintf("86: %v", err))
		return false
	} else {
		return true
	}
}

func encodePassword(inPassword string) (string, error) {
	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	return utf16.NewEncoder().String("\"" + inPassword + "\"")
	// if err != nil {
	// 	log.Printf("Error encoding password:  %s", err)
	// 	return err
	// }

}
