// Package sharedaction handles all operations that do not require a cloud
// controller
package sharedaction

// Actor handles all shared actions
type Actor struct {
	Config Config
}

// NewActor returns an Actor with default settings
func NewActor(config Config) *Actor {
	return &Actor{
		Config: config,
	}
}
