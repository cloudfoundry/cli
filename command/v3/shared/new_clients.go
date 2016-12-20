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
		DialTimeout:       config.DialTimeout(),
		SkipSSLValidation: config.SkipSSLValidation(),
		Store:             config,
		URL:               ccClient.UAA(),
	})

	uaaClient.WrapConnection(uaaWrapper.NewErrorWrapper())

	verbose, location := config.Verbose()
	if verbose {
		ccClient.WrapConnection(ccWrapper.NewRequestLogger(command.NewRequestLoggerTerminalDisplay(ui)))
		uaaClient.WrapConnection(uaaWrapper.NewRequestLogger(command.NewRequestLoggerTerminalDisplay(ui)))
	}
	if location != nil {
		ccClient.WrapConnection(ccWrapper.NewRequestLogger(command.NewRequestLoggerFileWriter(ui, location)))
		uaaClient.WrapConnection(uaaWrapper.NewRequestLogger(command.NewRequestLoggerFileWriter(ui, location)))
	}

	ccClient.WrapConnection(ccWrapper.NewUAAAuthentication(uaaClient))
	ccClient.WrapConnection(ccWrapper.NewRetryRequest(2))

	return ccClient, nil
}
