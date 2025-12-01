package resources

type Stack struct {
	// GUID is a unique stack identifier.
	GUID string `json:"guid"`
	// Name is the name of the stack.
	Name string `json:"name"`
	// Description is the description for the stack
	Description string `json:"description"`
	// State is the state of the stack (ACTIVE, RESTRICTED, DEPRECATED, DISABLED)
	State string `json:"state,omitempty"`

	// Metadata is used for custom tagging of API resources
	Metadata *Metadata `json:"metadata,omitempty"`
}
