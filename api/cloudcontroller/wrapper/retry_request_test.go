package wrapper_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/cloudcontrollerfakes"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Retry Request", func() {
	DescribeTable("number of retries",
		func(requestMethod string, responseStatusCode int, expectedNumberOfRetries int) {
			rawRequestBody := "banana pants"
			body := strings.NewReader(rawRequestBody)

			req, err := http.NewRequest(requestMethod, "https://foo.bar.com/banana", body)
			Expect(err).NotTo(HaveOccurred())
			request := cloudcontroller.NewRequest(req, body)

			response := &cloudcontroller.Response{
				HTTPResponse: &http.Response{
					StatusCode: responseStatusCode,
				},
			}

			fakeConnection := new(cloudcontrollerfakes.FakeConnection)
			expectedErr := ccerror.RawHTTPStatusError{
				StatusCode: responseStatusCode,
			}
			fakeConnection.MakeStub = func(req *cloudcontroller.Request, passedResponse *cloudcontroller.Response) error {
				defer req.Body.Close()
				body, readErr := ioutil.ReadAll(request.Body)
				Expect(readErr).ToNot(HaveOccurred())
				Expect(string(body)).To(Equal(rawRequestBody))
				return expectedErr
			}

			wrapper := NewRetryRequest(2).Wrap(fakeConnection)
			err = wrapper.Make(request, response)
			Expect(err).To(MatchError(expectedErr))
			Expect(fakeConnection.MakeCallCount()).To(Equal(expectedNumberOfRetries))
		},

		Entry("maxRetries for Non-Post (500) Internal Server Error", http.MethodGet, http.StatusInternalServerError, 3),
		Entry("maxRetries for Non-Post (502) Bad Gateway", http.MethodGet, http.StatusBadGateway, 3),
		Entry("maxRetries for Non-Post (503) Service Unavailable", http.MethodGet, http.StatusServiceUnavailable, 3),
		Entry("maxRetries for Non-Post (504) Gateway Timeout", http.MethodGet, http.StatusGatewayTimeout, 3),

		Entry("1 for Post (500) Internal Server Error", http.MethodPost, http.StatusInternalServerError, 1),
		Entry("1 for Post (502) Bad Gateway", http.MethodPost, http.StatusBadGateway, 1),
		Entry("1 for Post (503) Service Unavailable", http.MethodPost, http.StatusServiceUnavailable, 1),
		Entry("1 for Post (504) Gateway Timeout", http.MethodPost, http.StatusGatewayTimeout, 1),

		Entry("1 for Get 4XX Errors", http.MethodGet, http.StatusNotFound, 1),
	)

	It("does not retry on success", func() {
		req, err := http.NewRequest(http.MethodGet, "https://foo.bar.com/banana", nil)
		Expect(err).NotTo(HaveOccurred())
		request := cloudcontroller.NewRequest(req, nil)
		response := &cloudcontroller.Response{
			HTTPResponse: &http.Response{
				StatusCode: http.StatusOK,
			},
		}

		fakeConnection := new(cloudcontrollerfakes.FakeConnection)
		wrapper := NewRetryRequest(2).Wrap(fakeConnection)

		err = wrapper.Make(request, response)
		Expect(err).ToNot(HaveOccurred())
		Expect(fakeConnection.MakeCallCount()).To(Equal(1))
	})

	Context("when a PipeSeekError is returned from ResetBody", func() {
		var (
			expectedErr error
			request     *cloudcontroller.Request
			response    *cloudcontroller.Response

			fakeConnection *cloudcontrollerfakes.FakeConnection
			wrapper        cloudcontroller.Connection
		)

		BeforeEach(func() {
			body, _ := cloudcontroller.NewPipeBomb()
			req, err := http.NewRequest(http.MethodGet, "https://foo.bar.com/banana", body)
			Expect(err).NotTo(HaveOccurred())
			request = cloudcontroller.NewRequest(req, body)
			response = &cloudcontroller.Response{
				HTTPResponse: &http.Response{
					StatusCode: http.StatusInternalServerError,
				},
			}

			fakeConnection = new(cloudcontrollerfakes.FakeConnection)
			expectedErr = errors.New("oh noes")
			fakeConnection.MakeReturns(expectedErr)

			wrapper = NewRetryRequest(2).Wrap(fakeConnection)
		})

		It("sets the err on PipeSeekError", func() {
			err := wrapper.Make(request, response)
			Expect(err).To(MatchError(ccerror.PipeSeekError{Err: expectedErr}))
			Expect(fakeConnection.MakeCallCount()).To(Equal(1))
		})
	})
})
