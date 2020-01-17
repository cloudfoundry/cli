package shared

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/noaabridge"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/util"
	"github.com/cloudfoundry/noaa/consumer"
)

// NewNOAAClient returns back a configured NOAA Client.
func NewNOAAClient(apiURL string, config command.Config, uaaClient *uaa.Client, ui command.UI) *consumer.Consumer {
	client := consumer.New(
		apiURL,
		util.NewTLSConfig(nil, config.SkipSSLValidation()),
		http.ProxyFromEnvironment,
	)
	client.RefreshTokenFrom(noaabridge.NewTokenRefresher(uaaClient, config))
	client.SetMaxRetryCount(config.NOAARequestRetryCount())

	noaaDebugPrinter := NOAADebugPrinter{}

	// if verbose, set debug printer on noaa client
	verbose, location := config.Verbose()

	client.SetDebugPrinter(&noaaDebugPrinter)

	if verbose {
		noaaDebugPrinter.addOutput(ui.RequestLoggerTerminalDisplay())
	}
	if location != nil {
		noaaDebugPrinter.addOutput(ui.RequestLoggerFileWriter(location))
	}

	return client
}
