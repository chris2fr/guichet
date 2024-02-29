/*
Centralises the models used in this application
*/

package models

import (
	// "crypto/tls"
	// "log"
	// "net"

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






// func openLdap(config *ConfigFile) (*ldap.Conn, error) {
// 	var ldapConn *ldap.Conn
// 	var err error
// 	if config.LdapTLS {
// 		tlsConf := &tls.Config{
// 			ServerName:         config.LdapServerAddr,
// 			InsecureSkipVerify: true,
// 		}
// 		ldapConn, err = ldap.DialTLS("tcp", net.JoinHostPort(config.LdapServerAddr, "636"), tlsConf)
// 	} else {
// 		ldapConn, err = ldap.DialURL("ldap://" + config.LdapServerAddr)
// 	}
// 	if err != nil {
// 		log.Printf("openLDAP %v", err)
// 		log.Printf("openLDAP %v", config.LdapServerAddr)
// 	}
// 	return ldapConn, err

// 	// l, err := ldap.DialURL(config.LdapServerAddr)
// 	// if err != nil {
// 	// 	log.Printf(fmt.Sprint("Erreur connect LDAP %v", err))
// 	// 	log.Printf(fmt.Sprint("Erreur connect LDAP %v", config.LdapServerAddr))
// 	// 	return nil
// 	// } else {
// 	// 	return l
// 	// }
// }
