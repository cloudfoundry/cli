package pluginaction

import "code.cloudfoundry.org/cli/util/configv3"

//go:generate counterfeiter . Config

// Config is a way of getting basic CF configuration
type Config interface {
	PluginHome() string
	Plugins() map[string]configv3.Plugin
	RemovePlugin(string)
	WritePluginConfig() error
}
