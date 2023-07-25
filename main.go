/*
Guichet provides a user-management system around an LDAP Directory

Oriniated with deuxfleurs.fr and advanced by resdigita.com
*/
package main

import (
	"crypto/rand"
	// "crypto/tls"

	// "encoding/json"
	"flag"
	// "fmt"
	// "io/ioutil"
	"log"

	// "os"
	// "strings"

	"github.com/gorilla/sessions"
)

const SESSION_NAME = "guichet_session"

var store sessions.Store = nil

func main() {

	flag.Parse()

	config_file := readConfig()
	config = &config_file

	session_key := make([]byte, 32)
	n, err := rand.Read(session_key)
	if err != nil || n != 32 {
		log.Fatal(err)
	}
	store = sessions.NewCookieStore(session_key)
	_, err = makeGVRouter()
	if err != nil {
		log.Fatal("Cannot start http server: ", err)
	}
}
