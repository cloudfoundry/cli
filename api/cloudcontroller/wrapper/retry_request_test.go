package wrapper_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/cloudcontrollerfakes"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Retry", func() {
	var (
		fakeConnection *cloudcontrollerfakes.FakeConnection
		connectionErr  error

		wrapper cloudcontroller.Connection

		request        *http.Request
		rawRequestBody string
		response       *cloudcontroller.Response
		err            error
	)

	BeforeEach(func() {
		fakeConnection = new(cloudcontrollerfakes.FakeConnection)

		wrapper = NewRetryRequest(2).Wrap(fakeConnection)

		var err error
		request, err = http.NewRequest(http.MethodGet, "https://foo.bar.com/banana", nil)
		Expect(err).NotTo(HaveOccurred())

		rawRequestBody = "banana pants"
		request.Body = ioutil.NopCloser(strings.NewReader(rawRequestBody))

		response = &cloudcontroller.Response{
			HTTPResponse: &http.Response{},
		}
	})

	JustBeforeEach(func() {
		err = wrapper.Make(request, response)
	})

	Describe("Make", func() {
		Context("when no error occurs", func() {
			It("does not retry", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeConnection.MakeCallCount()).To(Equal(1))
			})
		})

		Context("when an error occurs and there's no HTTP Response (aka protocol level error)", func() {
			BeforeEach(func() {
				response.HTTPResponse = nil
				connectionErr = errors.New("ZOMG WAAT")
				fakeConnection.MakeReturns(connectionErr)
			})

			It("does not retry", func() {
				Expect(err).To(Equal(connectionErr))
				Expect(fakeConnection.MakeCallCount()).To(Equal(1))
			})
		})

		Context("when the request receives a 4XX status code", func() {
			BeforeEach(func() {
				response.HTTPResponse.StatusCode = 400
				connectionErr = cloudcontroller.RawHTTPStatusError{
					StatusCode: 400,
				}
				fakeConnection.MakeReturns(connectionErr)
			})

			It("does not retry", func() {
				Expect(err).To(Equal(connectionErr))
				Expect(fakeConnection.MakeCallCount()).To(Equal(1))
			})
		})

		Context("when the request receives a 500 status code", func() {
			BeforeEach(func() {
				connectionErr = cloudcontroller.RawHTTPStatusError{
					StatusCode: 500,
				}
				response.HTTPResponse.StatusCode = 500

				fakeConnection.MakeStub = func(req *http.Request, passedResponse *cloudcontroller.Response) error {
					defer req.Body.Close()
					body, err := ioutil.ReadAll(request.Body)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(Equal(rawRequestBody))
					return connectionErr
				}
			})

			It("retries maxRetries times", func() {
				Expect(err).To(Equal(connectionErr))
				Expect(fakeConnection.MakeCallCount()).To(Equal(3))
			})
		})

		Context("when the request receives a 5XX status code other than 500", func() {
			BeforeEach(func() {
				connectionErr = cloudcontroller.RawHTTPStatusError{
					StatusCode: 501,
				}
				response.HTTPResponse.StatusCode = 501

				fakeConnection.MakeStub = func(req *http.Request, passedResponse *cloudcontroller.Response) error {
					defer req.Body.Close()
					body, err := ioutil.ReadAll(request.Body)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(Equal(rawRequestBody))
					return connectionErr
				}
			})

			It("does not retry", func() {
				Expect(err).To(Equal(connectionErr))
				Expect(fakeConnection.MakeCallCount()).To(Equal(1))
			})
		})
	})
})
