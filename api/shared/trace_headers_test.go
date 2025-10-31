package shared_test

import (
	"net/http"

	. "code.cloudfoundry.org/cli/v9/api/shared"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("B3 Trace Headers", func() {
	Describe("SetHeaders", func() {
		Context("when there are already headers set", func() {
			It("does not add the headers", func() {
				traceHeaders := NewTraceHeaders("new_trace_id")
				request := &http.Request{
					Header: http.Header{},
				}
				request.Header.Set("X-B3-TraceId", "old_trace_id")
				request.Header.Set("X-B3-SpanId", "old_span_id")
				traceHeaders.SetHeaders(request)

				Expect(request.Header.Get("X-B3-TraceId")).To(Equal("old_trace_id"))
				Expect(request.Header.Get("X-B3-SpanId")).To(Equal("old_span_id"))
			})
		})

		Context("when there are no headers set", func() {
			It("adds the headers", func() {
				traceHeaders := NewTraceHeaders("new_trace_id")
				request := &http.Request{
					Header: http.Header{},
				}
				traceHeaders.SetHeaders(request)

				Expect(request.Header.Get("X-B3-TraceId")).To(Equal("new_trace_id"))
				Expect(request.Header.Get("X-B3-SpanId")).ToNot(BeEmpty())
			})
		})
	})
})
