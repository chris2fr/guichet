package models

import (
	"context"
	"fmt"
	"net/http"
	"os"

	httptransport "github.com/go-openapi/runtime/client"
	api "goauthentik.io/api/v3"
	// openapi"github.com/chris2fr/go-authentik-oapi-desgv"

)


func GetTLSTransport(insecure bool) http.RoundTripper {
	tlsTransport, err := httptransport.TLSTransport(httptransport.TLSClientOptions{
		// InsecureSkipVerify: insecure,
	})
	if err != nil {
		panic(err)
	}
	return tlsTransport
}


func SyncAuthentikLDAP () error {

	// slug := "des-grands-voisins"
	// patchedLDAPSourceRequest := *openapi.NewPatchedLDAPSourceRequest() // PatchedLDAPSourceRequest |  (optional)
	// configuration := openapi.NewConfiguration()
  // apiClient := openapi.NewAPIClient(configuration)
  // resp, r, err := apiClient.SourcesApi.SourcesLdapPartialUpdate(context.Background(), slug).PatchedLDAPSourceRequest(patchedLDAPSourceRequest).Execute()
  // if err != nil {
  //     fmt.Fprintf(os.Stderr, "Error when calling `SourcesApi.SourcesLdapPartialUpdate``: %v\n", err)
  //     fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
  // }
  // // response from `SourcesLdapPartialUpdate`: LDAPSource
  // fmt.Fprintf(os.Stdout, "Response from `SourcesApi.SourcesLdapPartialUpdate`: %v\n", resp)



// 	// os.Setenv("HTTP_PROXY", "https://auth.lesgrandsvoisins.com")

	authConfig := api.NewConfiguration()
	authConfig.Debug = true
	authConfig.Scheme = "http"
	// authConfig.Host = "auth.lesgrandsvoisins.com"
	authConfig.Host = "10.245.101.35:9000"
	authConfig.HTTPClient = &http.Client{}
	// authConfig.HTTPClient = &http.Client{
	// 	Transport: GetTLSTransport(true),
	// }

	authConfig.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", config.AuthentikAPIBearerToken)) 
	
  apiClient := api.NewAPIClient(authConfig)
	patchedLDAPSourceRequest := api.NewPatchedLDAPSourceRequest()
	patchedLDAPSourceRequest.SetName("Des Grands Voisins")
	patchedLDAPSourceRequest.SetSlug("des-grands-voisins")
	patchedLDAPSourceRequest.SetEnabled(true)
	patchedLDAPSourceRequest.SetPolicyEngineMode("all")
	patchedLDAPSourceRequest.SetSyncUsers(true)

	resp, r, err := apiClient.SourcesApi.SourcesLdapPartialUpdate(context.Background(),"des-grands-voisins").PatchedLDAPSourceRequest(*patchedLDAPSourceRequest).Execute()

// 	// return nil
// 	// resp, r, err := apiClient.AdminApi.AdminAppsList(context.Background()).Execute()
// 	// ctx := context.Background()

// 	// appreq	:= api.NewLDAPSyncStatus(true, )
// 	// apiClient.SourcesApi.SourcesLdapSyncStatusRetrieveExecute()

// 	ldapsrcreq := api.LDAPSourceRequest{
// 		Slug: "des-grands-voisins",
// 		Name: "Des Grands Voisins",
// 		BaseDn: "dc=resdigita,dc=org",
// 		ServerUri: "ldap://mail.lesgrandsvoisins.com",
// 	}
// 	ldapsrcreq.SetSyncUsers(true)
// 	ldapsrcsupreq := api.ApiSourcesLdapUpdateRequest{
// 		ApiService: apiClient.SourcesApi,
// 	}
// 	// api.LDAPSourceRequest{
// 	// 	Slug: "des-grands-voisins",
// 	// }
// 	req := api.ApiSourcesLdapUpdateRequest.LDAPSourceRequest(ldapsrcsupreq, ldapsrcreq)

// 	resp, r, err := req.Execute()

// 	// Request(ldapsrcreq)


	
 
// 	// resp, r, err := apiClient.SourcesApi.SourcesLdapUpdateExecute(req)


// 	// resp, r, err := apiClient.SourcesApi.SourcesLdapUpdate(context.WithValue(context.Background(),))
	
// 	apiClient.SourcesApi.SourcesLdapUpdate(context.Background(),"des-grands-voisins").Execute()
// 	// .AdminApi.AdminAppsList(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SourcesLdapPartialUpdate``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		// return err
	}
	// response from `AdminAppsList`: []App
	fmt.Fprintf(os.Stdout, "Response from `SourcesLdapPartialUpdate`: %v\n", resp)

// 	// PATCH /sources/ldap/des-grands-voisins
// // 	curl -X PATCH "https://auth.lesgrandsvoisins.com/api/v3/sources/ldap/des-grands-voisins/" \
// //  -H "accept: application/json"\
// //  -H "content-type: application/json" \
// //  -H "content-type: application/json" \
// //  -d '{}' 

// 	// // Replace these with your Authentik server details
// 	// authURL := "https://auth.lesgrandsvoisins.com/api/v3"
// 	// username := "apiuser"
// 	// password := "e5MZCP7mH2h5JTqA"

// 	// // Create a new Authentik client
// 	// client := authentik.NewClient(authURL, nil)

// 	// // Login to the Authentik server
// 	// err := client.Login(username, password)
// 	// if err != nil {
// 	// 	log.Fatal("Error logging in:", err)
// 	// 	return err
// 	// }

// 	// // Trigger synchronization for the LDAP user source with slug "des-grands-voisins"
// 	// slug := "des-grands-voisins"
// 	// err = client.Providers.LDAP.Update(slug)
// 	// if err != nil {
// 	// 	log.Fatal("Error triggering synchronization:", err)
// 	// 	return err
// 	// }

// 	// fmt.Println("Synchronization triggered successfully for slug:", slug)
	return nil
}