package main

import (
	"fmt"
	"log"

	"math/rand"

	"github.com/go-ldap/ldap/v3"
	// "golang.org/x/text/encoding/unicode"
)

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
