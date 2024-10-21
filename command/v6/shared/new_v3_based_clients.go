package shared

import (
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccv3"
	ccWrapper "code.cloudfoundry.org/cli/v7/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/v7/api/uaa"
	uaaWrapper "code.cloudfoundry.org/cli/v7/api/uaa/wrapper"
	"code.cloudfoundry.org/cli/v7/command"
	"code.cloudfoundry.org/cli/v7/command/translatableerror"
)

// NewV3BasedClients creates a new V3 Cloud Controller client and UAA client using the
// passed in config.
func NewV3BasedClients(config command.Config, ui command.UI, targetCF bool) (*ccv3.Client, *uaa.Client, error) {
	ccWrappers := []ccv3.ConnectionWrapper{}

	verbose, location := config.Verbose()
	if verbose {
		ccWrappers = append(ccWrappers, ccWrapper.NewRequestLogger(ui.RequestLoggerTerminalDisplay()))
	}
	if location != nil {
		ccWrappers = append(ccWrappers, ccWrapper.NewRequestLogger(ui.RequestLoggerFileWriter(location)))
	}

	authWrapper := ccWrapper.NewUAAAuthentication(nil, config)

	ccWrappers = append(ccWrappers, authWrapper)
	ccWrappers = append(ccWrappers, ccWrapper.NewRetryRequest(config.RequestRetryCount()))

	ccClient := ccv3.NewClient(ccv3.Config{
		AppName:            config.BinaryName(),
		AppVersion:         config.BinaryVersion(),
		JobPollingTimeout:  config.OverallPollingTimeout(),
		JobPollingInterval: config.PollingInterval(),
		Wrappers:           ccWrappers,
	})

	if !targetCF {
		return ccClient, nil, nil
	}

	if config.Target() == "" {
		return nil, nil, translatableerror.NoAPISetError{
			BinaryName: config.BinaryName(),
		}
	}

	_, _, err := ccClient.TargetCF(ccv3.TargetSettings{
		URL:               config.Target(),
		SkipSSLValidation: config.SkipSSLValidation(),
		DialTimeout:       config.DialTimeout(),
	})
	if err != nil {
		return nil, nil, err
	}

	if ccClient.UAA() == "" {
		return nil, nil, translatableerror.UAAEndpointNotFoundError{}
	}

	uaaClient := uaa.NewClient(config)

	if verbose {
		uaaClient.WrapConnection(uaaWrapper.NewRequestLogger(ui.RequestLoggerTerminalDisplay()))
	}
	if location != nil {
		uaaClient.WrapConnection(uaaWrapper.NewRequestLogger(ui.RequestLoggerFileWriter(location)))
	}

	uaaAuthWrapper := uaaWrapper.NewUAAAuthentication(uaaClient, config)
	uaaClient.WrapConnection(uaaAuthWrapper)
	uaaClient.WrapConnection(uaaWrapper.NewRetryRequest(config.RequestRetryCount()))

	err = uaaClient.SetupResources(ccClient.Login())
	if err != nil {
		return nil, nil, err
	}

	uaaAuthWrapper.SetClient(uaaClient)
	authWrapper.SetClient(uaaClient)

	return ccClient, uaaClient, nil
}
