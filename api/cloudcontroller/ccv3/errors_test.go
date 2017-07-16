package ccv3_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Error Wrapper", func() {
	var (
		serverResponse     string
		serverResponseCode int
		makeError          error

		client *Client
	)

	Describe("Make", func() {
		BeforeEach(func() {
			serverResponse = `
{
  "errors": [
    {
      "code": 777,
      "detail": "SomeCC Error Message",
      "title": "CF-SomeError"
    }
  ]
}`

			client = NewTestClient()
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v3/apps"),
					RespondWith(serverResponseCode, serverResponse, http.Header{
						"X-Vcap-Request-Id": {
							"6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95",
							"6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95::7445d9db-c31e-410d-8dc5-9f79ec3fc26f",
						},
					},
					),
				),
			)

			_, _, makeError = client.GetApplications(nil)
		})

		Context("when the error is not from the cloud controller", func() {
			Context("and the raw status is 404", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusNotFound
					serverResponse = "some not found message"
				})
				It("returns a NotFoundError", func() {
					Expect(makeError).To(MatchError(ccerror.NotFoundError{Message: serverResponse}))
				})
			})

			Context("and the raw status is another error", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusTeapot
					serverResponse = "418 I'm a teapot: Requested route ('some-url.com') does not exist."
				})
				It("returns a RawHTTPStatusError", func() {
					Expect(makeError).To(MatchError(ccerror.RawHTTPStatusError{
						StatusCode:  http.StatusTeapot,
						RawResponse: []byte(serverResponse),
						RequestIDs: []string{
							"6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95",
							"6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95::7445d9db-c31e-410d-8dc5-9f79ec3fc26f",
						},
					}))
				})
			})
		})

		Context("when the error is from the cloud controller", func() {
			Context("when an empty list of errors is returned", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusUnauthorized
					serverResponse = `{ "errors": [] }`
				})

				It("returns an UnexpectedResponseError", func() {
					Expect(makeError).To(MatchError(ccerror.V3UnexpectedResponseError{
						ResponseCode:    http.StatusUnauthorized,
						V3ErrorResponse: ccerror.V3ErrorResponse{Errors: []ccerror.V3Error{}},
					}))
				})
			})

			Context("(401) Unauthorized", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusUnauthorized
				})

				Context("generic 401", func() {
					It("returns a UnauthorizedError", func() {
						Expect(makeError).To(MatchError(ccerror.UnauthorizedError{Message: "SomeCC Error Message"}))
					})
				})

				Context("invalid token", func() {
					BeforeEach(func() {
						serverResponse = `{
							"errors": [
								{
									"code": 1000,
									"detail": "Invalid Auth Token",
									"title": "CF-InvalidAuthToken"
								}
							]
						}`
					})

					It("returns an InvalidAuthTokenError", func() {
						Expect(makeError).To(MatchError(ccerror.InvalidAuthTokenError{Message: "Invalid Auth Token"}))
					})
				})
			})

			Context("(403) Forbidden", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusForbidden
				})

				It("returns a ForbiddenError", func() {
					Expect(makeError).To(MatchError(ccerror.ForbiddenError{Message: "SomeCC Error Message"}))
				})
			})

			Context("(404) Not Found", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusNotFound
				})

				It("returns a ResourceNotFoundError", func() {
					Expect(makeError).To(MatchError(ccerror.ResourceNotFoundError{Message: "SomeCC Error Message"}))
				})

			})

			Context("(422) Unprocessable Entity", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusUnprocessableEntity
					serverResponse = `
{
  "errors": [
    {
      "code": 10008,
      "detail": "SomeCC Error Message",
      "title": "CF-UnprocessableEntity"
    }
  ]
}`
				})

				Context("when the name isn't unique to space", func() {
					BeforeEach(func() {
						serverResponse = `
{
  "errors": [
    {
      "code": 10008,
      "detail": "name must be unique in space",
      "title": "CF-UnprocessableEntity"
    }
  ]
}`
					})

					It("returns a NameNotUniqueInSpaceError", func() {
						Expect(makeError).To(MatchError(ccerror.NameNotUniqueInSpaceError{}))
					})
				})

				Context("when the buildpack is invalid", func() {
					BeforeEach(func() {
						serverResponse = `
{
  "errors": [
    {
      "code": 10008,
      "detail": "Buildpack must be an existing admin buildpack or a valid git URI",
      "title": "CF-UnprocessableEntity"
    }
  ]
}`
					})

					It("returns an InvalidBuildpackError", func() {
						Expect(makeError).To(MatchError(ccerror.InvalidBuildpackError{}))
					})
				})

				Context("when the detail describes something else", func() {
					It("returns a UnprocessableEntityError", func() {
						Expect(makeError).To(MatchError(ccerror.UnprocessableEntityError{Message: "SomeCC Error Message"}))
					})
				})
			})

			Context("(503) Service Unavailable", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusServiceUnavailable
				})

				It("returns a ServiceUnavailableError", func() {
					Expect(makeError).To(MatchError(ccerror.ServiceUnavailableError{Message: "SomeCC Error Message"}))
				})

				Context("when the title is 'CF-TaskWorkersUnavailable'", func() {
					BeforeEach(func() {
						serverResponse = `{
  "errors": [
    {
      "code": 170020,
      "detail": "Task workers are unavailable: Failed to open TCP connection to nsync.service.cf.internal:8787 (getaddrinfo: Name or service not known)",
      "title": "CF-TaskWorkersUnavailable"
    }
  ]
}`
					})

					It("returns a TaskWorkersUnavailableError", func() {
						Expect(makeError).To(MatchError(ccerror.TaskWorkersUnavailableError{Message: "Task workers are unavailable: Failed to open TCP connection to nsync.service.cf.internal:8787 (getaddrinfo: Name or service not known)"}))
					})
				})
			})

			Context("Unhandled Error Codes", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusTeapot
				})

				It("returns an UnexpectedResponseError", func() {
					Expect(makeError).To(MatchError(ccerror.V3UnexpectedResponseError{
						ResponseCode: http.StatusTeapot,
						V3ErrorResponse: ccerror.V3ErrorResponse{
							Errors: []ccerror.V3Error{
								{
									Code:   777,
									Detail: "SomeCC Error Message",
									Title:  "CF-SomeError",
								},
							},
						},
						RequestIDs: []string{
							"6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95",
							"6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95::7445d9db-c31e-410d-8dc5-9f79ec3fc26f",
						},
					}))
				})
			})
		})
	})
})
