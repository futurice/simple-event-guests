package mypkg

// A URL entry in the top navigation bar
type navURL struct {
	URL, Text, Tooltip string
}

const (
	homepageURL = "/"
)

var navURLs []*navURL = []*navURL{
	&navURL{homepageURL, "Homepage", "Home Page"},
}
