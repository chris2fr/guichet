package models

import (
	"context"
	"fmt"
	"net/http"
	"os"

	httptransport "github.com/go-openapi/runtime/client"
	api "goauthentik.io/api/v3"
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

	authConfig := api.NewConfiguration()
	authConfig.Debug = true
	authConfig.Scheme = "https"
	authConfig.Host = "auth.lesgrandsvoisins.com"
	authConfig.HTTPClient = &http.Client{
		Transport: GetTLSTransport(true),
	}

	authConfig.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", config.AuthentikAPIBearerToken)) // <- how to obtain it
        apiClient := api.NewAPIClient(authConfig)

	// return nil
	// resp, r, err := apiClient.AdminApi.AdminAppsList(context.Background()).Execute()
	// ctx := context.Background()
	ldapsrcreq := api.LDAPSourceRequest{
		Slug: "des-grands-voisins",
		Name: "Des Grands Voisins",
		BaseDn: "dc=resdigita,dc=org",
		ServerUri: "ldap://mail.lesgrandsvoisins.com",
	}
	ldapsrcreq.SetSyncUsers(true)
	ldapsrcsupreq := api.ApiSourcesLdapUpdateRequest{
		ApiService: apiClient.SourcesApi,
	}
	// api.LDAPSourceRequest{
	// 	Slug: "des-grands-voisins",
	// }
	req := api.ApiSourcesLdapUpdateRequest.LDAPSourceRequest(ldapsrcsupreq, ldapsrcreq)

	resp, r, err := req.Execute()

	// Request(ldapsrcreq)


	
 
	// resp, r, err := apiClient.SourcesApi.SourcesLdapUpdateExecute(req)


	// resp, r, err := apiClient.SourcesApi.SourcesLdapUpdate(context.WithValue(context.Background(),))
	
	apiClient.SourcesApi.SourcesLdapUpdate(context.Background(),"des-grands-voisins").Execute()
	// .AdminApi.AdminAppsList(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.AdminAppsList``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		// return err
	}
	// response from `AdminAppsList`: []App
	fmt.Fprintf(os.Stdout, "Response from `AdminApi.AdminAppsList`: %v\n", resp)

	// PATCH /sources/ldap/des-grands-voisins
// 	curl -X PATCH "https://auth.lesgrandsvoisins.com/api/v3/sources/ldap/des-grands-voisins/" \
//  -H "accept: application/json"\
//  -H "content-type: application/json" \
//  -H "content-type: application/json" \
//  -d '{}' 

	// // Replace these with your Authentik server details
	// authURL := "https://auth.lesgrandsvoisins.com/api/v3"
	// username := "apiuser"
	// password := "e5MZCP7mH2h5JTqA"

	// // Create a new Authentik client
	// client := authentik.NewClient(authURL, nil)

	// // Login to the Authentik server
	// err := client.Login(username, password)
	// if err != nil {
	// 	log.Fatal("Error logging in:", err)
	// 	return err
	// }

	// // Trigger synchronization for the LDAP user source with slug "des-grands-voisins"
	// slug := "des-grands-voisins"
	// err = client.Providers.LDAP.Update(slug)
	// if err != nil {
	// 	log.Fatal("Error triggering synchronization:", err)
	// 	return err
	// }

	// fmt.Println("Synchronization triggered successfully for slug:", slug)
	return nil
}