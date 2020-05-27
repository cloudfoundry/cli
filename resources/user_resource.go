package resources

// User represents a Cloud Controller User.
type User struct {
	// GUID is the unique user identifier.
	GUID             string `json:"guid"`
	Username         string `json:"username"`
	PresentationName string `json:"presentation_name"`
	Origin           string `json:"origin"`
}
