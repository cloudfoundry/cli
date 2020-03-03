package ccv3

import "code.cloudfoundry.org/cli/api/cloudcontroller"

//go:generate counterfeiter . ConnectionWrapper

// ConnectionWrapper can wrap a given connection allowing the wrapper to modify
// all requests going in and out of the given connection.
type ConnectionWrapper interface {
	cloudcontroller.Connection
	Wrap(innerconnection cloudcontroller.Connection) cloudcontroller.Connection
}

// WrapConnection wraps the current Client connection in the wrapper.
func (requester *RealRequester) WrapConnection(wrapper ConnectionWrapper) {
	requester.connection = wrapper.Wrap(requester.connection)
}
