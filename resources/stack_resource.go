package resources

type Stack struct {
	// GUID is a unique stack identifier.
	GUID string `json:"guid"`
	// Name is the name of the stack.
	Name string `json:"name"`
	// Description is the description for the stack
	Description string `json:"description"`

	// Metadata is used for custom tagging of API resources
	Metadata *Metadata `json:"metadata,omitempty"`
}
