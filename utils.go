package main

import (
	"fmt"
	"log"

	"github.com/go-ldap/ldap/v3"
	// "bytes"
	// "crypto/rand"
	// "encoding/binary"
	// "encoding/hex"
	// "fmt"
	// "html/template"
	// "log"
	// "net/http"
	// "regexp"
	// "strings"
	// "github.com/emersion/go-sasl"
	// "github.com/emersion/go-smtp"
	// "github.com/gorilla/mux"
	// "golang.org/x/crypto/argon2"
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

func addNewUser(newUser NewUser, config ConfigFile, login LoginStatus) bool {
	log.Printf(fmt.Sprint("Adding New User"))
	// l := openLdap(config)
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
	// if newUser.Member != "" {
	// 	req.Attribute("member", []string{newUser.Member})
	// }
	if newUser.SN != "" {
		req.Attribute("sn", []string{newUser.SN})
	}
	if newUser.Description != "" {
		req.Attribute("description", []string{newUser.Description})
	}
	err := login.conn.Add(req)
	// log.Printf(fmt.Sprintf("71: %v",err))
	// log.Printf(fmt.Sprintf("72: %v",req))
	// log.Printf(fmt.Sprintf("73: %v",newUser))
	if err != nil {
		log.Printf(fmt.Sprintf("75: %v", err))
		return false
	} else {
		return true
	}
}
