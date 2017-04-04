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
		response           string
		serverResponseCode int

		client *Client
	)

	Describe("Make", func() {
		BeforeEach(func() {
			response = `
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
					RespondWith(serverResponseCode, response, http.Header{
						"X-Vcap-Request-Id": {
							"6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95",
							"6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95::7445d9db-c31e-410d-8dc5-9f79ec3fc26f",
						},
					},
					),
				),
			)
		})

		Context("when the error is not from the cloud controller", func() {
			Context("and the raw status is 404", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusNotFound
					response = "some not found message"
				})
				It("returns a NotFoundError", func() {
					_, _, err := client.GetApplications(nil)
					Expect(err).To(MatchError(ccerror.NotFoundError{Message: response}))
				})
			})

			Context("and the raw status is another error", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusTeapot
					response = "418 I'm a teapot: Requested route ('some-url.com') does not exist."
				})
				It("returns a RawHTTPStatusError", func() {
					_, _, err := client.GetApplications(nil)
					Expect(err).To(MatchError(ccerror.RawHTTPStatusError{
						StatusCode:  http.StatusTeapot,
						RawResponse: []byte(response),
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
					response = `{ "errors": [] }`
				})

				It("returns an UnexpectedResponseError", func() {
					_, _, err := client.GetApplications(nil)
					Expect(err).To(MatchError(ccerror.V3UnexpectedResponseError{
						ResponseCode:    http.StatusUnauthorized,
						V3ErrorResponse: ccerror.V3ErrorResponse{Errors: []ccerror.V3Error{}},
					}))
				})
			})

			Context("when there no errors key in the response", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusNotFound
					response = `
						{
							"code": 10000,
							"description": "Unknown request",
							"error_code": "CF-NotFound"
						}`
				})

				It("returns a NotFoundError", func() {
					_, _, err := client.GetApplications(nil)
					Expect(err).To(MatchError(ccerror.NotFoundError{Message: response}))
				})
			})

			Context("(401) Unauthorized", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusUnauthorized
				})

				Context("generic 401", func() {
					It("returns a UnauthorizedError", func() {
						_, _, err := client.GetApplications(nil)
						Expect(err).To(MatchError(ccerror.UnauthorizedError{Message: "SomeCC Error Message"}))
					})
				})

				Context("invalid token", func() {
					BeforeEach(func() {
						response = `{
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
						_, _, err := client.GetApplications(nil)
						Expect(err).To(MatchError(ccerror.InvalidAuthTokenError{Message: "Invalid Auth Token"}))
					})
				})
			})

			Context("(403) Forbidden", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusForbidden
				})

				It("returns a ForbiddenError", func() {
					_, _, err := client.GetApplications(nil)
					Expect(err).To(MatchError(ccerror.ForbiddenError{Message: "SomeCC Error Message"}))
				})
			})

			Context("(404) Not Found", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusNotFound
				})

				It("returns a ResourceNotFoundError", func() {
					_, _, err := client.GetApplications(nil)
					Expect(err).To(MatchError(ccerror.ResourceNotFoundError{Message: "SomeCC Error Message"}))
				})

			})

			Context("(422) Unprocessable Entity", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusUnprocessableEntity
				})

				It("returns a UnprocessableEntityError", func() {
					_, _, err := client.GetApplications(nil)
					Expect(err).To(MatchError(ccerror.UnprocessableEntityError{Message: "SomeCC Error Message"}))
				})
			})

			Context("(503) Service Unavailable", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusServiceUnavailable
				})

				It("returns a ServiceUnavailableError", func() {
					_, _, err := client.GetApplications(nil)
					Expect(err).To(MatchError(ccerror.ServiceUnavailableError{Message: "SomeCC Error Message"}))
				})

				Context("when the title is 'CF-TaskWorkersUnavailable'", func() {
					BeforeEach(func() {
						response = `{
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
						_, _, err := client.GetApplications(nil)
						Expect(err).To(MatchError(ccerror.TaskWorkersUnavailableError{Message: "Task workers are unavailable: Failed to open TCP connection to nsync.service.cf.internal:8787 (getaddrinfo: Name or service not known)"}))
					})
				})
			})

			Context("Unhandled Error Codes", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusTeapot
				})

				It("returns an UnexpectedResponseError", func() {
					_, _, err := client.GetApplications(nil)
					Expect(err).To(MatchError(ccerror.V3UnexpectedResponseError{
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
