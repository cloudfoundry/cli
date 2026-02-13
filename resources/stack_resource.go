package resources

import "strings"

// Valid stack states
const (
	StackStateActive     = "ACTIVE"
	StackStateRestricted = "RESTRICTED"
	StackStateDeprecated = "DEPRECATED"
	StackStateDisabled   = "DISABLED"
)

// ValidStackStates contains all valid stack state values
var ValidStackStates = []string{
	StackStateActive,
	StackStateRestricted,
	StackStateDeprecated,
	StackStateDisabled,
}

// ValidStackStatesLowercase returns the valid stack states in lowercase
func ValidStackStatesLowercase() []string {
	lowercase := make([]string, len(ValidStackStates))
	for i, state := range ValidStackStates {
		lowercase[i] = strings.ToLower(state)
	}
	return lowercase
}

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
