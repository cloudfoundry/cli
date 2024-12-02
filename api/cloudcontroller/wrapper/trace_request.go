package wrapper

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/shared"
)

// CCTraceHeaderRequest is a wrapper that adds b3 trace headers to requests.
type CCTraceHeaderRequest struct {
	headers    *shared.TraceHeaders
	connection cloudcontroller.Connection
}

// NewCCTraceHeaderRequest returns a pointer to a CCTraceHeaderRequest wrapper.
func NewCCTraceHeaderRequest(trace string) *CCTraceHeaderRequest {
	return &CCTraceHeaderRequest{
		headers: shared.NewTraceHeaders(trace),
	}
}

// Add tracing headers
func (t *CCTraceHeaderRequest) Make(request *cloudcontroller.Request, passedResponse *cloudcontroller.Response) error {
	t.headers.SetHeaders(request.Request)
	return t.connection.Make(request, passedResponse)
}

// Wrap sets the connection in the CCTraceHeaderRequest and returns itself.
func (t *CCTraceHeaderRequest) Wrap(innerconnection cloudcontroller.Connection) cloudcontroller.Connection {
	t.connection = innerconnection
	return t
}
