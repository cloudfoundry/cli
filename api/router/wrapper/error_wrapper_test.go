package wrapper_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/router"
	"code.cloudfoundry.org/cli/api/router/routererror"
	"code.cloudfoundry.org/cli/api/router/routerfakes"
	. "code.cloudfoundry.org/cli/api/router/wrapper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Error Wrapper", func() {
	var (
		fakeConnection    *routerfakes.FakeConnection
		wrapper           router.Connection
		request           *router.Request
		response          *router.Response
		makeErr           error
		fakeConnectionErr routererror.RawHTTPStatusError
	)

	BeforeEach(func() {
		fakeConnection = new(routerfakes.FakeConnection)
		wrapper = NewErrorWrapper().Wrap(fakeConnection)
		request = &router.Request{}
		response = &router.Response{}
		fakeConnectionErr = routererror.RawHTTPStatusError{}
	})

	JustBeforeEach(func() {
		makeErr = wrapper.Make(request, response)
	})

	Describe("Make", func() {
		When("the HTTP error code is unexpected", func() {
			BeforeEach(func() {
				fakeConnectionErr.StatusCode = http.StatusTeapot
				fakeConnectionErr.RawResponse = []byte("something else blew up")
				fakeConnection.MakeReturns(fakeConnectionErr)
			})

			It("returns a RawHTTPStatusError", func() {
				Expect(fakeConnection.MakeCallCount()).To(Equal(1))
				requestCall, responseCall := fakeConnection.MakeArgsForCall(0)
				Expect(requestCall).To(Equal(request))
				Expect(responseCall).To(Equal(response))

				Expect(makeErr).To(MatchError(fakeConnectionErr))
			})
		})

		Context("401 Unauthorized", func() {
			BeforeEach(func() {
				fakeConnectionErr.StatusCode = http.StatusUnauthorized
			})

			When("the token has expired", func() {
				BeforeEach(func() {
					fakeConnectionErr.RawResponse = []byte(`{"name":"UnauthorizedError","message":"Token is expired"}`)
					fakeConnection.MakeReturns(fakeConnectionErr)
				})

				It("returns an InvalidAuthTokenError", func() {
					Expect(fakeConnection.MakeCallCount()).To(Equal(1))

					Expect(makeErr).To(MatchError(routererror.InvalidAuthTokenError{Message: "Token is expired"}))
				})
			})

			When("the error is a generic 401", func() {
				BeforeEach(func() {
					fakeConnectionErr.RawResponse = []byte(`{"name":"UnauthorizedError","message":"no you can't"}`)
					fakeConnection.MakeReturns(fakeConnectionErr)
				})

				It("returns a RawHTTPStatusError", func() {
					Expect(fakeConnection.MakeCallCount()).To(Equal(1))

					Expect(makeErr).To(MatchError(routererror.RawHTTPStatusError{
						StatusCode:  http.StatusUnauthorized,
						RawResponse: fakeConnectionErr.RawResponse,
					}))
				})
			})

			When("the response cannot be parsed", func() {
				BeforeEach(func() {
					fakeConnectionErr.RawResponse = []byte(`not valid JSON}`)
					fakeConnection.MakeReturns(fakeConnectionErr)
				})

				It("returns the RawHTTPStatusError", func() {
					Expect(fakeConnection.MakeCallCount()).To(Equal(1))

					Expect(makeErr).To(MatchError(routererror.RawHTTPStatusError{
						StatusCode:  http.StatusUnauthorized,
						RawResponse: fakeConnectionErr.RawResponse,
					}))
				})
			})
		})

		Context("404 Not Found", func() {
			BeforeEach(func() {
				fakeConnectionErr.RawResponse = []byte(`{"name":"ResourceNotFoundError","message":"Router Group 'not-a-thing' not found"}`)
				fakeConnectionErr.StatusCode = http.StatusNotFound
				fakeConnection.MakeReturns(fakeConnectionErr)
			})

			It("returns a ResourceNotFoundError", func() {
				Expect(fakeConnection.MakeCallCount()).To(Equal(1))

				Expect(makeErr).To(MatchError(routererror.ResourceNotFoundError{
					Message: "Router Group 'not-a-thing' not found",
				}))
			})
		})
	})
})
