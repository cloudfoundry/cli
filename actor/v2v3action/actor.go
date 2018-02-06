// Package v2v3action contains business logic that involves both v2action
// and v3action
package v2v3action

// Warnings is a list of warnings returned back from the cloud controller
type Warnings []string

// Actor handles all business logic for Cloud Controller v2 and v3 operations.
type Actor struct {
	V2Actor V2Actor
	V3Actor V3Actor
}

// NewActor returns a new actor.
func NewActor(v2Actor V2Actor, v3Actor V3Actor) *Actor {
	return &Actor{
		V2Actor: v2Actor,
		V3Actor: v3Actor,
	}
}
