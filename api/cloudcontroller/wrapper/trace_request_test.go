package wrapper_test

import (
	"bytes"
	"net/http"

	"code.cloudfoundry.org/cli/v9/api/cloudcontroller"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/cloudcontrollerfakes"
	. "code.cloudfoundry.org/cli/v9/api/cloudcontroller/wrapper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CCTraceHeaderRequest", func() {
	var (
		fakeConnection *cloudcontrollerfakes.FakeConnection

		wrapper cloudcontroller.Connection

		request  *cloudcontroller.Request
		response *cloudcontroller.Response
		makeErr  error

		traceHeader string
	)

	BeforeEach(func() {
		fakeConnection = new(cloudcontrollerfakes.FakeConnection)

		traceHeader = "trace-id"

		wrapper = NewCCTraceHeaderRequest(traceHeader).Wrap(fakeConnection)

		body := bytes.NewReader([]byte("foo"))

		req, err := http.NewRequest(http.MethodGet, "https://foo.bar.com/banana", body)
		Expect(err).NotTo(HaveOccurred())

		response = &cloudcontroller.Response{
			RawResponse:  []byte("some-response-body"),
			HTTPResponse: &http.Response{},
		}
		request = cloudcontroller.NewRequest(req, body)
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
