package configactions

import "code.cloudfoundry.org/cli/api/cloudcontrollerv2"

//go:generate counterfeiter . CloudControllerClient

type CloudControllerClient interface {
	TargetCF(APIURL string, skipSSLValidation bool) (cloudcontrollerv2.Warnings, error)

	API() string
	APIVersion() string
	AuthorizationEndpoint() string
	LoggregatorEndpoint() string
	DopplerEndpoint() string
	TokenEndpoint() string
}
