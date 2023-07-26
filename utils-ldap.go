/*
Utilities related to LDAP
*/
package main

import (
	"github.com/go-ldap/ldap/v3"
)

func replaceIfContent(modifReq *ldap.ModifyRequest, key string, value string, previousValue string) error {
	if value != "" {
		modifReq.Replace(key, []string{value})
	} else if previousValue != "" {
		modifReq.Delete(key, []string{previousValue})
	}
	return nil
}
