package wrapper

import (
	"net/http"

	"code.cloudfoundry.org/cli/v8/api/shared"
	"code.cloudfoundry.org/cli/v8/api/uaa"
)

// UAATraceHeaderRequest is a wrapper that adds b3 trace headers to requests.
type UAATraceHeaderRequest struct {
	headers    *shared.TraceHeaders
	connection uaa.Connection
}

// NewUAATraceHeaderRequest returns a pointer to a UAATraceHeaderRequest wrapper.
func NewUAATraceHeaderRequest(trace string) *UAATraceHeaderRequest {
	return &UAATraceHeaderRequest{
		headers: shared.NewTraceHeaders(trace),
	}
}

// Add tracing headers
func (t *UAATraceHeaderRequest) Make(request *http.Request, passedResponse *uaa.Response) error {
	t.headers.SetHeaders(request)
	return t.connection.Make(request, passedResponse)
}

// Wrap sets the connection in the UAATraceHeaderRequest and returns itself.
func (t *UAATraceHeaderRequest) Wrap(innerconnection uaa.Connection) uaa.Connection {
	t.connection = innerconnection
	return t
}
