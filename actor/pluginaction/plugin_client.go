package pluginaction

import "code.cloudfoundry.org/cli/v7/api/plugin"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . PluginClient

type PluginClient interface {
	GetPluginRepository(repositoryURL string) (plugin.PluginRepository, error)
	DownloadPlugin(pluginURL string, path string, proxyReader plugin.ProxyReader) error
}
