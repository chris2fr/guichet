package main

import (
	"html/template"
	"net/http"
	"sort"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

const FIELD_NAME_PROFILE_PICTURE = "profilePicture"
const FIELD_NAME_DIRECTORY_VISIBILITY = "directoryVisibility"

func handleDirectory(w http.ResponseWriter, r *http.Request) {
	templateDirectory := getTemplate("directory.html")

	login := checkLogin(w, r)
	if login == nil {
		return
	}

	templateDirectory.Execute(w, nil)
}

func handleDirectorySearch(w http.ResponseWriter, r *http.Request) {
	templateDirectoryResults := template.Must(template.ParseFiles(templatePath + "/directory_results.html"))

	//Get input value by user
	r.ParseMultipartForm(1024)
	input := strings.TrimSpace(strings.Join(r.Form["query"], ""))

	if r.Method != "POST" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	//Log to allow the research
	login := checkLogin(w, r)
	if login == nil {
		http.Error(w, "Login required", http.StatusUnauthorized)
		return
	}

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

	sr, err := login.conn.Search(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Transform the researh's result in a correct struct to send JSON
	results := []SearchResult{}

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

	search_results := SearchResults{
		Results: results,
	}
	sort.Sort(&search_results)

	templateDirectoryResults.Execute(w, search_results)
}

func ContainsI(a string, b string) bool {
	return strings.Contains(
		strings.ToLower(a),
		strings.ToLower(b),
	)
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
