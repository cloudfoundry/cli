package shared

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	ccWrapper "code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/api/uaa"
	uaaWrapper "code.cloudfoundry.org/cli/api/uaa/wrapper"
	"code.cloudfoundry.org/cli/command"
)

// NewClients creates a new V3 Cloud Controller client and UAA client using the
// passed in config.
func NewClients(config command.Config, ui command.UI) (*ccv3.Client, error) {
	if config.Target() == "" {
		return nil, command.NoAPISetError{
			BinaryName: config.BinaryName(),
		}
	}

	ccClient := ccv3.NewClient(config.BinaryName(), config.BinaryVersion())
	_, err := ccClient.TargetCF(ccv3.TargetSettings{
		URL:               config.Target(),
		SkipSSLValidation: config.SkipSSLValidation(),
		DialTimeout:       config.DialTimeout(),
	})
	if err != nil {
		return nil, ClientTargetError{Message: err.Error()}
	}

	uaaClient := uaa.NewClient(uaa.Config{
		AppName:           config.BinaryName(),
		AppVersion:        config.BinaryVersion(),
		ClientID:          config.UAAOAuthClient(),
		ClientSecret:      config.UAAOAuthClientSecret(),
		DialTimeout:       config.DialTimeout(),
		SkipSSLValidation: config.SkipSSLValidation(),
		URL:               ccClient.UAA(),
	})

	verbose, location := config.Verbose()
	if verbose {
		ccClient.WrapConnection(ccWrapper.NewRequestLogger(ui.RequestLoggerTerminalDisplay()))
		uaaClient.WrapConnection(uaaWrapper.NewRequestLogger(ui.RequestLoggerTerminalDisplay()))
	}
	if location != nil {
		ccClient.WrapConnection(ccWrapper.NewRequestLogger(ui.RequestLoggerFileWriter(location)))
		uaaClient.WrapConnection(uaaWrapper.NewRequestLogger(ui.RequestLoggerFileWriter(location)))
	}

	ccClient.WrapConnection(ccWrapper.NewUAAAuthentication(uaaClient, config))
	ccClient.WrapConnection(ccWrapper.NewRetryRequest(2))

	return ccClient, nil
}
