package pluginaction

import "code.cloudfoundry.org/cli/util/configv3"

//go:generate counterfeiter . Config

// Config is a way of getting basic CF configuration
type Config interface {
	AddPlugin(configv3.Plugin)
	AddPluginRepository(repoName string, repoURL string)
	GetPlugin(pluginName string) (configv3.Plugin, bool)
	PluginHome() string
	PluginRepositories() []configv3.PluginRepository
	Plugins() []configv3.Plugin
	RemovePlugin(string)
	WritePluginConfig() error
}
