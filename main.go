/*
Guichet provides a user-management system around an LDAP Directory

Oriniated with deuxfleurs.fr and advanced by resdigita.com
*/
package main

import (
	"log"
	"net/http"
	"os"

	// "crypto/tls"
	// "encoding/json"
	// "fmt"
	// "io/ioutil"
	// "os"
	// "strings"

	"guichet/views"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	// "html/template"
	// "io"
)

var templatePath = "./templates"
var loggedIn = false


func main() {

	// var n int
	var err error
	// var html string
	app := pocketbase.New()
	// renderer := &EchoTemplateRenderer{
	// 	templates: template.Must(template.New("layout.html").Funcs(template.FuncMap{
	// 		"contains": strings.Contains,
	// 	}).ParseFiles(
	// 		templatePath+"/layout.html",
	// 		templatePath+"/"+"home.html",
	// 	)),
	// }

  // serves static files from the provided public dir (if exists)
  app.OnBeforeServe().Add(func(e *core.ServeEvent) error {




		e.Router.GET("/", views.EchoHome)
		e.Router.POST("/", views.EchoHome)
		e.Router.POST("/login", func(c echo.Context) error {
			data := struct {
					Identity    string `json:"identity" form:"identity"`
					Password string `json:"password" form:"password"`
			}{}
			if err := c.Bind(&data); err != nil {
					return apis.NewBadRequestError("Failed to read request data", err)
			}

			record, err := app.Dao().FindFirstRecordByData("users", "username", data.Identity)
			if err != nil {
				record, err = app.Dao().FindFirstRecordByData("users", "email", data.Identity)
			}
			if err != nil || !record.ValidatePassword(data.Password) {
					// return generic 400 error to prevent phones enumeration
					return apis.NewBadRequestError("Invalid credentials", err)
			}

			

			// session, err := views.GuichetSessionStore.Get(c.Request(), views.SESSION_NAME)

			// session.Values["pocketbase_token"] = record.TokenKey()

		  // apis.RecordAuthResponse(app, c, record, nil)
			return c.Redirect(http.StatusTemporaryRedirect, "/")
	},)
		e.Router.GET("/login", views.EchoLoginGet)
		// e.Router.POST("/login", views.EchoLoginPost)

		e.Router.GET("/favicon.ico", apis.StaticDirectoryHandler(os.DirFS("./static/favicon.ico"), false))

    e.Router.GET("/static/*", apis.StaticDirectoryHandler(os.DirFS("./static"), false))

    e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS("./pb_public"), false))
    return nil
  })

  if err = app.Start(); err != nil {
      log.Fatal(err)
  }

	// flag.Parse()
	// session_key := make([]byte, 32)
	// n, err = rand.Read(session_key)
	// if err != nil || n != 32 {
	// 	log.Fatal(err)
	// }
	// views.GuichetSessionStore = sessions.NewCookieStore(session_key)
	
	// _, err = controllers.MakeGVRouter()
	// if err != nil {
	// 	log.Fatal("Cannot start http server: ", err)
	// }
		



}

