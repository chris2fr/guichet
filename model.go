/*
Centralises the models used in this application
*/

package main

import (
	"github.com/go-ldap/ldap/v3"
)

/*
Represents a user
*/
type User struct {
	DN           string
	CN           string
	GivenName    string
	DisplayName  string
	Mail         string
	SN           string
	UID          string
	Description  string
	Password     string
	OtherMailbox string
	CanAdmin     bool
	CanInvite    bool
	UserEntry    *ldap.Entry
	SeeAlso      string
}

type EntryList []*ldap.Entry
