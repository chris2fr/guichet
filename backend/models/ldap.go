/*
Utilities related to LDAP
*/
package models

import (
	"log"
	"sort"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

var config = ReadConfig()
const FIELD_NAME_PROFILE_PICTURE = "profilePicture"
const FIELD_NAME_DIRECTORY_VISIBILITY = "directoryVisibility"

type SearchResult struct {
	DN             string
	Id             string
	DisplayName    string
	Email          string
	Description    string
	ProfilePicture string
}
type SearchResults struct {
	Results []SearchResult
}


func (r *SearchResults) Len() int {
	return len(r.Results)
}

func (r *SearchResults) Less(i, j int) bool {
	return r.Results[i].Id < r.Results[j].Id
}

func (r *SearchResults) Swap(i, j int) {
	r.Results[i], r.Results[j] = r.Results[j], r.Results[i]
}


func ContainsI(a string, b string) bool {
	return strings.Contains(
		strings.ToLower(a),
		strings.ToLower(b),
	)
}


// New account creation directly from interface


func OpenNewUserLdap(config *ConfigFile) (*ldap.Conn, error) {
	l, err := openLdap(config)
	if err != nil {
		log.Printf("OpenNewUserLdap 1 : %v %v", err, l)
		log.Printf("OpenNewUserLdap 1 : %v", config)
		// data.Common.ErrorMessage = err.Error()
	}
	err = l.Bind(config.NewUserDN, config.NewUserPassword)
	if err != nil {
		log.Printf("OpenNewUserLdap Bind : %v", err)
		log.Printf("OpenNewUserLdap Bind : %v", config.NewUserDN)
		log.Printf("OpenNewUserLdap Bind : %v", config.NewUserPassword)
		log.Printf("OpenNewUserLdap Bind : %v", config)
		// data.Common.ErrorMessage = err.Error()
	}
	return l, err
}

func DoDirectorySearch(ldapConn *ldap.Conn, input string) (SearchResults, error) {
	//Search values with ldap and filter

	searchRequest := ldap.NewSearchRequest(
		config.UserBaseDN,
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectclass=organizationalPerson)("+FIELD_NAME_DIRECTORY_VISIBILITY+"=on))",
		[]string{
			config.UserNameAttr,
			"displayname",
			"mail",
			"description",
			FIELD_NAME_PROFILE_PICTURE,
		},
		nil)
  //Transform the researh's result in a correct struct to send JSON
	results := []SearchResult{}
	sr, err := ldapConn.Search(searchRequest)
  if err != nil {
		return SearchResults{}, err
	}



	for _, values := range sr.Entries {
		if input == "" ||
			ContainsI(values.GetAttributeValue(config.UserNameAttr), input) ||
			ContainsI(values.GetAttributeValue("displayname"), input) ||
			ContainsI(values.GetAttributeValue("mail"), input) {
			results = append(results, SearchResult{
				DN:             values.DN,
				Id:             values.GetAttributeValue(config.UserNameAttr),
				DisplayName:    values.GetAttributeValue("displayname"),
				Email:          values.GetAttributeValue("mail"),
				Description:    values.GetAttributeValue("description"),
				ProfilePicture: values.GetAttributeValue(FIELD_NAME_PROFILE_PICTURE),
			})
		}
	}
	// sort.Sort(&results)
	search_results := SearchResults{
		Results: results,
	}
	sort.Sort(&search_results)
	return search_results, nil
}