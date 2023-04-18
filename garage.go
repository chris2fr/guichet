package main

import (
	"net/http"
)


func handleGarageKey(w http.ResponseWriter, r *http.Request) {
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
