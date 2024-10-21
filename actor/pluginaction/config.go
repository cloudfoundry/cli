package pluginaction

import "code.cloudfoundry.org/cli/v9/util/configv3"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Config

// Config is a way of getting basic CF configuration
type Config interface {
	AddPlugin(configv3.Plugin)
	AddPluginRepository(repoName string, repoURL string)
	BinaryVersion() string
	GetPlugin(pluginName string) (configv3.Plugin, bool)
	PluginHome() string
	PluginRepositories() []configv3.PluginRepository
	Plugins() []configv3.Plugin
	RemovePlugin(string)
	WritePluginConfig() error
}
