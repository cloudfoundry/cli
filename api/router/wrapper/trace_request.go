package wrapper

import (
	"code.cloudfoundry.org/cli/v9/api/router"
	"code.cloudfoundry.org/cli/v9/api/shared"
)

// RoutingTraceHeaderRequest is a wrapper that adds b3 trace headers to requests.
type RoutingTraceHeaderRequest struct {
	headers    *shared.TraceHeaders
	connection router.Connection
}

// NewRoutingTraceHeaderRequest returns a pointer to a RoutingTraceHeaderRequest wrapper.
func NewRoutingTraceHeaderRequest(trace string) *RoutingTraceHeaderRequest {
	return &RoutingTraceHeaderRequest{
		headers: shared.NewTraceHeaders(trace),
	}
}

// Add tracing headers
func (t *RoutingTraceHeaderRequest) Make(request *router.Request, passedResponse *router.Response) error {
	t.headers.SetHeaders(request.Request)
	return t.connection.Make(request, passedResponse)
}

// Wrap sets the connection in the RoutingTraceHeaderRequest and returns itself.
func (t *RoutingTraceHeaderRequest) Wrap(innerconnection router.Connection) router.Connection {
	t.connection = innerconnection
	return t
}
