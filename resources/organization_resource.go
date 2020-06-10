package resources

// Space represents a Cloud Controller V3 Space.
type Space struct {
	// GUID is a unique space identifier.
	GUID string `json:"guid,omitempty"`
	// Name is the name of the space.
	Name string `json:"name"`
	// Relationships list the relationships to the space.
	Relationships Relationships `json:"relationships,omitempty"`
	// Metadata is used for custom tagging of API resources
	Metadata *Metadata `json:"metadata,omitempty"`
}
