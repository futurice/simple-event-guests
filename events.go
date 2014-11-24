package mypkg

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type eventT struct {
	Name   string
	Guests []guestT
}

const (
	eventKind = "event"
)

func init() {
	http.Handle(createEventURL, appHandler(createEventHandler))
	http.Handle(eventListURL, appHandler(eventListHandler))
	http.Handle(eventDetailURL, appHandler(eventDetailHandler))
	http.Handle(editEventURL, appHandler(editEventHandler))
	http.Handle(deleteEventURL, appHandler(deleteEventHandler))
}

func createEventHandler(w http.ResponseWriter, r *http.Request) *appError {
	if !user.IsAdmin(appengine.NewContext(r)) {
		msg := "Forbidden"
		return &appError{errors.New(msg), msg, http.StatusForbidden}
	}

	switch r.Method {
	case "GET":
		return createEventGETHandler(w, r)
	case "POST":
		return createEventPOSTHandler(w, r)
	default:
		msg := "Method not allowed ‘" + r.Method + "’"
		return &appError{errors.New(msg), msg,
			http.StatusMethodNotAllowed}
	}
}

func createEventGETHandler(w http.ResponseWriter, r *http.Request) *appError {
	return execNavTempl(r, w, "create-event.html", nil)
}

func createEventPOSTHandler(w http.ResponseWriter, r *http.Request) *appError {
	name := r.FormValue("name")
	if len(name) == 0 {
		msg := "New event name must not be empty"
		return &appError{errors.New(msg), msg, http.StatusBadRequest}
	}
	event := &eventT{Name: name}

	c := appengine.NewContext(r)
	key := datastore.NewIncompleteKey(c, eventKind, nil)
	if _, err := datastore.Put(c, key, event); err != nil {
		return &appError{err, "Error creating event",
			http.StatusInternalServerError}
	}
	http.Redirect(w, r, eventListURL, http.StatusFound)
	return nil
}

func eventListHandler(w http.ResponseWriter, r *http.Request) *appError {
	c := appengine.NewContext(r)
	if user.Current(c) == nil {
		msg := "Forbidden"
		return &appError{errors.New(msg), msg, http.StatusForbidden}
	}

	q := datastore.NewQuery(eventKind).Project("Name")
	events := []*eventT{}
	keys, err := q.GetAll(c, &events)
	if err != nil {
		return &appError{err, "Error reading events",
			http.StatusInternalServerError}
	}

	return execNavTempl(r, w, "event-list.html", map[string]interface{}{
		"title":  "Events",
		"events": events,
		"keys":   keys,
	})
}

func eventDetailHandler(w http.ResponseWriter, r *http.Request) *appError {
	c := appengine.NewContext(r)
	u := user.Current(c)
	if u == nil {
		msg := "Forbidden"
		return &appError{errors.New(msg), msg, http.StatusForbidden}
	}

	idStr, showAllStr := r.FormValue("id"), r.FormValue("show_all_guests")
	idInt, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &appError{err, "Invalid ID", http.StatusBadRequest}
	}
	// show all guests, or only those invited by the current user?
	showAll := false
	if len(showAllStr) > 0 {
		var err error
		if showAll, err = strconv.ParseBool(showAllStr); err != nil {
			return &appError{err, "Invalid show_all_guests value",
				http.StatusBadRequest}
		}
	}

	event := &eventT{}
	key := datastore.NewKey(c, eventKind, "", idInt, nil)
	err = datastore.RunInTransaction(c, func(c appengine.Context) error {
		return datastore.Get(c, key, event)
	}, nil)
	if err == datastore.ErrNoSuchEntity {
		return &appError{err, "Not found", http.StatusNotFound}
	}
	if err != nil {
		return &appError{err, "Error getting event",
			http.StatusInternalServerError}
	}

	totalGuests := len(event.Guests)
	if !showAll {
		myGuests, myEmail := []guestT{}, strings.ToLower(u.Email)
		for i := range event.Guests {
			if strings.ToLower(event.Guests[i].HostEmail) == myEmail {
				myGuests = append(myGuests, event.Guests[i])
			}
		}
		event.Guests = myGuests
	}

	return execNavTempl(r, w, "event-detail.html", map[string]interface{}{
		"title":         event.Name,
		"event":         event,
		"id":            idInt,
		"totalGuests":   totalGuests,
		"showAllGuests": showAll,
	})
}

func editEventHandler(w http.ResponseWriter, r *http.Request) *appError {
	if !user.IsAdmin(appengine.NewContext(r)) {
		msg := "Forbidden"
		return &appError{errors.New(msg), msg, http.StatusForbidden}
	}

	switch r.Method {
	case "GET":
		return editEventGETHandler(w, r)
	case "POST":
		return editEventPOSTHandler(w, r)
	default:
		msg := "Method not allowed ‘" + r.Method + "’"
		return &appError{errors.New(msg), msg,
			http.StatusMethodNotAllowed}
	}
}

func editEventGETHandler(w http.ResponseWriter, r *http.Request) *appError {
	idStr := r.FormValue("id")
	idInt, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &appError{err, "Invalid ID", http.StatusBadRequest}
	}

	event := &eventT{}
	c := appengine.NewContext(r)
	key := datastore.NewKey(c, eventKind, "", idInt, nil)
	err = datastore.RunInTransaction(c, func(c appengine.Context) error {
		return datastore.Get(c, key, event)
	}, nil)
	if err == datastore.ErrNoSuchEntity {
		return &appError{err, "Not found", http.StatusNotFound}
	}
	if err != nil {
		return &appError{err, "Error getting event",
			http.StatusInternalServerError}
	}

	return execNavTempl(r, w, "edit-event.html", map[string]interface{}{
		"title": "Edit event ‘" + event.Name + "’",
		"event": event,
		"id":    idInt,
	})
}

func editEventPOSTHandler(w http.ResponseWriter, r *http.Request) *appError {
	name := r.FormValue("name")
	if len(name) == 0 {
		msg := "The event name must not be empty"
		return &appError{errors.New(msg), msg, http.StatusBadRequest}
	}

	idStr := r.FormValue("id")
	idInt, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &appError{err, "Invalid ID", http.StatusBadRequest}
	}

	c := appengine.NewContext(r)
	key := datastore.NewKey(c, eventKind, "", idInt, nil)
	err = datastore.RunInTransaction(c, func(c appengine.Context) error {
		event := &eventT{}
		if err := datastore.Get(c, key, event); err != nil {
			return err
		}

		event.Name = name

		_, err := datastore.Put(c, key, event)
		return err
	}, nil)
	if err == datastore.ErrNoSuchEntity {
		return &appError{err, "Not found", http.StatusNotFound}
	}
	if err != nil {
		return &appError{err, "Error updating event",
			http.StatusInternalServerError}
	}

	http.Redirect(w, r, eventDetailURL+"?id="+idStr, http.StatusFound)
	return nil
}

func deleteEventHandler(w http.ResponseWriter, r *http.Request) *appError {
	if !user.IsAdmin(appengine.NewContext(r)) {
		msg := "Forbidden"
		return &appError{errors.New(msg), msg, http.StatusForbidden}
	}

	idStr := r.FormValue("id")
	idInt, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return &appError{err, "Invalid ID", http.StatusBadRequest}
	}
	c := appengine.NewContext(r)
	key := datastore.NewKey(c, eventKind, "", idInt, nil)
	err = datastore.RunInTransaction(c, func(c appengine.Context) error {
		return datastore.Delete(c, key)
	}, nil)
	if err == datastore.ErrNoSuchEntity {
		return &appError{err, "Not found", http.StatusNotFound}
	}
	if err != nil {
		return &appError{err, "Error updating event",
			http.StatusInternalServerError}
	}

	http.Redirect(w, r, eventListURL, http.StatusFound)
	return nil
}
