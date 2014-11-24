package mypkg

// A URL entry in the top navigation bar
type navURL struct {
	URL, Text, Tooltip string
}

const (
	homepageURL    = "/"
	createEventURL = "/create_event"
	eventListURL   = "/events"
	eventDetailURL = "/event"
)

var (
	publicNavURLs []*navURL = []*navURL{
		&navURL{"/", "Home", ""},
	}

	loggedInNavURLs []*navURL = addNavURLs(publicNavURLs, []*navURL{
		&navURL{eventListURL, "Events", "List all events"},
	})

	adminNavURLs []*navURL = addNavURLs(loggedInNavURLs, []*navURL{
		&navURL{createEventURL, "Add Event", "Create a new Event"},
	})
)

// Make a new slice, append base and extra to it, and return it.
func addNavURLs(base, extra []*navURL) []*navURL {
	result := []*navURL{}
	result = append(result, base...)
	result = append(result, extra...)
	return result
}
