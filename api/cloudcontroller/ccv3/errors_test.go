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
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("Make", func() {
		var (
			serverResponse     string
			serverResponseCode int
			makeError          error
		)

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

			_, _, makeError = client.GetApplications()
		})

		When("we can't unmarshal the response successfully", func() {
			BeforeEach(func() {
				serverResponse = "I am not unmarshallable"
				serverResponseCode = http.StatusNotFound
			})

			It("returns an unknown http source error", func() {
				Expect(makeError).To(MatchError(ccerror.UnknownHTTPSourceError{StatusCode: serverResponseCode, RawResponse: []byte(serverResponse)}))
			})
		})

		When("the error is from the cloud controller", func() {
			When("an empty list of errors is returned", func() {
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

			When("the error is a 4XX error", func() {
				Context("(400) Bad Request", func() {
					BeforeEach(func() {
						serverResponseCode = http.StatusBadRequest
					})

					When("the query parameter is invalid", func() {
						BeforeEach(func() {
							serverResponse = `
{
   "errors": [
      {
         "detail": "The query parameter is invalid: Missing label_selector value",
         "title": "CF-BadQueryParameter",
         "code": 10005
      }
   ]
}`
						})

						It("returns a BadRequestError", func() {
							Expect(makeError).To(MatchError(ccerror.BadRequestError{Message: "The query parameter is invalid: Missing label_selector value"}))
						})

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

					Context("API is not found", func() {

						BeforeEach(func() {
							serverResponse = `{
								"errors": [
									{
										"detail": "Unknown request",
										"title": "CF-NotFound",
										"code": 10000
									}
								]
							}`
						})

						It("returns a APINotFoundError", func() {
							Expect(makeError).To(MatchError(ccerror.APINotFoundError{URL: server.URL() + "/v3/apps"}))
						})
					})

					When("a process is not found", func() {
						BeforeEach(func() {
							serverResponse = `
{
  "errors": [
    {
      "code": 10010,
      "detail": "Process not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
						})

						It("returns a ProcessNotFoundError", func() {
							Expect(makeError).To(MatchError(ccerror.ProcessNotFoundError{}))
						})
					})

					When("an instance is not found", func() {
						BeforeEach(func() {
							serverResponse = `
{
  "errors": [
    {
      "code": 10010,
      "detail": "Instance not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
						})

						It("returns an InstanceNotFoundError", func() {
							Expect(makeError).To(MatchError(ccerror.InstanceNotFoundError{}))
						})
					})

					When("an application is not found", func() {
						BeforeEach(func() {
							serverResponse = `
{
  "errors": [
    {
      "code": 10010,
      "detail": "App not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
						})

						It("returns an AppNotFoundError", func() {
							Expect(makeError).To(MatchError(ccerror.ApplicationNotFoundError{}))
						})
					})

					When("a droplet is not found", func() {
						BeforeEach(func() {
							serverResponse = `
{
  "errors": [
    {
      "code": 10010,
      "detail": "Droplet not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
						})

						It("returns a DropletNotFoundError", func() {
							Expect(makeError).To(MatchError(ccerror.DropletNotFoundError{}))
						})
					})

					Context("generic not found", func() {

						It("returns a ResourceNotFoundError", func() {
							Expect(makeError).To(MatchError(ccerror.ResourceNotFoundError{Message: "SomeCC Error Message"}))
						})
					})
				})

				Context("(422) Unprocessable Entity", func() {
					BeforeEach(func() {
						serverResponseCode = http.StatusUnprocessableEntity
					})

					When("the name isn't unique to space", func() {
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

					When("the name isn't unique to organization", func() {
						BeforeEach(func() {
							serverResponse = `
{
  "errors": [
    {
      "code": 10008,
      "detail": "Name must be unique per organization",
      "title": "CF-UnprocessableEntity"
    }
  ]
}`
						})

						It("returns a NameNotUniqueInOrgError", func() {
							Expect(makeError).To(MatchError(ccerror.NameNotUniqueInOrgError{}))
						})
					})

					When("the buildpack is invalid", func() {
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

					When("the buildpack is invalid", func() {
						BeforeEach(func() {
							serverResponse = `
{
  "errors": [
    {
      "code": 10008,
      "detail": "Assign a droplet before starting this app.",
      "title": "CF-UnprocessableEntity"
    }
  ]
}`
						})

						It("returns an InvalidBuildpackError", func() {
							Expect(makeError).To(MatchError(ccerror.InvalidStartError{}))
						})
					})

					When("the detail describes something else", func() {
						BeforeEach(func() {
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

						It("returns a UnprocessableEntityError", func() {
							Expect(makeError).To(MatchError(ccerror.UnprocessableEntityError{Message: "SomeCC Error Message"}))
						})
					})
				})
			})

			When("the error is a 5XX error", func() {
				Context("(503) Service Unavailable", func() {
					BeforeEach(func() {
						serverResponseCode = http.StatusServiceUnavailable
					})

					It("returns a ServiceUnavailableError", func() {
						Expect(makeError).To(MatchError(ccerror.ServiceUnavailableError{Message: "SomeCC Error Message"}))
					})

					When("the title is 'CF-TaskWorkersUnavailable'", func() {
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

				Context("all other 5XX", func() {
					BeforeEach(func() {
						serverResponseCode = http.StatusBadGateway
						serverResponse = "I am some text"
					})

					It("returns a ServiceUnavailableError", func() {
						Expect(makeError).To(MatchError(ccerror.V3UnexpectedResponseError{
							ResponseCode: http.StatusBadGateway,
							RequestIDs: []string{
								"6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95",
								"6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95::7445d9db-c31e-410d-8dc5-9f79ec3fc26f",
							},
							V3ErrorResponse: ccerror.V3ErrorResponse{
								Errors: []ccerror.V3Error{{
									Detail: serverResponse,
								}},
							},
						}))
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

			Context("multiple errors", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusTeapot
					serverResponse = `{
							"errors": [
								{
									"code": 1000,
									"detail": "Some CC Error Message",
									"title": "CF-UnprocessableEntity"
								},
								{
									"code": 1001,
									"detail": "Some CC Error Message",
									"title": "CF-UnprocessableEntity"
								}
							]
						}`
				})

				It("returns a MultiError", func() {
					Expect(makeError).To(MatchError(ccerror.MultiError{
						ResponseCode: http.StatusTeapot,
						Errors: []ccerror.V3Error{
							{
								Code:   1000,
								Detail: "Some CC Error Message",
								Title:  "CF-UnprocessableEntity",
							},
							{
								Code:   1001,
								Detail: "Some CC Error Message",
								Title:  "CF-UnprocessableEntity",
							},
						},
					}))
				})
			})
		})
	})
})
