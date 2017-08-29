// Package cfnetworkingaction contains the business logic for the cf networking commands.
package cfnetworkingaction

// Warnings is a list of warnings returned back
type Warnings []string

// Actor handles all business logic for cf networking operations.
type Actor struct {
	NetworkingClient NetworkingClient
	V3Actor          V3Actor
}

// NewActor returns a new actor.
func NewActor(networkingClient NetworkingClient, v3Actor V3Actor) *Actor {
	return &Actor{
		NetworkingClient: networkingClient,
		V3Actor:          v3Actor,
	}
}
