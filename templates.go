package mypkg

import (
	"appengine"
	"appengine/user"
	"errors"
	"html/template"
	"io"
	"net/http"
)

var appTempl *template.Template = template.Must(template.ParseGlob("templates/*"))

// Executes the named template from appTempl on the Writer
func execTempl(w io.Writer, name string, data interface{}) *appError {
	if err := appTempl.ExecuteTemplate(w, name, data); err != nil {
		return &appError{err, "Error executing template",
			http.StatusInternalServerError}
	}
	return nil
}

// Execute a ‘navigation template’.
// Adds "navData" to the caller's map, modifying it, then calls execTempl.
// Returns an error if the map already contains the "navData" key.
func execNavTempl(r *http.Request, w io.Writer, name string,
	data map[string]interface{}) *appError {
	if _, ok := data["navData"]; ok {
		msg := "navData already present in template map"
		return &appError{errors.New(msg), msg,
			http.StatusInternalServerError}
	}

	c := appengine.NewContext(r)
	loginURL, err := user.LoginURL(c, r.RequestURI)
	if err != nil {
		return &appError{err, "Error getting login URL",
			http.StatusInternalServerError}
	}
	logoutURL, err := user.LogoutURL(c, r.RequestURI)
	if err != nil {
		return &appError{err, "Error getting logout URL",
			http.StatusInternalServerError}
	}

	data["navData"] = map[string]interface{}{
		"navURLs":   navURLs,
		"loginURL":  loginURL,
		"logoutURL": logoutURL,
		"user":      user.Current(c),
	}
	return execTempl(w, name, data)
}
