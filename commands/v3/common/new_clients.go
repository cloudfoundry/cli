package common

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/commands"
)

// NewClients creates a new V3 Cloud Controller client and UAA client using the
// passed in config.
func NewClients(config commands.Config, ui TerminalDisplay) (*ccv3.Client, error) {
	if config.Target() == "" {
		return nil, NoAPISetError{
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
	ccClient.WrapConnection(wrapper.NewUAAAuthentication(uaaClient))

	if verbose, location := config.Verbose(); verbose {
		var logger *wrapper.RequestLogger
		if location == "" {
			logger = wrapper.NewRequestLogger(NewRequestLoggerTerminalDisplay(ui))
		} else {
			logger = wrapper.NewRequestLogger(NewRequestLoggerFileWriter(ui, location))
		}
		ccClient.WrapConnection(logger)
	}

	ccClient.WrapConnection(wrapper.NewRetryRequest(2))
	return ccClient, nil
}
