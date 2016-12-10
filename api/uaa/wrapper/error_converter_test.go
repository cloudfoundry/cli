package wrapper_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/uaafakes"
	. "code.cloudfoundry.org/cli/api/uaa/wrapper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Error Wrapper", func() {
	var (
		fakeConnection    *uaafakes.FakeConnection
		wrapper           uaa.Connection
		request           *http.Request
		response          *uaa.Response
		makeErr           error
		fakeConnectionErr uaa.RawHTTPStatusError
	)

	BeforeEach(func() {
		fakeConnection = new(uaafakes.FakeConnection)
		wrapper = NewErrorWrapper().Wrap(fakeConnection)
		request = &http.Request{}
		response = &uaa.Response{}
		fakeConnectionErr = uaa.RawHTTPStatusError{}
	})

	JustBeforeEach(func() {
		makeErr = wrapper.Make(request, response)
	})

	Describe("Make", func() {
		Context("when the error is not from the UAA", func() {
			BeforeEach(func() {
				fakeConnectionErr.StatusCode = http.StatusTeapot
				fakeConnectionErr.RawResponse = []byte("an error that's not from the UAA server")
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

		Context("when the error is from the UAA", func() {
			Context("(400) Bad Request", func() {
				BeforeEach(func() {
					fakeConnectionErr.StatusCode = http.StatusBadRequest
				})

				Context("generic 400", func() {
					BeforeEach(func() {
						fakeConnectionErr.RawResponse = []byte(`{"error":"not invalid_scim_resource"}`)
						fakeConnection.MakeReturns(fakeConnectionErr)
					})

					It("returns a RawHTTPStatusError", func() {
						Expect(fakeConnection.MakeCallCount()).To(Equal(1))

						Expect(makeErr).To(MatchError(fakeConnectionErr))
					})
				})

				Context("invalid scim resource", func() {
					BeforeEach(func() {
						fakeConnectionErr.RawResponse = []byte(`{
  "error": "invalid_scim_resource",
  "error_description": "A username must be provided"
}`)
						fakeConnection.MakeReturns(fakeConnectionErr)
					})

					It("returns an InvalidAuthTokenError", func() {
						Expect(fakeConnection.MakeCallCount()).To(Equal(1))

						Expect(makeErr).To(MatchError(uaa.InvalidSCIMResourceError{Message: "A username must be provided"}))
					})
				})
			})

			Context("(401) Unauthorized", func() {
				BeforeEach(func() {
					fakeConnectionErr.StatusCode = http.StatusUnauthorized
				})

				Context("generic 401", func() {
					BeforeEach(func() {
						fakeConnectionErr.RawResponse = []byte(`{"error":"not invalid_token"}`)
						fakeConnection.MakeReturns(fakeConnectionErr)
					})

					It("returns a RawHTTPStatusError", func() {
						Expect(fakeConnection.MakeCallCount()).To(Equal(1))

						Expect(makeErr).To(MatchError(fakeConnectionErr))
					})
				})

				Context("invalid token", func() {
					BeforeEach(func() {
						fakeConnectionErr.RawResponse = []byte(`{
  "error": "invalid_token",
  "error_description": "your token is invalid!"
}`)
						fakeConnection.MakeReturns(fakeConnectionErr)
					})

					It("returns an InvalidAuthTokenError", func() {
						Expect(fakeConnection.MakeCallCount()).To(Equal(1))

						Expect(makeErr).To(MatchError(uaa.InvalidAuthTokenError{Message: "your token is invalid!"}))
					})
				})
			})

			Context("(403) Forbidden", func() {
				BeforeEach(func() {
					fakeConnectionErr.StatusCode = http.StatusForbidden
				})

				Context("generic 403", func() {
					BeforeEach(func() {
						fakeConnectionErr.RawResponse = []byte(`{"error":"not insufficient_scope"}`)
						fakeConnection.MakeReturns(fakeConnectionErr)
					})

					It("returns a RawHTTPStatusError", func() {
						Expect(fakeConnection.MakeCallCount()).To(Equal(1))

						Expect(makeErr).To(MatchError(fakeConnectionErr))
					})
				})

				Context("insufficient scope", func() {
					BeforeEach(func() {
						fakeConnectionErr.RawResponse = []byte(`
							{
								"error": "insufficient_scope",
								"error_description": "Insufficient scope for this resource",
								"scope": "uaa.admin scim.write scim.create zones.uaa.admin"
							}
`)
						fakeConnection.MakeReturns(fakeConnectionErr)
					})

					It("returns an InsufficientScopeError", func() {
						Expect(fakeConnection.MakeCallCount()).To(Equal(1))

						Expect(makeErr).To(MatchError(uaa.InsufficientScopeError{Message: "Insufficient scope for this resource"}))
					})
				})
			})

			Context("(409) Conflict", func() {
				BeforeEach(func() {
					fakeConnectionErr.StatusCode = http.StatusConflict
					fakeConnectionErr.RawResponse = []byte(`{
	"error": "scim_resource_already_exists",
  "error_description": "Username already in use: some-user"
}`)
					fakeConnection.MakeReturns(fakeConnectionErr)
				})

				It("returns a ConflictError", func() {
					Expect(fakeConnection.MakeCallCount()).To(Equal(1))

					Expect(makeErr).To(MatchError(uaa.ConflictError{Message: "Username already in use: some-user"}))
				})
			})

			Context("unhandled Error Codes", func() {
				BeforeEach(func() {
					fakeConnectionErr.StatusCode = http.StatusTeapot
					fakeConnectionErr.RawResponse = []byte(`{"error":"some-teapot-error"}`)
					fakeConnection.MakeReturns(fakeConnectionErr)
				})

				It("returns a RawHTTPStatusError", func() {
					Expect(fakeConnection.MakeCallCount()).To(Equal(1))

					Expect(makeErr).To(MatchError(fakeConnectionErr))
				})
			})
		})
	})
})
