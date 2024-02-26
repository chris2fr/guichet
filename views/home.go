/*
home show the home page
*/

package views

import (
	// "html/template"
	// "io"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"

	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
)




func EchoHome(c echo.Context) error {
	data := PrepareHome()
	data.Common.LoggedIn = CheckLoggedIn(c);
	return doEchoRender(c, "home.html", data)
}

func EchoHomePost(c echo.Context) error {
	data := PrepareHome()

	log.Printf(apis.ContextAuthRecordKey)

	postBody, _ := json.Marshal(map[string]string{
		"identity":  c.Request().PostFormValue("identity"),
		"password": c.Request().PostFormValue("password"),
 	})

 	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post("http://127.0.0.1:8090/api/collections/users/auth-with-password", "application/json", responseBody)
	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	 }
	 defer resp.Body.Close()
   body, err := ioutil.ReadAll(resp.Body)
   if err != nil {
      log.Fatalln(err)
   }
   sb := string(body)
   log.Printf(sb)
	 log.Printf(apis.ContextAuthRecordKey)
	data.Common.LoggedIn = CheckLoggedIn(c);
	return doEchoRender(c, "home.html", data)
}

func CheckLoggedIn (c echo.Context) bool {
	// session, _ := GuichetSessionStore.Get(c.Request(), SESSION_NAME)

	// session, _ := GuichetSessionStore.Get(c.Request(), SESSION_NAME)
	// return session.Values["pocketbase_token"] != nil

	// admin, _ := c.Get(apis.ContextAdminKey).(*models.Admin)
	// record, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)

	// alternatively, you can also read the auth state form the cached request info
	info   := apis.RequestInfo(c)
	admin  := info.Admin      // nil if not authenticated as admin
	record := info.AuthRecord // nil if not authenticated as regular auth record

	return admin != nil && record != nil
}


func PrepareHome () HomePageData {
	data := HomePageData{
		Login: NestedLoginTplData{},
		BaseDN: config.BaseDN,
		Org:    config.Org,
		Common: NestedCommonTplData{
			CanAdmin:  false,
			CanInvite: true,
			LoggedIn:  true,
		},
	}
	return data
}

// type EchoTemplateRenderer struct {
// 	templates *template.Template
// }

// func (t *EchoTemplateRenderer) EchoTemplateRender (w io.Writer, name string, data interface{}, c echo.Context) error {

// 	return t.templates.ExecuteTemplate(w, name, data)
// }

// func EchoHandleHome(c echo.Context) error {
// 	data := PrepareHome()
// 	data.Common.WebsiteURL = config.WebAddress
// 	data.Common.WebsiteName = config.Org

// 	// return t.Execute(c.Response(), LayoutTemplateData{
// 	// 	Common: data.Common,
// 	// 	Login:  data.Login,
// 	// 	Data:   data,
// 	// })
// }

func HandleHome(w http.ResponseWriter, r *http.Request) {
	templateHome := getTemplate("home.html")
  login := checkLogin(w, r)
	if login == nil {
		status, _ := HandleLogin(w, r)
		if status == nil {
			return
		}
		login = checkLogin(w, r)
	}
	data := PrepareHome()
	data.Common.CanAdmin = login.Common.CanAdmin
	data.Login.Login = login
	execTemplate(w, templateHome, data.Common, data.Login, data)
	login.conn.Close()
	// templateHome.Execute(w, data)
	

}
