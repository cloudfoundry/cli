package wrapper_test

import (
	"bytes"
	"net/http"

	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/uaafakes"
	. "code.cloudfoundry.org/cli/api/uaa/wrapper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CCTraceHeaderRequest", func() {
	var (
		fakeConnection *uaafakes.FakeConnection

		wrapper uaa.Connection

		request  *http.Request
		response *uaa.Response
		makeErr  error

		traceHeader string
		spanHeader  string
	)

	BeforeEach(func() {
		fakeConnection = new(uaafakes.FakeConnection)

		traceHeader = "trace-id"
		spanHeader = "span-id"

		wrapper = NewUAATraceHeaderRequest(traceHeader, spanHeader).Wrap(fakeConnection)

		body := bytes.NewReader([]byte("foo"))

		var err error
		request, err = http.NewRequest(http.MethodGet, "https://foo.bar.com/banana", body)
		Expect(err).NotTo(HaveOccurred())

		response = &uaa.Response{
			RawResponse:  []byte("some-response-body"),
			HTTPResponse: &http.Response{},
		}
	})

	JustBeforeEach(func() {
		makeErr = wrapper.Make(request, response)
	})

	Describe("Make", func() {
		It("Adds the request headers", func() {
			Expect(makeErr).NotTo(HaveOccurred())
			Expect(request.Header.Get("X-B3-TraceId")).To(Equal(traceHeader))
			Expect(request.Header.Get("X-B3-SpanId")).To(Equal(spanHeader))
		})

		It("Calls the inner connection", func() {
			Expect(fakeConnection.MakeCallCount()).To(Equal(1))
			req, resp := fakeConnection.MakeArgsForCall(0)
			Expect(req).To(Equal(request))
			Expect(resp).To(Equal(response))
		})
	})
})
