package shared

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	ccWrapper "code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/api/router"
	routerWrapper "code.cloudfoundry.org/cli/api/router/wrapper"
	"code.cloudfoundry.org/cli/api/uaa"
	uaaWrapper "code.cloudfoundry.org/cli/api/uaa/wrapper"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

// NewClients creates a new V2 Cloud Controller client and UAA client using the
// passed in config.
func NewClients(config command.Config, ui command.UI, targetCF bool) (*ccv2.Client, *uaa.Client, error) {
	ccWrappers := []ccv2.ConnectionWrapper{}

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

	ccClient := ccv2.NewClient(ccv2.Config{
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

	_, err := ccClient.TargetCF(ccv2.TargetSettings{
		URL:               config.Target(),
		SkipSSLValidation: config.SkipSSLValidation(),
		DialTimeout:       config.DialTimeout(),
	})
	if err != nil {
		return nil, nil, err
	}

	if err = command.WarnIfAPIVersionBelowSupportedMinimum(ccClient.APIVersion(), ui); err != nil {
		return nil, nil, err
	}

	if ccClient.AuthorizationEndpoint() == "" {
		return nil, nil, translatableerror.AuthorizationEndpointNotFoundError{}
	}

	uaaClient := uaa.NewClient(config)

	if verbose {
		uaaClient.WrapConnection(uaaWrapper.NewRequestLogger(ui.RequestLoggerTerminalDisplay()))
	}
	if location != nil {
		uaaClient.WrapConnection(uaaWrapper.NewRequestLogger(ui.RequestLoggerFileWriter(location)))
	}

	uaaAuthWrapper := uaaWrapper.NewUAAAuthentication(nil, config)
	uaaClient.WrapConnection(uaaAuthWrapper)
	uaaClient.WrapConnection(uaaWrapper.NewRetryRequest(config.RequestRetryCount()))

	err = uaaClient.SetupResources(ccClient.AuthorizationEndpoint())
	if err != nil {
		return nil, nil, err
	}

	uaaAuthWrapper.SetClient(uaaClient)
	authWrapper.SetClient(uaaClient)

	return ccClient, uaaClient, err
}

func NewRouterClient(config command.Config, ui command.UI, uaaClient *uaa.Client) (*router.Client, error) {
	routerConfig := router.Config{
		AppName:    config.BinaryName(),
		AppVersion: config.BinaryVersion(),
	}

	routerWrappers := []router.ConnectionWrapper{}

	verbose, location := config.Verbose()

	if verbose {
		routerWrappers = append(routerWrappers, routerWrapper.NewRequestLogger(ui.RequestLoggerTerminalDisplay()))
	}

	if location != nil {
		routerWrappers = append(routerWrappers, routerWrapper.NewRequestLogger(ui.RequestLoggerFileWriter(location)))
	}

	authWrapper := routerWrapper.NewUAAAuthentication(nil, config)
	// TODO: add verbose wrapper, location wrapper, and retry wrapper

	routerWrappers = append(routerWrappers, authWrapper)

	routerClient := router.NewClient(routerConfig, routerWrappers)
	err := routerClient.SetupResources(config.RoutingEndpoint(), router.ConnectionConfig{
		SkipSSLValidation: config.SkipSSLValidation(),
		DialTimeout:       config.DialTimeout(),
	})

	if err != nil {
		return nil, err
	}
	authWrapper.SetClient(uaaClient)
	return routerClient, nil
}
