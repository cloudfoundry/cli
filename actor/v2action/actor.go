// Package v2action contains the business logic for the commands/v2 package
package v2action

// Warnings is a list of warnings returned back from the cloud controller
type Warnings []string

type Actor struct {
	CloudControllerClient CloudControllerClient
	UAAClient             UAAClient
	Config                Config
}

func NewActor(ccClient CloudControllerClient, uaaClient UAAClient, config Config) Actor {
	return Actor{
		CloudControllerClient: ccClient,
		Config:                config,
		UAAClient:             uaaClient,
	}
}
