package common

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/commands"
)

type NoAPISetError struct {
	BinaryName string
}

func (e NoAPISetError) Error() string {
	return "No API endpoint set. Use '{{.LoginTip}}' or '{{.APITip}}' to target an endpoint."
}

func NewCloudControllerClient(config commands.Config) (*ccv3.CloudControllerClient, error) {
	if config.Target() == "" {
		return nil, NoAPISetError{
			BinaryName: config.BinaryName(),
		}
	}

	client := ccv3.NewCloudControllerClient()
	_, err := client.TargetCF(config.Target(), config.SkipSSLValidation())
	if err != nil {
		return nil, err
	}

	uaaClient := uaa.NewClient(client.UAA, config)
	client.WrapConnection(wrapper.NewUAAAuthentication(uaaClient))
	//Retry Wrapper
	return client, nil
}
