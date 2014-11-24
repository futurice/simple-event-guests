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
	event := &eventT{}
	key := datastore.NewKey(c, eventKind, "", evIdInt, nil)
	err = datastore.RunInTransaction(c, func(c appengine.Context) error {
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
