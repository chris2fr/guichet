package main

import (
	"crypto/tls"
	"net"

	"math/rand"

	"github.com/go-ldap/ldap/v3"
	// "golang.org/x/text/encoding/unicode"
)

func openLdap(config *ConfigFile) (*ldap.Conn, error) {
	if config.LdapTLS {
		tlsConf := &tls.Config{
			ServerName:         config.LdapServerAddr,
			InsecureSkipVerify: true,
		}
		return ldap.DialTLS("tcp", net.JoinHostPort(config.LdapServerAddr, "636"), tlsConf)
	} else {
		return ldap.DialURL("ldap://" + config.LdapServerAddr)
	}

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
