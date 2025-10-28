package resources

// APILink represents a generic link from a response object.
type APILink struct {
	// HREF is the fully qualified URL for the link.
	HREF string `json:"href"`
	// Method indicate the desired action to be performed on the identified
	// resource.
	Method string `json:"method"`

	// Meta contains additional metadata about the API.
	Meta APILinkMeta `json:"meta"`
}

type APILinkMeta struct {
	// Version of the API
	Version string `json:"version"`

	// Fingerprint to authenticate api with
	HostKeyFingerprint string `json:"host_key_fingerprint"`

	// Identifier for UAA queries
	OAuthClient string `json:"oauth_client"`
}

// APILinks is a directory of follow-up urls for the resource.
type APILinks map[string]APILink
