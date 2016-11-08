package ccv3_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Cloud Controller Connection", func() {
	var (
		response           string
		serverResponseCode int

		client *Client
	)

	Describe("UnexpectedResponseError", func() {
		Describe("Error", func() {
			It("returns all of the errors joined with newlines", func() {
				err := UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					CCErrorResponse: CCErrorResponse{
						Errors: []CCError{
							{
								Code:   282010,
								Detail: "detail 1",
								Title:  "title-1",
							},
							{
								Code:   10242013,
								Detail: "detail 2",
								Title:  "title-2",
							},
						},
					},
				}

				Expect(err.Error()).To(Equal(`Unexpected Response
Response Code: 418
Code: 282010, Title: title-1, Detail: detail 1
Code: 10242013, Title: title-2, Detail: detail 2`))
			})
		})
	})

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
					RespondWith(serverResponseCode, response),
				),
			)
		})

		Context("when the error is not from the cloud controller", func() {
			BeforeEach(func() {
				serverResponseCode = http.StatusNotFound
				response = "404 Not Found: Requested route ('some-url.com') does not exist."
			})

			It("returns a RawHTTPStatusError", func() {
				_, _, err := client.GetApplications(nil)
				Expect(err).To(MatchError(cloudcontroller.RawHTTPStatusError{
					StatusCode:  http.StatusNotFound,
					RawResponse: []byte(response),
				}))
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
					Expect(err).To(MatchError(UnexpectedResponseError{
						ResponseCode:    http.StatusUnauthorized,
						CCErrorResponse: CCErrorResponse{Errors: []CCError{}},
					}))
				})
			})

			Context("(401) Unauthorized", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusUnauthorized
				})

				Context("generic 401", func() {
					It("returns a UnauthorizedError", func() {
						_, _, err := client.GetApplications(nil)
						Expect(err).To(MatchError(cloudcontroller.UnauthorizedError{Message: "SomeCC Error Message"}))
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
						Expect(err).To(MatchError(cloudcontroller.InvalidAuthTokenError{Message: "Invalid Auth Token"}))
					})
				})
			})

			Context("(403) Forbidden", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusForbidden
				})

				It("returns a ForbiddenError", func() {
					_, _, err := client.GetApplications(nil)
					Expect(err).To(MatchError(cloudcontroller.ForbiddenError{Message: "SomeCC Error Message"}))
				})
			})

			Context("(404) Not Found", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusNotFound
				})

				It("returns a ResourceNotFoundError", func() {
					_, _, err := client.GetApplications(nil)
					Expect(err).To(MatchError(cloudcontroller.ResourceNotFoundError{Message: "SomeCC Error Message"}))
				})
			})

			Context("Unhandled Error Codes", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusTeapot
				})

				It("returns an UnexpectedResponseError", func() {
					_, _, err := client.GetApplications(nil)
					Expect(err).To(MatchError(UnexpectedResponseError{
						ResponseCode: http.StatusTeapot,
						CCErrorResponse: CCErrorResponse{
							Errors: []CCError{
								{
									Code:   777,
									Detail: "SomeCC Error Message",
									Title:  "CF-SomeError",
								},
							},
						},
					}))
				})
			})
		})
	})
})
