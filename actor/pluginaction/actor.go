// Package pluginaction handles all operations related to plugin commands
package pluginaction

// Actor handles all plugin actions
type Actor struct {
	config Config
}

// NewActor returns a pluginaction Actor
func NewActor(config Config) *Actor {
	return &Actor{
		config: config,
	}
}
