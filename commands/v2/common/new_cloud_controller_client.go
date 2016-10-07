package common

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/commands"
)

func NewCloudControllerClient(config commands.Config) (*ccv2.CloudControllerClient, error) {
	client := ccv2.NewCloudControllerClient()
	_, err := client.TargetCF(config.Target(), config.SkipSSLValidation())
	if err != nil {
		return nil, err
	}
	client.WrapConnection(wrapper.NewTokenRefreshWrapper(config))
	//Retry Wrapper
	return client, err
}
