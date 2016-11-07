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
func NewClients(config commands.Config) (*ccv3.Client, error) {
	if config.Target() == "" {
		return nil, NoAPISetError{
			BinaryName: config.BinaryName(),
		}
	}

	// TODO: If there is ever a need to create a CC client without the config,
	// this should be pulled out into a NewCloudControllerClient function similar
	// to v2.NewCloudControllerClient
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
		URL:               ccClient.UAA(),
		SkipSSLValidation: config.SkipSSLValidation(),
		Store:             config,
	})
	ccClient.WrapConnection(wrapper.NewUAAAuthentication(uaaClient))
	//Retry Wrapper
	return ccClient, nil
}
