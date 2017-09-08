package wrapper_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking"
	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/cfnetworkingfakes"
	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/networkerror"
	. "code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/wrapper"
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
			request := cfnetworking.NewRequest(req, body)

			response := &cfnetworking.Response{
				HTTPResponse: &http.Response{
					StatusCode: responseStatusCode,
				},
			}

			fakeConnection := new(cfnetworkingfakes.FakeConnection)
			expectedErr := networkerror.RawHTTPStatusError{
				StatusCode: responseStatusCode,
			}
			fakeConnection.MakeStub = func(req *cfnetworking.Request, passedResponse *cfnetworking.Response) error {
				defer req.Body.Close()
				body, readBodyErr := ioutil.ReadAll(request.Body)
				Expect(readBodyErr).ToNot(HaveOccurred())
				Expect(string(body)).To(Equal(rawRequestBody))
				return expectedErr
			}

			wrapper := NewRetryRequest(2).Wrap(fakeConnection)
			err = wrapper.Make(request, response)
			Expect(err).To(MatchError(expectedErr))
			Expect(fakeConnection.MakeCallCount()).To(Equal(expectedNumberOfRetries))
		},

		Entry("maxRetries for Non-Post (500) Internal Server Error", http.MethodGet, http.StatusInternalServerError, 3),
		Entry("1 for Post (502) Bad Gateway", http.MethodGet, http.StatusBadGateway, 1),
		Entry("1 for Post (503) Service Unavailable", http.MethodGet, http.StatusServiceUnavailable, 1),
		Entry("1 for Post (504) Gateway Timeout", http.MethodGet, http.StatusGatewayTimeout, 1),

		Entry("1 for 4XX Errors", http.MethodGet, http.StatusNotFound, 1),
	)

	It("does not retry on success", func() {
		req, err := http.NewRequest(http.MethodGet, "https://foo.bar.com/banana", nil)
		Expect(err).NotTo(HaveOccurred())
		request := cfnetworking.NewRequest(req, nil)
		response := &cfnetworking.Response{
			HTTPResponse: &http.Response{
				StatusCode: http.StatusOK,
			},
		}

		fakeConnection := new(cfnetworkingfakes.FakeConnection)
		wrapper := NewRetryRequest(2).Wrap(fakeConnection)

		err = wrapper.Make(request, response)
		Expect(err).ToNot(HaveOccurred())
		Expect(fakeConnection.MakeCallCount()).To(Equal(1))
	})

	Context("when seeking errors", func() {
		var (
			request  *cfnetworking.Request
			response *cfnetworking.Response

			fakeConnection *cfnetworkingfakes.FakeConnection
			wrapper        cfnetworking.Connection
		)

		BeforeEach(func() {
			fakeReadSeeker := new(cfnetworkingfakes.FakeReadSeeker)
			fakeReadSeeker.SeekReturns(0, errors.New("oh noes"))

			req, err := http.NewRequest(http.MethodGet, "https://foo.bar.com/banana", fakeReadSeeker)
			Expect(err).NotTo(HaveOccurred())
			request = cfnetworking.NewRequest(req, fakeReadSeeker)

			response = &cfnetworking.Response{
				HTTPResponse: &http.Response{
					StatusCode: http.StatusInternalServerError,
				},
			}
			fakeConnection = new(cfnetworkingfakes.FakeConnection)
			fakeConnection.MakeReturns(errors.New("some error"))
			wrapper = NewRetryRequest(3).Wrap(fakeConnection)
		})

		It("sets the err on SeekError", func() {
			err := wrapper.Make(request, response)
			Expect(err).To(MatchError("oh noes"))
			Expect(fakeConnection.MakeCallCount()).To(Equal(1))
		})
	})
})
