package main

import (
    "errors"
    "log"
	"net/http"
    "context"
    "fmt"
	"github.com/go-ldap/ldap/v3"
     garage "git.deuxfleurs.fr/garage-sdk/garage-admin-sdk-golang"
)

func gadmin() (*garage.APIClient, context.Context) {
    // Set Host and other parameters
    configuration := garage.NewConfiguration()
    configuration.Host = config.S3AdminEndpoint

    // We can now generate a client
    client := garage.NewAPIClient(configuration)

    // Authentication is handled through the context pattern
    ctx := context.WithValue(context.Background(), garage.ContextAccessToken, config.S3AdminToken)
    return client, ctx
}


func grgCreateKey(name string) (*garage.KeyInfo, error) {
    client, ctx := gadmin()

    kr := garage.AddKeyRequest{Name: &name}
    resp, _, err := client.KeyApi.AddKey(ctx).AddKeyRequest(kr).Execute()
    if err != nil {
        fmt.Printf("%+v\n", err)
        return nil, err
    }
    return resp, nil
}

func grgGetKey(accessKey string) (*garage.KeyInfo, error) {
    client, ctx := gadmin()

    resp, _, err := client.KeyApi.GetKey(ctx, accessKey).Execute()
    if err != nil {
        fmt.Printf("%+v\n", err)
        return nil, err
    }
    return resp, nil
}


func checkLoginAndS3(w http.ResponseWriter, r *http.Request) (*LoginStatus, *garage.KeyInfo, error) {
	login := checkLogin(w, r)
	if login == nil {
		return nil, nil, errors.New("LDAP login failed")
	}

	keyID := login.UserEntry.GetAttributeValue("garage_s3_access_key")
    if keyID == "" {
        keyPair, err := grgCreateKey(login.Info.Username)
        if err != nil {
            return login, nil, err
        }
		modify_request := ldap.NewModifyRequest(login.Info.DN, nil)
		modify_request.Replace("garage_s3_access_key", []string{*keyPair.AccessKeyId})
        // @FIXME compatibility feature for bagage (SFTP+webdav)
        // you can remove it once bagage will be updated to fetch the key from garage directly
        // or when bottin will be able to dynamically fetch it.
		modify_request.Replace("garage_s3_secret_key", []string{*keyPair.SecretAccessKey})
        err = login.conn.Modify(modify_request)
        return login, keyPair, err
    }
    // Note: we could simply return the login info, but LX asked we do not 
    // store the secrets in LDAP in the future.
    keyPair, err := grgGetKey(keyID)
    return login, keyPair, err
}

type keyView struct {
	Status         *LoginStatus
    Key            *garage.KeyInfo
}

func handleGarageKey(w http.ResponseWriter, r *http.Request) {
    login, s3key, err := checkLoginAndS3(w, r)
    if err != nil {
        log.Println(err)
        return
    }
    view := keyView{Status: login, Key: s3key}

    tKey := getTemplate("garage_key.html")
	tKey.Execute(w, &view)
}

type webListView struct {
	Status         *LoginStatus
    Key            *garage.KeyInfo
}
func handleGarageWebsiteList(w http.ResponseWriter, r *http.Request) {
    login, s3key, err := checkLoginAndS3(w, r)    
    if err != nil {
        log.Println(err)
        return
    }
    view := webListView{Status: login, Key: s3key}

    tWebsiteList := getTemplate("garage_website_list.html")
	tWebsiteList.Execute(w, &view)
}

func handleGarageWebsiteNew(w http.ResponseWriter, r *http.Request) {
    tWebsiteNew := getTemplate("garage_website_new.html")
	tWebsiteNew.Execute(w, nil)
}

func handleGarageWebsiteInspect(w http.ResponseWriter, r *http.Request) {
    login, s3key, err := checkLoginAndS3(w, r)
    if err != nil {
        log.Println(err)
        return
    }
    log.Println(login, s3key)

    tWebsiteInspect := getTemplate("garage_website_inspect.html")
	tWebsiteInspect.Execute(w, nil)
}
