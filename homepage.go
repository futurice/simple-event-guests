package mypkg

import (
	"fmt"
	"net/http"
)

func init() {
	http.Handle("/", appHandler(rootHandler))
}

func rootHandler(w http.ResponseWriter, r *http.Request) *appError {
	if r.RequestURI != "/" {
		return &appError{
			fmt.Errorf("URL not found: ‘%v’", r.RequestURI),
			"Not Found", http.StatusNotFound}
	}
	return execNavTempl(r, w, "homepage.html", map[string]interface{}{})
}
