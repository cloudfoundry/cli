package helpers

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	ccWrapper "code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/api/uaa"
	uaaWrapper "code.cloudfoundry.org/cli/api/uaa/wrapper"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/configv3"
)

// CreateCCV2Client constructs a client object able to communicate with the
// cloudcontroller V2 API.
func CreateCCV2Client() (*ccv2.Client, error) {
	config, err := configv3.LoadConfig(configv3.FlagOverride{})
	if err != nil {
		return nil, err
	}

	ccWrappers := []ccv2.ConnectionWrapper{}
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

	_, err = ccClient.TargetCF(ccv2.TargetSettings{
		URL:               config.Target(),
		SkipSSLValidation: config.SkipSSLValidation(),
		DialTimeout:       config.DialTimeout(),
	})
	if err != nil {
		return nil, err
	}

	if ccClient.AuthorizationEndpoint() == "" {
		return nil, translatableerror.AuthorizationEndpointNotFoundError{}
	}

	uaaClient := uaa.NewClient(config)

	uaaAuthWrapper := uaaWrapper.NewUAAAuthentication(nil, config)
	uaaClient.WrapConnection(uaaAuthWrapper)
	uaaClient.WrapConnection(uaaWrapper.NewRetryRequest(config.RequestRetryCount()))

	err = uaaClient.SetupResources(ccClient.AuthorizationEndpoint())
	if err != nil {
		return nil, err
	}

	uaaAuthWrapper.SetClient(uaaClient)
	authWrapper.SetClient(uaaClient)
	return ccClient, nil
}
