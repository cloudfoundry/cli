package shared

import (
	"net/http"

	"code.cloudfoundry.org/cli/v9/util/trace"
)

const (
	B3TraceIDHeader = "X-B3-TraceId"
	B3SpanIDHeader  = "X-B3-SpanId"
)

// TraceHeaders sets b3 trace headers to requests.
type TraceHeaders struct {
	b3trace string
}

// NewTraceHeaders returns a pointer to a TraceHeaderRequest.
func NewTraceHeaders(trace string) *TraceHeaders {
	return &TraceHeaders{
		b3trace: trace,
	}
}

// Add tracing headers if they are not already set.
func (t *TraceHeaders) SetHeaders(request *http.Request) {
	// only override the trace headers if they are not already set (e.g. already explicitly set by cf curl)
	if request.Header.Get(B3TraceIDHeader) == "" {
		request.Header.Add(B3TraceIDHeader, t.b3trace)
	}
	if request.Header.Get(B3SpanIDHeader) == "" {
		request.Header.Add(B3SpanIDHeader, trace.GenerateRandomTraceID(16))
	}

	// request.Header.Add(("B3", request.Header.Get(B3TraceIDHeader)+request.Header.Get(B3SpanIDHeader)))
}
