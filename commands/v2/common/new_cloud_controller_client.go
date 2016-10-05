package common

import (
	"code.cloudfoundry.org/cli/api/cloudcontrollerv2"
	"code.cloudfoundry.org/cli/commands"
)

func NewCloudControllerClient(config commands.Config) (*cloudcontrollerv2.CloudControllerClient, error) {
	client := cloudcontrollerv2.NewCloudControllerClient()
	_, err := client.TargetCF(config.Target(), config.SkipSSLValidation())
	return client, err
}
