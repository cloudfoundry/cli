// Package cfnetworkingaction contains the business logic for the cf networking commands.
package cfnetworkingaction

// Warnings is a list of warnings returned back
type Warnings []string

// Actor handles all business logic for cf networking operations.
type Actor struct {
	NetworkingClient      NetworkingClient
	CloudControllerClient CloudControllerClient
}

// NewActor returns a new actor.
func NewActor(networkingClient NetworkingClient, ccClient CloudControllerClient) *Actor {
	return &Actor{
		NetworkingClient:      networkingClient,
		CloudControllerClient: ccClient,
	}
}
