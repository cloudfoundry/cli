// Package v2action contains the business logic for the commands/v6 package
// Actors in this package should only call CC v2 API endpoints
package v2action

// Warnings is a list of warnings returned back from the cloud controller
type Warnings []string

// Actor handles all business logic for Cloud Controller v2 operations.
type Actor struct {
	CloudControllerClient CloudControllerClient
	Config                Config
	UAAClient             UAAClient

	domainCache map[string]Domain
}

// NewActor returns a new actor.
func NewActor(ccClient CloudControllerClient, uaaClient UAAClient, config Config) *Actor {
	return &Actor{
		CloudControllerClient: ccClient,
		Config:                config,
		UAAClient:             uaaClient,
		domainCache:           map[string]Domain{},
	}
}
