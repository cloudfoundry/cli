package common

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/commands"
)

func NewCloudControllerClient(config commands.Config, ui TerminalDisplay) (*ccv2.Client, error) {
	if config.Target() == "" {
		return nil, NoAPISetError{
			BinaryName: config.BinaryName(),
		}
	}

	client := ccv2.NewClient(config.BinaryName(), cf.Version)
	_, err := client.TargetCF(config.Target(), config.SkipSSLValidation())
	if err != nil {
		return nil, err
	}

	uaaClient := uaa.NewClient(client.TokenEndpoint(), config)
	client.WrapConnection(wrapper.NewUAAAuthentication(uaaClient))

	if verbose, location := config.Verbose(); verbose {
		var logger *wrapper.RequestLogger
		if location == "" {
			logger = wrapper.NewRequestLogger(NewRequestLoggerTerminalDisplay(ui))
		} else {
			logger = wrapper.NewRequestLogger(NewRequestLoggerFileWriter(ui, location))
		}
		client.WrapConnection(logger)
	}
	//Retry Wrapper
	return client, err
}
