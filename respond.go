package mypkg

import (
	"appengine"
	"appengine/datastore"
	"net/http"
	"strconv"
)

func init() {
	http.Handle(respondURL, appHandler(respondHandler))
}

func respondHandler(w http.ResponseWriter, r *http.Request) *appError {
	save, response := false, false
	if r.Method == "POST" {
		save = true
		var err error
		response, err = strconv.ParseBool(r.FormValue("response"))
		if err != nil {
			return &appError{err, "Invalid response",
				http.StatusBadRequest}
		}
	}

	evIdStr, guestCode := r.FormValue("event_id"), r.FormValue("guest_code")
	evId, err := strconv.ParseInt(evIdStr, 10, 64)
	if err != nil {
		return &appError{err, "Invalid ID", http.StatusBadRequest}
	}

	event := &eventT{}
	var guestP *guestT
	c := appengine.NewContext(r)
	key := datastore.NewKey(c, eventKind, "", evId, nil)
	err = datastore.RunInTransaction(c, func(c appengine.Context) error {
		if err := datastore.Get(c, key, event); err != nil {
			return err
		}

		for i := range event.Guests {
			if event.Guests[i].Code == guestCode {
				guestP = &event.Guests[i]
				break
			}
		}
		if guestP == nil {
			return datastore.ErrNoSuchEntity
		}

		if save {
			guestP.HasResponded = true
			guestP.Response = response

			_, err := datastore.Put(c, key, event)
			return err
		}
		return nil
	}, nil)
	if err == datastore.ErrNoSuchEntity {
		return &appError{err, "Not found", http.StatusNotFound}
	}
	if err != nil {
		return &appError{err, "Error saving your response, " +
			"please try again", http.StatusInternalServerError}
	}

	return execTempl(w, "respond.html", map[string]interface{}{
		"title": event.Name,
		"event": event,
		"guest": guestP,
		"saved": save,
	})
}
