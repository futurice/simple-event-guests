package mypkg

import (
	"fmt"
	"net/http"
)

func init() {
	http.Handle("/", appHandler(homepageHandler))
}

func homepageHandler(w http.ResponseWriter, r *http.Request) *appError {
	fmt.Fprint(w, "Hello world!")
	return nil
}
