package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/gorilla/mux"
)

const FIELD_NAME_PROFILE_PICTURE = "profilePicture"
const FIELD_NAME_DIRECTORY_VISIBILITY = "directoryVisibility"

func handleDirectory(w http.ResponseWriter, r *http.Request) {
	templateDirectory := template.Must(template.ParseFiles("templates/layout.html", "templates/directory.html"))

	login := checkLogin(w, r)
	if login == nil {
		return
	}

	templateDirectory.Execute(w, nil)
}

type SearchResult struct {
	Id          string `json:"id"`
	Displayname string `json:"displayname"`
	Email       string `json:"email"`
	Description string `json:"description"`
	DN          string `json:"dn"`
}

type Results struct {
	Search    []SearchResult `json:"search"`
	MessageID uint32         `json:"id"`
}

type UniqueID struct {
	Id int `json:"id"`
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	//Get input value by user
	input := mux.Vars(r)["input"]
	//Log to allow the research
	login := checkLogin(w, r)
	if login == nil {
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
	var result Results
	for _, values := range sr.Entries {

		if strings.Contains(values.GetAttributeValue(config.UserNameAttr), input) || strings.Contains(values.GetAttributeValue("displayname"), input) ||
			(values.GetAttributeValue("email") != "" && strings.Contains(values.GetAttributeValue("email"), input)) {
			result = Results{
				Search: append(result.Search, SearchResult{
					Id:          values.GetAttributeValue(config.UserNameAttr),
					Displayname: values.GetAttributeValue("displayname"),
					Email:       values.GetAttributeValue("email"),
					Description: values.GetAttributeValue("description"),
					DN:          values.DN,
				}),
			}
		}

	}
	if result.Search == nil {
		result = Results{
			Search: append(result.Search, SearchResult{}),
		}
	}

	var id UniqueID
	//Decode JSON body
	err = json.NewDecoder(r.Body).Decode(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	result.MessageID = uint32(id.Id)

	//Send JSON through xhttp
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
