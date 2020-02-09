package main

import (
	"os"
	"log"
	"net/http"
	"fmt"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

func handleHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, world!")
}

func main() {
	http.HandleFunc("/", handleHome)

	bind_addr := os.Getenv("HTTP_BIND_ADDR")
	if bind_addr == "" {
		bind_addr = ":9991"
	}

	err := http.ListenAndServe(bind_addr, nil)
	if err != nil {
		log.Fatal("Cannot start http server: ", err)
	}
}
