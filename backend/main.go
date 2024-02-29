/*
Guichet provides a user-management system around an LDAP Directory

Oriniated with deuxfleurs.fr and advanced by resdigita.com
*/
package main

import (
	"flag"
	"path/filepath"
	"strings"

	"guichet/controllers"
	"guichet/models"
	"guichet/views"
	"log"

	"os"

	"github.com/gorilla/sessions"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func defaultPublicDir() string {
	if strings.HasPrefix(os.Args[0], os.TempDir()) {
		// most likely ran with go run
		return "./pb_public"
	}

	return filepath.Join(os.Args[0], "./pb_public")
}

func launchPocketBase () {
	config := models.ReadConfig()
	views.GuichetSessionStore = sessions.NewCookieStore([]byte(config.SessionKey))
	
	// fmt.Println(string(session_key))
	_, err := controllers.MakeGVRouter()
	if err != nil {
		log.Fatal("Cannot start http server: ", err)
	}
}

func main() {
	var publicDirFlag string

	flag.Parse()

	app := pocketbase.New()

	// add "--publicDir" option flag
	app.RootCmd.PersistentFlags().StringVar(
		&publicDirFlag,
		"publicDir",
		defaultPublicDir(),
		"the directory to serve static files",
	)

	// serves static files from the provided public dir (if exists)
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
			e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS(publicDirFlag), false))
			return nil
	})

	// Launch pocketbase router n background
	go launchPocketBase()

	// Launch legacy Guichet router
	if err := app.Start(); err != nil {
			log.Fatal(err)
	}

	// session_key := make([]byte, 32)
	// n, err := rand.Read(session_key)
	// if err != nil || n != 32 {
	// 	log.Fatal(err)
	// }
	// views.GuichetSessionStore = sessions.NewCookieStore(session_key)
	

}
