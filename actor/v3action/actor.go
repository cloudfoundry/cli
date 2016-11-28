// Package v3action contains the business logic for the commands/v3 package
package v3action

// This is used for sorting.
type SortOrder string

const (
	Ascending  SortOrder = "Ascending"
	Descending SortOrder = "Descending"
)

// Warnings is a list of warnings returned back from the cloud controller
type Warnings []string

// Actor represents a V3 actor.
type Actor struct {
	CloudControllerClient CloudControllerClient
}

// NewActor returns a new V3 actor.
func NewActor(client CloudControllerClient) Actor {
	return Actor{
		CloudControllerClient: client,
	}
}
