package shared

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	ccWrapper "code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/api/router"
	routingWrapper "code.cloudfoundry.org/cli/api/router/wrapper"
	"code.cloudfoundry.org/cli/api/uaa"
	uaaWrapper "code.cloudfoundry.org/cli/api/uaa/wrapper"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

func GetNewClientsAndConnectToCF(config command.Config, ui command.UI, minVersionV3 string) (*ccv3.Client, *uaa.Client, *router.Client, error) {
	var err error

	ccClient, authWrapper := NewWrappedCloudControllerClient(config, ui)

	ccClient, err = connectToCF(config, ui, ccClient, minVersionV3)
	if err != nil {
		return nil, nil, nil, err
	}

	uaaClient, err := newWrappedUAAClient(config, ui, ccClient, authWrapper)
	if err != nil {
		return nil, nil, nil, err
	}

	routingClient, err := newWrappedRoutingClient(config, ui, uaaClient)

	return ccClient, uaaClient, routingClient, err
}

func NewWrappedCloudControllerClient(config command.Config, ui command.UI) (*ccv3.Client, *ccWrapper.UAAAuthentication) {
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
	return ccClient, authWrapper
}

func newWrappedUAAClient(config command.Config, ui command.UI, ccClient *ccv3.Client, authWrapper *ccWrapper.UAAAuthentication) (*uaa.Client, error) {
	var err error
	verbose, location := config.Verbose()

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

	err = uaaClient.SetupResources(ccClient.UAA())
	if err != nil {
		return nil, err
	}

	uaaAuthWrapper.SetClient(uaaClient)
	authWrapper.SetClient(uaaClient)

	return uaaClient, nil
}

func newWrappedRoutingClient(config command.Config, ui command.UI, uaaClient *uaa.Client) (*router.Client, error) {
	routingConfig := router.Config{
		AppName:    config.BinaryName(),
		AppVersion: config.BinaryVersion(),
		ConnectionConfig: router.ConnectionConfig{
			DialTimeout:       config.DialTimeout(),
			SkipSSLValidation: config.SkipSSLValidation(),
		},
		RoutingEndpoint: config.RoutingEndpoint(),
	}

	routingWrappers := []router.ConnectionWrapper{routingWrapper.NewErrorWrapper()}

	verbose, location := config.Verbose()

	if verbose {
		routingWrappers = append(routingWrappers, routingWrapper.NewRequestLogger(ui.RequestLoggerTerminalDisplay()))
	}

	if location != nil {
		routingWrappers = append(routingWrappers, routingWrapper.NewRequestLogger(ui.RequestLoggerFileWriter(location)))
	}

	authWrapper := routingWrapper.NewUAAAuthentication(uaaClient, config)

	routingWrappers = append(routingWrappers, authWrapper)
	routingConfig.Wrappers = routingWrappers

	routingClient := router.NewClient(routingConfig)

	return routingClient, nil
}

func connectToCF(config command.Config, ui command.UI, ccClient *ccv3.Client, minVersionV3 string) (*ccv3.Client, error) {
	if config.Target() == "" {
		return nil, translatableerror.NoAPISetError{
			BinaryName: config.BinaryName(),
		}
	}

	_, _, err := ccClient.TargetCF(ccv3.TargetSettings{
		URL:               config.Target(),
		SkipSSLValidation: config.SkipSSLValidation(),
		DialTimeout:       config.DialTimeout(),
	})
	if err != nil {
		return nil, err
	}

	if minVersionV3 != "" {
		err = command.MinimumCCAPIVersionCheck(ccClient.CloudControllerAPIVersion(), minVersionV3)
		if err != nil {
			if _, ok := err.(translatableerror.MinimumCFAPIVersionNotMetError); ok {
				return nil, translatableerror.V3V2SwitchError{}
			}
			return nil, err
		}
	}
	return ccClient, nil
}
