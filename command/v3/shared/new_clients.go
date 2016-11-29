package shared

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/cf"
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

	ccClient := ccv3.NewClient(config.BinaryName(), cf.Version)
	_, err := ccClient.TargetCF(ccv3.TargetSettings{
		URL:               config.Target(),
		SkipSSLValidation: config.SkipSSLValidation(),
		DialTimeout:       config.DialTimeout(),
	})
	if err != nil {
		return nil, err
	}

	uaaClient := uaa.NewClient(uaa.Config{
		DialTimeout:       config.DialTimeout(),
		SkipSSLValidation: config.SkipSSLValidation(),
		Store:             config,
		URL:               ccClient.UAA(),
	})

	verbose, location := config.Verbose()
	if verbose {
		logger := wrapper.NewRequestLogger(command.NewRequestLoggerTerminalDisplay(ui))
		ccClient.WrapConnection(logger)
	}
	if location != nil {
		logger := wrapper.NewRequestLogger(command.NewRequestLoggerFileWriter(ui, location))
		ccClient.WrapConnection(logger)
	}

	ccClient.WrapConnection(wrapper.NewUAAAuthentication(uaaClient))
	ccClient.WrapConnection(wrapper.NewRetryRequest(2))
	return ccClient, nil
}
