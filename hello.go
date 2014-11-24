package mypkg

import (
	"net/http"
)

func init() {
	http.Handle("/", appHandler(homepageHandler))
}

func homepageHandler(w http.ResponseWriter, r *http.Request) *appError {
	return execNavTempl(r, w, "homepage.html", map[string]interface{}{})
}
