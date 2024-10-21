package shared

import (
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccv2"
	ccWrapper "code.cloudfoundry.org/cli/v7/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/v7/api/router"
	routerWrapper "code.cloudfoundry.org/cli/v7/api/router/wrapper"
	"code.cloudfoundry.org/cli/v7/api/uaa"
	uaaWrapper "code.cloudfoundry.org/cli/v7/api/uaa/wrapper"
	"code.cloudfoundry.org/cli/v7/command"
	"code.cloudfoundry.org/cli/v7/command/translatableerror"
)

func NewRouterClient(config command.Config, ui command.UI, uaaClient *uaa.Client) (*router.Client, error) {
	routerConfig := router.Config{
		AppName:    config.BinaryName(),
		AppVersion: config.BinaryVersion(),
		ConnectionConfig: router.ConnectionConfig{
			DialTimeout:       config.DialTimeout(),
			SkipSSLValidation: config.SkipSSLValidation(),
		},
		RoutingEndpoint: config.RoutingEndpoint(),
	}

	routerWrappers := []router.ConnectionWrapper{}

	verbose, location := config.Verbose()

	if verbose {
		routerWrappers = append(routerWrappers, routerWrapper.NewRequestLogger(ui.RequestLoggerTerminalDisplay()))
	}

	if location != nil {
		routerWrappers = append(routerWrappers, routerWrapper.NewRequestLogger(ui.RequestLoggerFileWriter(location)))
	}

	authWrapper := routerWrapper.NewUAAAuthentication(uaaClient, config)
	errorWrapper := routerWrapper.NewErrorWrapper()

	routerWrappers = append(routerWrappers, authWrapper, errorWrapper)
	routerConfig.Wrappers = routerWrappers

	routerClient := router.NewClient(routerConfig)
	return routerClient, nil
}

func GetNewClientsAndConnectToCF(config command.Config, ui command.UI) (*ccv2.Client, *uaa.Client, error) {
	var err error

	ccClient, authWrapper := NewWrappedCloudControllerClient(config, ui)

	ccClient, err = connectToCF(config, ui, ccClient)
	if err != nil {
		return nil, nil, err
	}
	uaaClient, err := newWrappedUAAClient(config, ui, ccClient, authWrapper)

	return ccClient, uaaClient, err
}

func NewWrappedCloudControllerClient(config command.Config, ui command.UI) (*ccv2.Client, *ccWrapper.UAAAuthentication) {
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
	return ccClient, authWrapper
}

func newWrappedUAAClient(config command.Config, ui command.UI, ccClient *ccv2.Client, authWrapper *ccWrapper.UAAAuthentication) (*uaa.Client, error) {
	var err error
	verbose, location := config.Verbose()

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
		return nil, err
	}
	uaaAuthWrapper.SetClient(uaaClient)
	authWrapper.SetClient(uaaClient)

	return uaaClient, nil
}

func connectToCF(config command.Config, ui command.UI, ccClient *ccv2.Client) (*ccv2.Client, error) {
	if config.Target() == "" {
		return nil, translatableerror.NoAPISetError{
			BinaryName: config.BinaryName(),
		}
	}

	_, err := ccClient.TargetCF(ccv2.TargetSettings{
		URL:               config.Target(),
		SkipSSLValidation: config.SkipSSLValidation(),
		DialTimeout:       config.DialTimeout(),
	})
	if err != nil {
		return nil, err
	}
	if err = command.WarnIfAPIVersionBelowSupportedMinimum(ccClient.APIVersion(), ui); err != nil {
		return nil, err
	}
	if ccClient.AuthorizationEndpoint() == "" {
		return nil, translatableerror.AuthorizationEndpointNotFoundError{}
	}

	return ccClient, err
}
