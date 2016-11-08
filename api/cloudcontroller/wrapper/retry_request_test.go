package wrapper_test

import (
	"errors"
	"net/http"

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

		request  *http.Request
		response *cloudcontroller.Response
		err      error
	)

	BeforeEach(func() {
		fakeConnection = new(cloudcontrollerfakes.FakeConnection)

		wrapper = NewRetryRequest(2).Wrap(fakeConnection)

		var err error
		request, err = http.NewRequest(http.MethodGet, "https://foo.bar.com/banana", nil)
		Expect(err).NotTo(HaveOccurred())

		response = &cloudcontroller.Response{
			HTTPResponse: &http.Response{},
		}
	})

	JustBeforeEach(func() {
		fakeConnection.MakeReturns(connectionErr)
		err = wrapper.Make(request, response)
	})

	Describe("Make", func() {
		Context("when no error occurs", func() {
			BeforeEach(func() {
				connectionErr = nil
			})

			It("does not retry", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeConnection.MakeCallCount()).To(Equal(1))
			})
		})

		Context("when an error occurs and there's no HTTP Response (aka protocol level error)", func() {
			BeforeEach(func() {
				connectionErr = errors.New("ZOMG WAAT")
				response.HTTPResponse = nil
			})

			It("does not retry", func() {
				Expect(err).To(Equal(connectionErr))
				Expect(fakeConnection.MakeCallCount()).To(Equal(1))
			})
		})

		Context("when the request recieves a 4XX status code", func() {
			BeforeEach(func() {
				connectionErr = cloudcontroller.RawHTTPStatusError{
					StatusCode: 400,
				}
				response.HTTPResponse.StatusCode = 400
			})

			It("does not retry", func() {
				Expect(err).To(Equal(connectionErr))
				Expect(fakeConnection.MakeCallCount()).To(Equal(1))
			})
		})

		Context("when the request recieves a 5XX status code", func() {
			BeforeEach(func() {
				connectionErr = cloudcontroller.RawHTTPStatusError{
					StatusCode: 500,
				}
				response.HTTPResponse.StatusCode = 500
			})

			It("retries maxRetries times", func() {
				Expect(err).To(Equal(connectionErr))
				Expect(fakeConnection.MakeCallCount()).To(Equal(3))
			})
		})
	})
})
