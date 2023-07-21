/*
http-utils provide utility functions that interact with http
*/

package main

import (
	"crypto/tls"
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
	l, err := ldap.DialURL(config.LdapServerAddr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	if config.LdapTLS {
		err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil
		}
	}

	return l
}
