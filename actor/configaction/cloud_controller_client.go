package configaction

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

//go:generate counterfeiter . CloudControllerClient

type CloudControllerClient interface {
	TargetCF(settings ccv2.TargetSettings) (ccv2.Warnings, error)

	API() string
	APIVersion() string
	AuthorizationEndpoint() string
	DopplerEndpoint() string
	LoggregatorEndpoint() string
	RoutingEndpoint() string
	TokenEndpoint() string
}
