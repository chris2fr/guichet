package models

import (
	// "context"
	// "fmt"
	// "net/http"
	// "os"

	httptransport "github.com/go-openapi/runtime/client"
	// api "goauthentik.io/api/v3"
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

	// authConfig := api.NewConfiguration()
	// authConfig.Debug = true
	// authConfig.Scheme = "http"
	// // authConfig.Host = "auth.lesgrandsvoisins.com"
	// authConfig.Host = "10.245.101.35:9000"
	// authConfig.HTTPClient = &http.Client{}
	// // authConfig.HTTPClient = &http.Client{
	// // 	Transport: GetTLSTransport(true),
	// // }

	// authConfig.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", config.AuthentikAPIBearerToken)) 
	
  // apiClient := api.NewAPIClient(authConfig)
	// patchedLDAPSourceRequest := api.NewPatchedLDAPSourceRequest()
	// patchedLDAPSourceRequest.SetName("Des Grands Voisins")
	// patchedLDAPSourceRequest.SetSlug("des-grands-voisins")
	// patchedLDAPSourceRequest.SetEnabled(true)
	// patchedLDAPSourceRequest.SetPolicyEngineMode("all")
	// patchedLDAPSourceRequest.SetSyncUsers(true)

	// resp, r, err := apiClient.SourcesApi.SourcesLdapPartialUpdate(context.Background(),"des-grands-voisins").PatchedLDAPSourceRequest(*patchedLDAPSourceRequest).Execute()

	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error when calling `SourcesLdapPartialUpdate``: %v\n", err)
	// 	fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	// 	// return err
	// }
	// // response from `AdminAppsList`: []App
	// fmt.Fprintf(os.Stdout, "Response from `SourcesLdapPartialUpdate`: %v\n", resp)

	return nil
}