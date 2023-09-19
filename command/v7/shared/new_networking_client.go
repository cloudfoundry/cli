package shared

import (
	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/cfnetv1"
	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/wrapper"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

// NewNetworkingClient creates a new cfnetworking client.
func NewNetworkingClient(apiURL string, config command.Config, uaaClient *uaa.Client, ui command.UI) (*cfnetv1.Client, error) {
	if apiURL == "" {
		return nil, translatableerror.CFNetworkingEndpointNotFoundError{}
	}

	wrappers := []cfnetv1.ConnectionWrapper{}

	verbose, location := config.Verbose()
	if verbose {
		wrappers = append(wrappers, wrapper.NewRequestLogger(ui.RequestLoggerTerminalDisplay()))
	}
	if location != nil {
		wrappers = append(wrappers, wrapper.NewRequestLogger(ui.RequestLoggerFileWriter(location)))
	}

	authWrapper := wrapper.NewUAAAuthentication(uaaClient, config)
	wrappers = append(wrappers, authWrapper)

	wrappers = append(wrappers, wrapper.NewRetryRequest(config.RequestRetryCount()))

	return cfnetv1.NewClient(cfnetv1.Config{
		AppName:           config.BinaryName(),
		AppVersion:        config.BinaryVersion(),
		DialTimeout:       config.DialTimeout(),
		SkipSSLValidation: config.SkipSSLValidation(),
		URL:               apiURL,
		Wrappers:          wrappers,
	}), nil
}
