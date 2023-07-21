/*
http-utils provide utility functions that interact with http
*/

package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/go-ldap/ldap/v3"
)

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func ldapOpen(w http.ResponseWriter) *ldap.Conn {
	if config.LdapTLS {
		err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil
		}
	}

	l, err := ldap.DialURL(config.LdapServerAddr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf(fmt.Sprintf("27: %v %v", err, l))
		return nil
	}
	return l
}
