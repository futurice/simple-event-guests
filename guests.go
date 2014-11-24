package mypkg

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"errors"
	"net/http"
	"strconv"
)

type guestT struct {
	Name, Email, HostEmail string
}

func init() {
	http.Handle(addGuestURL, appHandler(addGuestHandler))
}

func addGuestHandler(w http.ResponseWriter, r *http.Request) *appError {
	c := appengine.NewContext(r)
	if user.Current(c) == nil {
		msg := "Forbidden"
		return &appError{errors.New(msg), msg, http.StatusForbidden}
	}

	switch r.Method {
	case "GET":
		return addGuestGETHandler(w, r)
	case "POST":
		return addGuestPOSTHandler(w, r)
	default:
		msg := "Method not allowed ‘" + r.Method + "’"
		return &appError{errors.New(msg), msg,
			http.StatusMethodNotAllowed}
	}
}

func addGuestGETHandler(w http.ResponseWriter, r *http.Request) *appError {
	return execNavTempl(r, w, "add-guest.html", nil)
}

func addGuestPOSTHandler(w http.ResponseWriter, r *http.Request) *appError {
	name, email, hostEmail := r.FormValue("name"), r.FormValue("email"),
		r.FormValue("host_email")
	if len(name)*len(email)*len(hostEmail) == 0 {
		msg := "Name, Email and HostEmail must not be empty"
		return &appError{errors.New(msg), msg, http.StatusBadRequest}
	}

	evIdStr := r.FormValue("event_id")
	evIdInt, err := strconv.ParseInt(evIdStr, 10, 64)
	if err != nil {
		return &appError{err, "Invalid ID", http.StatusBadRequest}
	}

	c := appengine.NewContext(r)
	event := &eventT{}
	key := datastore.NewKey(c, eventKind, "", evIdInt, nil)
	err = datastore.RunInTransaction(c, func(c appengine.Context) error {
		if err := datastore.Get(c, key, event); err != nil {
			return err
		}

		event.Guests = append(event.Guests, guestT{
			Name:      name,
			Email:     email,
			HostEmail: hostEmail,
		})

		_, err := datastore.Put(c, key, event)
		return err
	}, nil)
	if err == datastore.ErrNoSuchEntity {
		return &appError{err, "Event not found", http.StatusNotFound}
	}
	if err != nil {
		return &appError{err, "Error getting event",
			http.StatusInternalServerError}
	}

	http.Redirect(w, r, eventDetailURL+"?id="+evIdStr, http.StatusFound)
	return nil
}
