// Package sharedaction handles all operations that do not require a cloud
// controller
package sharedaction

type AuthActor interface {
	IsLoggedIn() bool
}

// Actor handles all shared actions
type Actor struct {
	Config Config
	AuthActor
}

// NewActor returns an Actor with default settings
func NewActor(config Config) *Actor {
	var authActor AuthActor = NewDefaultAuthActor(config)
	if config.IsCFOnK8s() {
		authActor = NewK8sAuthActor(config)
	}

	return &Actor{
		AuthActor: authActor,
		Config:    config,
	}
}
