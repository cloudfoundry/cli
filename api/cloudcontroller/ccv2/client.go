package ccv2

import "code.cloudfoundry.org/cli/api/cloudcontroller"

type Warnings []string

//go:generate counterfeiter . ConnectionWrapper

// ConnectionWrapper can wrap a given connection allowing the wrapper to modify
// all requests going in and out of the given connection.
type ConnectionWrapper interface {
	cloudcontroller.Connection
	Wrap(innerconnection cloudcontroller.Connection) cloudcontroller.Connection
}

type CloudControllerClient struct {
	authorizationEndpoint     string
	cloudControllerAPIVersion string
	cloudControllerURL        string
	dopplerEndpoint           string
	loggregatorEndpoint       string
	routingEndpoint           string
	tokenEndpoint             string

	connection cloudcontroller.Connection
}

// NewCloudControllerClient returns a new CloudControllerClient
func NewCloudControllerClient() *CloudControllerClient {
	return new(CloudControllerClient)
}

// WrapConnection wraps the current CloudControllerClient connection in the
// wrapper
func (client *CloudControllerClient) WrapConnection(wrapper ConnectionWrapper) {
	client.connection = wrapper.Wrap(client.connection)
}
