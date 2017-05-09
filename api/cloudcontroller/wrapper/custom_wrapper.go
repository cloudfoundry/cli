package wrapper

import "code.cloudfoundry.org/cli/api/cloudcontroller"

// CustomWrapper is a wrapper that can execute arbitrary code via the
// CustomMake function on every request that passes through Make.
type CustomWrapper struct {
	connection cloudcontroller.Connection
	CustomMake func(connection cloudcontroller.Connection, request *cloudcontroller.Request, passedResponse *cloudcontroller.Response) error
}

func (e *CustomWrapper) Wrap(innerconnection cloudcontroller.Connection) cloudcontroller.Connection {
	e.connection = innerconnection
	return e
}

func (e *CustomWrapper) Make(request *cloudcontroller.Request, passedResponse *cloudcontroller.Response) error {
	return e.CustomMake(e.connection, request, passedResponse)
}
