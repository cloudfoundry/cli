package common

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/commands"
)

func NewCloudControllerClient(config commands.Config) (*ccv2.CloudControllerClient, error) {
	if config.Target() == "" {
		return nil, NoAPISetError{
			BinaryName: config.BinaryName(),
		}
	}

	client := ccv2.NewCloudControllerClient()
	_, err := client.TargetCF(config.Target(), config.SkipSSLValidation())
	if err != nil {
		return nil, err
	}

	uaaClient := uaa.NewClient(client.AuthorizationEndpoint(), config)
	client.WrapConnection(wrapper.NewUAAAuthentication(uaaClient))
	//Retry Wrapper
	return client, err
}
