/*
http-utils provide utility functions that interact with http
*/

package main

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"

	"github.com/go-ldap/ldap/v3"
)

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func ldapOpen(w http.ResponseWriter) (*ldap.Conn, error) {
	if config.LdapTLS {
		tlsConf := &tls.Config{
			ServerName:         config.LdapServerAddr,
			InsecureSkipVerify: true,
		}
		return ldap.DialTLS("tcp", net.JoinHostPort(config.LdapServerAddr, "636"), tlsConf)
	} else {
		return ldap.DialURL("ldap://" + config.LdapServerAddr)
	}

	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	log.Printf(fmt.Sprintf("27: %v %v", err, l))
	// 	return nil
	// }

	// return l
}
