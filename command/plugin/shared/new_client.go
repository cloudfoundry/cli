package shared

import (
	"code.cloudfoundry.org/cli/api/plugin"
	"code.cloudfoundry.org/cli/api/plugin/wrapper"
	"code.cloudfoundry.org/cli/command"
)

// NewClients creates a new V2 Cloud Controller client and UAA client using the
// passed in config.
func NewClient(config command.Config, ui command.UI, skipSSLValidation bool) *plugin.Client {

	verbose, location := config.Verbose()

	pluginClient := plugin.NewClient(plugin.Config{
		AppName:           config.BinaryName(),
		AppVersion:        config.BinaryVersion(),
		DialTimeout:       config.DialTimeout(),
		SkipSSLValidation: skipSSLValidation,
	})

	if verbose {
		pluginClient.WrapConnection(wrapper.NewRequestLogger(ui.RequestLoggerTerminalDisplay()))
	}
	if location != nil {
		pluginClient.WrapConnection(wrapper.NewRequestLogger(ui.RequestLoggerFileWriter(location)))
	}

	pluginClient.WrapConnection(wrapper.NewRetryRequest(config.RequestRetryCount()))

	return pluginClient
}
