package views

import (
	"guichet/models"
	"html/template"
	"net/http"

	// "sort"
	"strings"
	// "golang.org/x/crypto/openpgp/errors"
	// "honnef.co/go/tools/analysis/facts/nilness"
	// "github.com/go-ldap/ldap/v3"
)

const FIELD_NAME_PROFILE_PICTURE = "profilePicture"
const FIELD_NAME_DIRECTORY_VISIBILITY = "directoryVisibility"

func HandleDirectory(w http.ResponseWriter, r *http.Request) {
	templateDirectory := getTemplate("directory.html")

	login := checkLogin(w, r)
	if login == nil {
		return
	}

	templateDirectory.Execute(w, nil)
}

func HandleDirectorySearch(w http.ResponseWriter, r *http.Request) error {
	templateDirectoryResults := template.Must(template.ParseFiles(templatePath + "/directory_results.html"))

	//Get input value by user
	r.ParseMultipartForm(1024)
	input := strings.TrimSpace(strings.Join(r.Form["query"], ""))

	if r.Method != "POST" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return nil
	}

	//Log to allow the research
	login := checkLogin(w, r)
	if login == nil {
		http.Error(w, "Login required", http.StatusUnauthorized)
		return nil
	}


	search_results, err := models.DoDirectorySearch(login.conn, input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	templateDirectoryResults.Execute(w, search_results)
	return nil

}