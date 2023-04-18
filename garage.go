package main

import (
	"net/http"
    "context"
    "fmt"
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


func createKey(name string) error {
    client, ctx := gadmin()

    kr := garage.AddKeyRequest{Name: &name}
    resp, _, err := client.KeyApi.AddKey(ctx).AddKeyRequest(kr).Execute()
    if err != nil {
        fmt.Printf("%+v\n", err)
        return err
    }
    fmt.Printf("%+v\n", resp)
    return nil
}

func handleGarageKey(w http.ResponseWriter, r *http.Request) {
    createKey("toto")
    tKey := getTemplate("garage_key.html")
	tKey.Execute(w, nil)
}

func handleGarageWebsiteList(w http.ResponseWriter, r *http.Request) {
    tWebsiteList := getTemplate("garage_website_list.html")
	tWebsiteList.Execute(w, nil)
}

func handleGarageWebsiteNew(w http.ResponseWriter, r *http.Request) {
    tWebsiteNew := getTemplate("garage_website_new.html")
	tWebsiteNew.Execute(w, nil)
}

func handleGarageWebsiteInspect(w http.ResponseWriter, r *http.Request) {
    tWebsiteInspect := getTemplate("garage_website_inspect.html")
	tWebsiteInspect.Execute(w, nil)
}
