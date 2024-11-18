package shared

import (
	"net/http"
)

const (
	B3TraceIDHeader = "X-B3-TraceId"
	B3SpanIDHeader  = "X-B3-SpanId"
)

// TraceHeaders sets b3 trace headers to requests.
type TraceHeaders struct {
	b3trace string
	b3span  string
}

// NewTraceHeaders returns a pointer to a TraceHeaderRequest.
func NewTraceHeaders(trace, span string) *TraceHeaders {
	return &TraceHeaders{
		b3trace: trace,
		b3span:  span,
	}
}

// Add tracing headers if they are not already set.
func (t *TraceHeaders) SetHeaders(request *http.Request) {
	// only override the trace headers if they are not already set (e.g. already explicitly set by cf curl)
	if request.Header.Get(B3TraceIDHeader) == "" {
		request.Header.Add(B3TraceIDHeader, t.b3trace)
	}
	if request.Header.Get(B3SpanIDHeader) == "" {
		request.Header.Add(B3SpanIDHeader, t.b3span)
	}
}
