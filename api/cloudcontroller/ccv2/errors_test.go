package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Error Wrapper", func() {
	var (
		serverResponse     string
		serverResponseCode int

		client *Client
	)

	Describe("Make", func() {
		BeforeEach(func() {
			serverResponse = `{
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
					RespondWith(serverResponseCode, serverResponse, http.Header{
						"X-Vcap-Request-Id": {
							"6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95",
							"6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95::7445d9db-c31e-410d-8dc5-9f79ec3fc26f",
						},
					},
					),
				),
			)
		})

		Context("when we can't unmarshal the response successfully", func() {
			BeforeEach(func() {
				serverResponse = "I am not unmarshallable"
				serverResponseCode = http.StatusNotFound
			})

			It("returns an unknown http source error", func() {
				_, _, err := client.GetApplications()
				Expect(err).To(MatchError(ccerror.UnknownHTTPSourceError{StatusCode: serverResponseCode, RawResponse: []byte(serverResponse)}))
			})
		})

		Context("when the error is from the cloud controller", func() {
			Context("when the error is a 4XX error", func() {
				Context("(400) Bad Request", func() {
					BeforeEach(func() {
						serverResponseCode = http.StatusBadRequest
					})

					Context("generic 400", func() {
						BeforeEach(func() {
							serverResponse = `{
							"description": "bad request",
							"error_code": "CF-BadRequest"
						}`
						})

						It("returns a BadRequestError", func() {
							_, _, err := client.GetApplications()
							Expect(err).To(MatchError(ccerror.BadRequestError{
								Message: "bad request",
							}))
						})
					})

					Context("when a not staged error is encountered", func() {
						BeforeEach(func() {
							serverResponse = `{
								"description": "App has not finished staging",
								"error_code": "CF-NotStaged"
							}`
						})

						It("returns a NotStagedError", func() {
							_, _, err := client.GetApplications()
							Expect(err).To(MatchError(ccerror.NotStagedError{
								Message: "App has not finished staging",
							}))
						})
					})

					Context("when an instances error is encountered", func() {
						BeforeEach(func() {
							serverResponse = `{
								"description": "instances went bananas",
								"error_code": "CF-InstancesError"
							}`
						})

						It("returns an InstancesError", func() {
							_, _, err := client.GetApplications()
							Expect(err).To(MatchError(ccerror.InstancesError{
								Message: "instances went bananas",
							}))
						})
					})

					Context("when creating a relation that is invalid", func() {
						BeforeEach(func() {
							serverResponse = `{
							"code": 1002,
							"description": "The requested app relation is invalid: the app and route must belong to the same space",
							"error_code": "CF-InvalidRelation"
						}`
						})

						It("returns an InvalidRelationError", func() {
							_, _, err := client.GetApplications()
							Expect(err).To(MatchError(ccerror.InvalidRelationError{
								Message: "The requested app relation is invalid: the app and route must belong to the same space",
							}))
						})
					})

					Context("getting stats for a stopped app", func() {
						BeforeEach(func() {
							serverResponse = `{
							"code": 200003,
							"description": "Could not fetch stats for stopped app: some-app",
							"error_code": "CF-AppStoppedStatsError"
						}`
						})

						It("returns an AppStoppedStatsError", func() {
							_, _, err := client.GetApplications()
							Expect(err).To(MatchError(ccerror.ApplicationStoppedStatsError{
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
							_, _, err := client.GetApplications()
							Expect(err).To(MatchError(ccerror.UnauthorizedError{Message: "SomeCC Error Message"}))
						})
					})

					Context("invalid token", func() {
						BeforeEach(func() {
							serverResponse = `{
						"code": 1000,
						"description": "Invalid Auth Token",
						"error_code": "CF-InvalidAuthToken"
					}`
						})

						It("returns an InvalidAuthTokenError", func() {
							_, _, err := client.GetApplications()
							Expect(err).To(MatchError(ccerror.InvalidAuthTokenError{Message: "Invalid Auth Token"}))
						})
					})
				})

				Context("(403) Forbidden", func() {
					BeforeEach(func() {
						serverResponseCode = http.StatusForbidden
					})

					It("returns a ForbiddenError", func() {
						_, _, err := client.GetApplications()
						Expect(err).To(MatchError(ccerror.ForbiddenError{Message: "SomeCC Error Message"}))
					})
				})

				Context("(404) Not Found", func() {
					BeforeEach(func() {
						serverResponseCode = http.StatusNotFound
					})

					Context("when the error is a json response from the cloud controller", func() {
						It("returns a ResourceNotFoundError", func() {
							_, _, err := client.GetApplications()
							Expect(err).To(MatchError(ccerror.ResourceNotFoundError{Message: "SomeCC Error Message"}))
						})
					})
				})

				Context("(422) Unprocessable Entity", func() {
					BeforeEach(func() {
						serverResponseCode = http.StatusUnprocessableEntity
					})

					It("returns a UnprocessableEntityError", func() {
						_, _, err := client.GetApplications()
						Expect(err).To(MatchError(ccerror.UnprocessableEntityError{Message: "SomeCC Error Message"}))
					})
				})

				Context("unhandled Error Codes", func() {
					BeforeEach(func() {
						serverResponseCode = http.StatusTeapot
					})

					It("returns an UnexpectedResponseError", func() {
						_, _, err := client.GetApplications()
						Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
							ResponseCode: http.StatusTeapot,
							V2ErrorResponse: ccerror.V2ErrorResponse{
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

			Context("when the error is a 5XX error", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusBadGateway
					serverResponse = "I am some text"
				})

				It("returns a V2UnexpectedResponseError with no json", func() {
					_, _, err := client.GetApplications()
					Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
						ResponseCode: http.StatusBadGateway,
						V2ErrorResponse: ccerror.V2ErrorResponse{
							Description: serverResponse,
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
