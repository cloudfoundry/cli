package wrapper_test

import (
	"bytes"
	"net/http"

	"code.cloudfoundry.org/cli/v8/api/router"
	"code.cloudfoundry.org/cli/v8/api/router/routerfakes"
	. "code.cloudfoundry.org/cli/v8/api/router/wrapper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CCTraceHeaderRequest", func() {
	var (
		fakeConnection *routerfakes.FakeConnection

		wrapper router.Connection

		request  *router.Request
		response *router.Response
		makeErr  error

		traceHeader string
	)

	BeforeEach(func() {
		fakeConnection = new(routerfakes.FakeConnection)

		traceHeader = "trace-id"
		wrapper = NewRoutingTraceHeaderRequest(traceHeader).Wrap(fakeConnection)

		body := bytes.NewReader([]byte("foo"))

		req, err := http.NewRequest(http.MethodGet, "https://foo.bar.com/banana", body)
		Expect(err).NotTo(HaveOccurred())

		response = &router.Response{
			RawResponse:  []byte("some-response-body"),
			HTTPResponse: &http.Response{},
		}
		request = router.NewRequest(req, body)
	})

	JustBeforeEach(func() {
		makeErr = wrapper.Make(request, response)
	})

	Describe("Make", func() {
		It("Adds the request headers", func() {
			Expect(makeErr).NotTo(HaveOccurred())
			Expect(request.Header.Get("X-B3-TraceId")).To(Equal(traceHeader))
			Expect(request.Header.Get("X-B3-SpanId")).ToNot(BeEmpty())
		})

		It("Calls the inner connection", func() {
			Expect(fakeConnection.MakeCallCount()).To(Equal(1))
			req, resp := fakeConnection.MakeArgsForCall(0)
			Expect(req).To(Equal(request))
			Expect(resp).To(Equal(response))
		})
	})
})
