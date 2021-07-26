package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/gorilla/mux"
)

func handleDirectory(w http.ResponseWriter, r *http.Request) {
	templateDirectory := template.Must(template.ParseFiles("templates/layout.html", "templates/directory.html"))

	login := checkLogin(w, r)
	if login == nil {
		return
	}

	templateDirectory.Execute(w, nil)
}

type SearchResult struct {
	Identifiant string `json:"identifiant"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Description string `json:"description"`
	DN          string `json:"dn"`
}

type Results struct {
	Search    []SearchResult `json:"search"`
	MessageID uint32         `json:"id"`
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	//Get input value by user
	input := mux.Vars(r)["input"]
	//Log to allow the research
	login := checkLogin(w, r)
	if login == nil {
		return
	}

	session, err := store.Get(r, SESSION_NAME)
	if err != nil {
		return
	}

	//Search values with ldap and filter
	searchRequest := ldap.NewSearchRequest(
		config.UserBaseDN,
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectclass=organizationalPerson)(visibility=on))",
		[]string{config.UserNameAttr, "displayname", "mail", "description"},
		nil)

	sr, err := login.conn.Search(searchRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Transform the researh's result in a correct struct to send HSON
	var result Results
	for _, values := range sr.Entries {

		if strings.Contains(values.GetAttributeValue("cn"), input) {
			result = Results{
				Search: append(result.Search, SearchResult{
					Identifiant: values.GetAttributeValue("cn"),
					Name:        values.GetAttributeValue("displayname"),
					Email:       values.GetAttributeValue("email"),
					Description: values.GetAttributeValue("description"),
					DN:          values.DN,
				}),
			}
		}

	}

	//Convert interface to uint32 with Type Assertions and not a simple convert for messageID
	if val_Raw, ok_raw := session.Values["MessageID"]; ok_raw {
		if val, ok := val_Raw.(uint32); ok {
			val += 1
			session.Values["MessageID"] = val
			result.MessageID = val
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	//Send JSON through xhttp
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
