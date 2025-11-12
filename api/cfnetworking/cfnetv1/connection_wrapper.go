package cfnetv1

import "code.cloudfoundry.org/cli/v8/api/cfnetworking"

// ConnectionWrapper can wrap a given connection allowing the wrapper to modify
// all requests going in and out of the given connection.
//
//counterfeiter:generate . ConnectionWrapper
type ConnectionWrapper interface {
	cfnetworking.Connection
	Wrap(innerconnection cfnetworking.Connection) cfnetworking.Connection
}

// WrapConnection wraps the current Client connection in the wrapper.
func (client *Client) WrapConnection(wrapper ConnectionWrapper) {
	client.connection = wrapper.Wrap(client.connection)
}
