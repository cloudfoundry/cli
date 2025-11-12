package shared

import (
	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"
	ccWrapper "code.cloudfoundry.org/cli/v8/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/v8/api/router"
	routingWrapper "code.cloudfoundry.org/cli/v8/api/router/wrapper"
	"code.cloudfoundry.org/cli/v8/api/uaa"
	uaaWrapper "code.cloudfoundry.org/cli/v8/api/uaa/wrapper"
	"code.cloudfoundry.org/cli/v8/command"
	"code.cloudfoundry.org/cli/v8/command/translatableerror"
)

func GetNewClientsAndConnectToCF(config command.Config, ui command.UI, minVersionV3 string) (*ccv3.Client, *uaa.Client, *router.Client, error) {
	var err error

	uaaClient, err := newWrappedUAAClient(config, ui)
	if err != nil {
		return nil, nil, nil, err
	}

	routingClient, err := newWrappedRoutingClient(config, ui, uaaClient)
	if err != nil {
		return nil, nil, nil, err
	}

	ccClient := NewAuthWrappedCloudControllerClient(config, ui, uaaClient)

	ccClient, err = connectToCF(config, ui, ccClient, minVersionV3)
	if err != nil {
		return nil, nil, nil, err
	}

	return ccClient, uaaClient, routingClient, err
}

func NewWrappedCloudControllerClient(config command.Config, ui command.UI, extraWrappers ...ccv3.ConnectionWrapper) *ccv3.Client {
	ccWrappers := []ccv3.ConnectionWrapper{}

	verbose, location := config.Verbose()
	if verbose {
		ccWrappers = append(ccWrappers, ccWrapper.NewRequestLogger(ui.RequestLoggerTerminalDisplay()))
	}
	if location != nil {
		ccWrappers = append(ccWrappers, ccWrapper.NewRequestLogger(ui.RequestLoggerFileWriter(location)))
	}

	ccWrappers = append(ccWrappers, extraWrappers...)
	ccWrappers = append(ccWrappers, ccWrapper.NewCCTraceHeaderRequest(config.B3TraceID()))
	ccWrappers = append(ccWrappers, ccWrapper.NewRetryRequest(config.RequestRetryCount()))

	return ccv3.NewClient(ccv3.Config{
		AppName:            config.BinaryName(),
		AppVersion:         config.BinaryVersion(),
		JobPollingTimeout:  config.OverallPollingTimeout(),
		JobPollingInterval: config.PollingInterval(),
		Wrappers:           ccWrappers,
	})
}

func NewAuthWrappedCloudControllerClient(config command.Config, ui command.UI, uaaClient *uaa.Client) *ccv3.Client {
	var authWrapper ccv3.ConnectionWrapper
	authWrapper = ccWrapper.NewUAAAuthentication(uaaClient, config)
	if config.IsCFOnK8s() {
		authWrapper = ccWrapper.NewKubernetesAuthentication(
			config,
			v7action.NewDefaultKubernetesConfigGetter(),
		)
	}

	return NewWrappedCloudControllerClient(config, ui, authWrapper)
}

func newWrappedUAAClient(config command.Config, ui command.UI) (*uaa.Client, error) {
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
	uaaClient.WrapConnection(uaaWrapper.NewUAATraceHeaderRequest(config.B3TraceID()))
	uaaClient.WrapConnection(uaaWrapper.NewRetryRequest(config.RequestRetryCount()))

	err = uaaClient.SetupResources(config.UAAEndpoint(), config.AuthorizationEndpoint())
	if err != nil {
		return nil, err
	}

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
	routingWrappers = append(routingWrappers, routingWrapper.NewRoutingTraceHeaderRequest(config.B3TraceID()))

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

	ccClient.TargetCF(ccv3.TargetSettings{
		URL:               config.Target(),
		SkipSSLValidation: config.SkipSSLValidation(),
		DialTimeout:       config.DialTimeout(),
	})

	if minVersionV3 != "" {
		err := command.MinimumCCAPIVersionCheck(config.APIVersion(), minVersionV3)
		if err != nil {
			if _, ok := err.(translatableerror.MinimumCFAPIVersionNotMetError); ok {
				return nil, translatableerror.V3V2SwitchError{}
			}
			return nil, err
		}
	}

	return ccClient, nil
}
