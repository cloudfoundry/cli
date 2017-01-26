package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

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

	Describe("UnexpectedResponseError", func() {
		Describe("Error", func() {
			It("formats the error", func() {
				err := UnexpectedResponseError{
					ResponseCode: 123,
					CCErrorResponse: CCErrorResponse{
						Code:        456,
						Description: "some-error-description",
						ErrorCode:   "some-error-code",
					},
					RequestIDs: []string{
						"6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95",
						"6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95::7445d9db-c31e-410d-8dc5-9f79ec3fc26f",
					},
				}
				Expect(err.Error()).To(Equal(`Unexpected Response
Response code: 123
CC code:       456
CC error code: some-error-code
Request ID:    6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95
Request ID:    6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95::7445d9db-c31e-410d-8dc5-9f79ec3fc26f
Description:   some-error-description`))
			})
		})
	})

	Describe("Make", func() {
		BeforeEach(func() {
			response = `{
					"code": 777,
					"description": "SomeCC Error Message",
					"error_code": "CF-SomeError"
				}`

			client = NewTestClient()
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v2/apps"),
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
			BeforeEach(func() {
				serverResponseCode = http.StatusTeapot
				response = "418 I'm a teapot: Requested route ('some-url.com') does not exist."
			})

			It("returns a RawHTTPStatusError", func() {
				_, _, err := client.GetApplications(nil)
				Expect(err).To(MatchError(cloudcontroller.RawHTTPStatusError{
					StatusCode:  http.StatusTeapot,
					RawResponse: []byte(response),
					RequestIDs: []string{
						"6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95",
						"6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95::7445d9db-c31e-410d-8dc5-9f79ec3fc26f",
					},
				}))
			})
		})

		Context("when the error is from the cloud controller", func() {
			Context("(400) Bad Request", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusBadRequest
				})

				Context("generic 400", func() {
					BeforeEach(func() {
						response = `{
							"description": "bad request",
							"error_code": "CF-BadRequest"
						}`
					})

					It("returns a BadRequestError", func() {
						_, _, err := client.GetApplications(nil)
						Expect(err).To(MatchError(cloudcontroller.BadRequestError{
							Message: "bad request",
						}))
					})
				})

				Context("getting stats for a stopped app", func() {
					BeforeEach(func() {
						response = `{
							"code": 200003,
							"description": "Could not fetch stats for stopped app: some-app",
							"error_code": "CF-AppStoppedStatsError"
						}`
					})

					It("returns an AppStoppedStatsError", func() {
						_, _, err := client.GetApplications(nil)
						Expect(err).To(MatchError(AppStoppedStatsError{
							Message: "Could not fetch stats for stopped app: some-app",
						}))
					})
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
						"code": 1000,
						"description": "Invalid Auth Token",
						"error_code": "CF-InvalidAuthToken"
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

				Context("when the error is a json response from the cloud controller", func() {
					It("returns a ResourceNotFoundError", func() {
						_, _, err := client.GetApplications(nil)
						Expect(err).To(MatchError(cloudcontroller.ResourceNotFoundError{Message: "SomeCC Error Message"}))
					})
				})

				Context("when the error is not from the cloud controller API", func() {
					BeforeEach(func() {
						response = "an error not from the CC API"
					})

					It("returns a NotFoundError", func() {
						_, _, err := client.GetApplications(nil)
						Expect(err).To(MatchError(cloudcontroller.NotFoundError{Message: response}))
					})
				})
			})

			Context("(422) Unprocessable Entity", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusUnprocessableEntity
				})

				It("returns a UnprocessableEntityError", func() {
					_, _, err := client.GetApplications(nil)
					Expect(err).To(MatchError(cloudcontroller.UnprocessableEntityError{Message: "SomeCC Error Message"}))
				})
			})

			Context("unhandled Error Codes", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusTeapot
				})

				It("returns an UnexpectedResponseError", func() {
					_, _, err := client.GetApplications(nil)
					Expect(err).To(MatchError(UnexpectedResponseError{
						ResponseCode: http.StatusTeapot,
						CCErrorResponse: CCErrorResponse{
							Code:        777,
							Description: "SomeCC Error Message",
							ErrorCode:   "CF-SomeError",
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
