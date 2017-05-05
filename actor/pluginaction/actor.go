// Package pluginaction handles all operations related to plugin commands
package pluginaction

// Actor handles all plugin actions
type Actor struct {
	config Config
	client PluginClient
}

// NewActor returns a pluginaction Actor
func NewActor(config Config, client PluginClient) *Actor {
	return &Actor{config: config, client: client}
}
