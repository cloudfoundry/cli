package common

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/commands"
)

// NewCloudControllerClient creates a new V3 Cloud Controller client using
// the passed in config.
func NewCloudControllerClient(config commands.Config) (*ccv3.Client, error) {
	if config.Target() == "" {
		return nil, NoAPISetError{
			BinaryName: config.BinaryName(),
		}
	}

	client := ccv3.NewClient(config.BinaryName(), cf.Version)
	_, err := client.TargetCF(config.Target(), config.SkipSSLValidation())
	if err != nil {
		return nil, err
	}

	uaaClient := uaa.NewClient(client.UAA, config)
	client.WrapConnection(wrapper.NewUAAAuthentication(uaaClient))
	//Retry Wrapper
	return client, nil
}
