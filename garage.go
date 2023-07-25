package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	garage "git.deuxfleurs.fr/garage-sdk/garage-admin-sdk-golang"
	"github.com/go-ldap/ldap/v3"
	"github.com/gorilla/mux"
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

func grgCreateWebsite(gkey, bucket string) (*garage.BucketInfo, error) {
	client, ctx := gadmin()

	br := garage.NewCreateBucketRequest()
	br.SetGlobalAlias(bucket)

	// Create Bucket
	binfo, _, err := client.BucketApi.CreateBucket(ctx).CreateBucketRequest(*br).Execute()
	if err != nil {
		fmt.Printf("%+v\n", err)
		return nil, err
	}

	// Allow user's key
	ar := garage.AllowBucketKeyRequest{
		BucketId:    *binfo.Id,
		AccessKeyId: gkey,
		Permissions: *garage.NewAllowBucketKeyRequestPermissions(true, true, true),
	}
	binfo, _, err = client.BucketApi.AllowBucketKey(ctx).AllowBucketKeyRequest(ar).Execute()
	if err != nil {
		fmt.Printf("%+v\n", err)
		return nil, err
	}

	// Expose website and set quota
	wr := garage.NewUpdateBucketRequestWebsiteAccess()
	wr.SetEnabled(true)
	wr.SetIndexDocument("index.html")
	wr.SetErrorDocument("error.html")

	qr := garage.NewUpdateBucketRequestQuotas()
	qr.SetMaxSize(1024 * 1024 * 50) // 50MB
	qr.SetMaxObjects(10000)         //10k objects

	ur := garage.NewUpdateBucketRequest()
	ur.SetWebsiteAccess(*wr)
	ur.SetQuotas(*qr)

	binfo, _, err = client.BucketApi.UpdateBucket(ctx, *binfo.Id).UpdateBucketRequest(*ur).Execute()
	if err != nil {
		fmt.Printf("%+v\n", err)
		return nil, err
	}

	// Return updated binfo
	return binfo, nil
}

func grgGetBucket(bid string) (*garage.BucketInfo, error) {
	client, ctx := gadmin()

	resp, _, err := client.BucketApi.GetBucketInfo(ctx, bid).Execute()
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
	Status *LoginStatus
	Key    *garage.KeyInfo
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
	Status *LoginStatus
	Key    *garage.KeyInfo
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
	_, s3key, err := checkLoginAndS3(w, r)
	if err != nil {
		log.Println(err)
		return
	}

	tWebsiteNew := getTemplate("garage_website_new.html")
	if r.Method == "POST" {
		r.ParseForm()
		log.Println(r.Form)

		bucket := strings.Join(r.Form["bucket"], "")
		if bucket == "" {
			bucket = strings.Join(r.Form["bucket2"], "")
		}
		if bucket == "" {
			log.Println("Form empty")
			// @FIXME we need to return the error to the user
			tWebsiteNew.Execute(w, nil)
			return
		}

		binfo, err := grgCreateWebsite(*s3key.AccessKeyId, bucket)
		if err != nil {
			log.Println(err)
			// @FIXME we need to return the error to the user
			tWebsiteNew.Execute(w, nil)
			return
		}

		http.Redirect(w, r, "/garage/website/b/"+*binfo.Id, http.StatusFound)
		return
	}

	tWebsiteNew.Execute(w, nil)
}

type webInspectView struct {
	Status      *LoginStatus
	Key         *garage.KeyInfo
	Bucket      *garage.BucketInfo
	IndexDoc    string
	ErrorDoc    string
	MaxObjects  int64
	MaxSize     int64
	UsedSizePct float64
}

func handleGarageWebsiteInspect(w http.ResponseWriter, r *http.Request) {
	login, s3key, err := checkLoginAndS3(w, r)
	if err != nil {
		log.Println(err)
		return
	}

	bucketId := mux.Vars(r)["bucket"]
	binfo, err := grgGetBucket(bucketId)
	if err != nil {
		log.Println(err)
		return
	}

	wc := binfo.GetWebsiteConfig()
	q := binfo.GetQuotas()

	view := webInspectView{
		Status:     login,
		Key:        s3key,
		Bucket:     binfo,
		IndexDoc:   (&wc).GetIndexDocument(),
		ErrorDoc:   (&wc).GetErrorDocument(),
		MaxObjects: (&q).GetMaxObjects(),
		MaxSize:    (&q).GetMaxSize(),
	}

	tWebsiteInspect := getTemplate("garage_website_inspect.html")
	tWebsiteInspect.Execute(w, &view)
}
