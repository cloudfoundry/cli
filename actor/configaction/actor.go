// Package configaction handles all operations done to the CF Config.
package configaction

// Warnings is a list of warnings returned back from the cloud controller
type Warnings []string

// Actor handles all operations done to the CF Config
type Actor struct {
	Config                Config
	CloudControllerClient CloudControllerClient
}

// NewActor returns an Actor with default settings
func NewActor(config Config, client CloudControllerClient) Actor {
	return Actor{
		Config:                config,
		CloudControllerClient: client,
	}
}
