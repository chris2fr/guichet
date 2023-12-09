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
	"guichet/controllers"
	"guichet/views"
	"log"

	// "os"

	// "strings"

	"github.com/gorilla/sessions"
)

func main() {

	flag.Parse()

	session_key := make([]byte, 32)
	n, err := rand.Read(session_key)
	if err != nil || n != 32 {
		log.Fatal(err)
	}
	views.GuichetSessionStore = sessions.NewCookieStore(session_key)
	_, err = controllers.MakeGVRouter()
	if err != nil {
		log.Fatal("Cannot start http server: ", err)
	}
}
