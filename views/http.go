/*
http-utils provide utility functions that interact with http
*/

package views

import (
	"net/http"
)

func LogRequest(Handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		Handler.ServeHTTP(w, r)
	})
}


