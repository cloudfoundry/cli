package common

import (
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/commands"
)

func NewClients(config commands.Config, ui TerminalDisplay) (*ccv2.Client, *uaa.Client, error) {
	if config.Target() == "" {
		return nil, nil, NoAPISetError{
			BinaryName: config.BinaryName(),
		}
	}

	ccClient, _, err := NewCloudControllerClient(
		config.BinaryName(),
		config.Target(),
		config.SkipSSLValidation(),
		config.DialTimeout(),
	)
	if err != nil {
		return nil, nil, err
	}

	uaaClient := uaa.NewClient(ccClient.TokenEndpoint(), config)
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
	//Retry Wrapper
	return ccClient, uaaClient, err
}

func NewCloudControllerClient(
	binaryName string,
	target string,
	skipSSLValidation bool,
	dialTimeout time.Duration,
) (*ccv2.Client, ccv2.Warnings, error) {
	ccClient := ccv2.NewClient(binaryName, cf.Version)
	warnings, err := ccClient.TargetCF(ccv2.TargetSettings{
		URL:               target,
		SkipSSLValidation: skipSSLValidation,
		DialTimeout:       dialTimeout,
	})
	return ccClient, warnings, err
}
