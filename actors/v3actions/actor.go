// Package v3actions contains the business logic for the commands/v3 package
package v3actions

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
