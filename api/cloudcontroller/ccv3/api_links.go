package ccv3

// APILink represents a generic link from a response object.
type APILink struct {
	// HREF is the fully qualified URL for the link.
	HREF string `json:"href"`

	// Meta contains additional metadata about the API.
	Meta struct {
		// Version of the API
		Version string `json:"version"`
	} `json:"meta"`
}
