// Package v7action contains the business logic for the commands/v7 package
package v7action

import (
	"code.cloudfoundry.org/clock"
)

// SortOrder is used for sorting.
type SortOrder string

const (
	Ascending  SortOrder = "Ascending"
	Descending SortOrder = "Descending"
)

// Warnings is a list of warnings returned back from the cloud controller
type Warnings []string

// Actor represents a V7 actor.
type Actor struct {
	CloudControllerClient CloudControllerClient
	Config                Config
	SharedActor           SharedActor
	UAAClient             UAAClient
	Clock                 clock.Clock
}

// NewActor returns a new V7 actor.
func NewActor(client CloudControllerClient, config Config, sharedActor SharedActor, uaaClient UAAClient, clk clock.Clock) *Actor {
	return &Actor{
		CloudControllerClient: client,
		Config:                config,
		SharedActor:           sharedActor,
		UAAClient:             uaaClient,
		Clock:                 clk,
	}
}
