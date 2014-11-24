package mypkg

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type guestT struct {
	Name, Email, HostEmail string
	// auto-generated, uniquely identifies a guest in an event
	Code string
}

const (
	guestCodeLen   = 5
	guestCodeSpace = "0123456789" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz"
	guestCodeAttempts = 10
)

func init() {
	http.Handle(addGuestURL, appHandler(addGuestHandler))
	http.Handle(deleteGuestURL, appHandler(deleteGuestHandler))
	http.Handle(editGuestURL, appHandler(editGuestHandler))
}

func init() {
	rand.Seed(time.Now().Unix())
}

// Generate a new guest code different from all codes in ‘exclude’.
// If a unique code isn't generated after several attempts, give up and
// return an error.
func generateGuestCode(exclude []guestT) (code string, err error) {
	set := make(map[string]interface{})
	for _, guest := range exclude {
		set[guest.Code] = nil
	}

	for n := 0; n < guestCodeAttempts; n++ {
		code = ""
		for len(code) < guestCodeLen {
			i := rand.Intn(len(guestCodeSpace))
			code += guestCodeSpace[i : i+1]
		}
		_, ok := set[code]
		if !ok {
			return
		}
	}
	return "", fmt.Errorf("Failed to generate unique guest code "+
		"after %v attempts", guestCodeAttempts)
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
	key := datastore.NewKey(c, eventKind, "", evIdInt, nil)
	err = datastore.RunInTransaction(c, func(c appengine.Context) error {
		event := &eventT{}
		if err := datastore.Get(c, key, event); err != nil {
			return err
		}

		code, err := generateGuestCode(event.Guests)
		if err != nil {
			return err
		}

		event.Guests = append(event.Guests, guestT{
			Name:      name,
			Email:     email,
			HostEmail: hostEmail,
			Code:      code,
		})

		_, err = datastore.Put(c, key, event)
		return err
	}, nil)
	if err == datastore.ErrNoSuchEntity {
		return &appError{err, "Event not found", http.StatusNotFound}
	}
	if err != nil {
		return &appError{err, "Error adding guest",
			http.StatusInternalServerError}
	}

	http.Redirect(w, r, eventDetailURL+"?id="+evIdStr, http.StatusFound)
	return nil
}

func deleteGuestHandler(w http.ResponseWriter, r *http.Request) *appError {
	c := appengine.NewContext(r)
	if user.Current(c) == nil {
		msg := "Forbidden"
		return &appError{errors.New(msg), msg, http.StatusForbidden}
	}

	evIdStr, guestCode := r.FormValue("event_id"), r.FormValue("guest_code")
	evId, err := strconv.ParseInt(evIdStr, 10, 64)
	if err != nil {
		return &appError{err, "Invalid ID", http.StatusBadRequest}
	}
	key := datastore.NewKey(c, eventKind, "", evId, nil)
	err = datastore.RunInTransaction(c, func(c appengine.Context) error {
		event := &eventT{}
		if err := datastore.Get(c, key, event); err != nil {
			return err
		}

		newGuests, found := []guestT{}, false
		for _, guest := range event.Guests {
			if guest.Code != guestCode {
				newGuests = append(newGuests, guest)
			} else {
				found = true
			}
		}
		if !found {
			return datastore.ErrNoSuchEntity
		}
		event.Guests = newGuests

		_, err = datastore.Put(c, key, event)
		return err
	}, nil)
	if err == datastore.ErrNoSuchEntity {
		return &appError{err, "Not found", http.StatusNotFound}
	}
	if err != nil {
		return &appError{err, "Error removing guest",
			http.StatusInternalServerError}
	}

	http.Redirect(w, r, eventDetailURL+"?id="+evIdStr, http.StatusFound)
	return nil
}
func editGuestHandler(w http.ResponseWriter, r *http.Request) *appError {
	c := appengine.NewContext(r)
	if user.Current(c) == nil {
		msg := "Forbidden"
		return &appError{errors.New(msg), msg, http.StatusForbidden}
	}

	switch r.Method {
	case "GET":
		return editGuestGETHandler(w, r)
	case "POST":
		return editGuestPOSTHandler(w, r)
	default:
		msg := "Method not allowed ‘" + r.Method + "’"
		return &appError{errors.New(msg), msg,
			http.StatusMethodNotAllowed}
	}
}

func editGuestGETHandler(w http.ResponseWriter, r *http.Request) *appError {
	evIdStr, guestCode := r.FormValue("event_id"), r.FormValue("guest_code")
	evId, err := strconv.ParseInt(evIdStr, 10, 64)
	if err != nil {
		return &appError{err, "Invalid ID", http.StatusBadRequest}
	}

	event := &eventT{}
	c := appengine.NewContext(r)
	key := datastore.NewKey(c, eventKind, "", evId, nil)
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

	var guestP *guestT
	for _, guest := range event.Guests {
		if guest.Code == guestCode {
			guestP = &guest
			break
		}
	}
	if guestP == nil {
		err := fmt.Errorf("Guest %v not found in event %v",
			guestCode, evIdStr)
		return &appError{err, err.Error(), http.StatusNotFound}
	}

	return execNavTempl(r, w, "edit-guest.html", map[string]interface{}{
		"title": "Edit guest ‘" + guestP.Name + "’",
		"guest": guestP,
	})
}

func editGuestPOSTHandler(w http.ResponseWriter, r *http.Request) *appError {
	evIdStr, guestCode := r.FormValue("event_id"), r.FormValue("guest_code")
	evId, err := strconv.ParseInt(evIdStr, 10, 64)
	if err != nil {
		return &appError{err, "Invalid ID", http.StatusBadRequest}
	}

	name, email, hostEmail := r.FormValue("name"), r.FormValue("email"),
		r.FormValue("host_email")
	if len(name)*len(email)*len(hostEmail) == 0 {
		msg := "Name, Email and HostEmail must not be empty"
		return &appError{errors.New(msg), msg, http.StatusBadRequest}
	}

	c := appengine.NewContext(r)
	key := datastore.NewKey(c, eventKind, "", evId, nil)
	err = datastore.RunInTransaction(c, func(c appengine.Context) error {
		event := &eventT{}
		if err := datastore.Get(c, key, event); err != nil {
			return err
		}

		var guestP *guestT
		for i := range event.Guests {
			if event.Guests[i].Code == guestCode {
				guestP = &event.Guests[i]
				break
			}
		}
		if guestP == nil {
			return datastore.ErrNoSuchEntity
		}

		guestP.Name = name
		guestP.Email = email
		guestP.HostEmail = hostEmail

		_, err := datastore.Put(c, key, event)
		return err
	}, nil)
	if err == datastore.ErrNoSuchEntity {
		return &appError{err, "Not found", http.StatusNotFound}
	}
	if err != nil {
		return &appError{err, "Error updating guest",
			http.StatusInternalServerError}
	}

	http.Redirect(w, r, eventDetailURL+"?id="+evIdStr, http.StatusFound)
	return nil
}
