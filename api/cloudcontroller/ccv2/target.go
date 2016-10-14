package ccv2

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// TargetCF sets the client to use the Cloud Controller at the fully qualified
// API URL. skipSSLValidation controls whether a client verifies the server's
// certificate chain and host name. If skipSSLValidation is true, TLS accepts
// any certificate presented by the server and any host name in that
// certificate for *all* client requests going forward.
//
// In this mode, TLS is susceptible to man-in-the-middle attacks. This should
// be used only for testing.
func (client *CloudControllerClient) TargetCF(APIURL string, skipSSLValidation bool) (Warnings, error) {
	client.cloudControllerURL = APIURL

	client.connection = cloudcontroller.NewConnection(client.cloudControllerURL, internal.APIRoutes, skipSSLValidation)
	client.WrapConnection(newErrorWrapper()) //Pretty Sneaky, Sis..

	info, warnings, err := client.Info()
	if err != nil {
		return warnings, err
	}

	client.authorizationEndpoint = info.AuthorizationEndpoint
	client.cloudControllerAPIVersion = info.APIVersion
	client.dopplerEndpoint = info.DopplerEndpoint
	client.loggregatorEndpoint = info.LoggregatorEndpoint
	client.routingEndpoint = info.RoutingEndpoint
	client.tokenEndpoint = info.TokenEndpoint

	return warnings, nil
}
