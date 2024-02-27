/*
Guichet provides a user-management system around an LDAP Directory

Oriniated with deuxfleurs.fr and advanced by resdigita.com
*/
package main

import (

	// "crypto/tls"
	// "encoding/json"
	"flag"
	// "fmt"
	// "io/ioutil"
	"guichet/controllers"
	"guichet/models"
	"guichet/views"
	"log"

	"os"

	// "strings"

	"github.com/gorilla/sessions"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func launchPocketBase () {
	config := models.ReadConfig()
	views.GuichetSessionStore = sessions.NewCookieStore([]byte(config.SessionKey))
	
	// fmt.Println(string(session_key))
	_, err := controllers.MakeGVRouter()
	if err != nil {
		log.Fatal("Cannot start http server: ", err)
	}
}

func launchGuichet () {
	app := pocketbase.New()

	// serves static files from the provided public dir (if exists)
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
			e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS("./pb_public"), false))
			return nil
	})

	if err := app.Start(); err != nil {
			log.Fatal(err)
	}
}

func main() {

	flag.Parse()

	go launchPocketBase()
	launchGuichet()

	// session_key := make([]byte, 32)
	// n, err := rand.Read(session_key)
	// if err != nil || n != 32 {
	// 	log.Fatal(err)
	// }
	// views.GuichetSessionStore = sessions.NewCookieStore(session_key)
	

}
