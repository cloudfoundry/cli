// Package pushaction contains the business logic for orchestrating a V2 app
// push.
package pushaction

// Warnings is a list of warnings returned back from the cloud controller
type Warnings []string

// Actor handles all business logic for Cloud Controller v2 operations.
type Actor struct {
	V2Actor V2Actor
}

// NewActor returns a new actor.
func NewActor(v2Actor V2Actor) *Actor {
	return &Actor{
		V2Actor: v2Actor,
	}
}
