package ccv2_test

import (
	"fmt"
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
		executeErr         error

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

			// Make a request to a CC endpoint - we stub the response, so the actual endpoint isn't relevant
			_, _, executeErr = client.GetApplications()
		})

		When("we can't unmarshal the response successfully", func() {
			BeforeEach(func() {
				serverResponse = "I am not unmarshallable"
			})

			When("response code is 4XX", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusNotFound
				})

				It("returns an unknown http source error", func() {
					Expect(executeErr).To(MatchError(ccerror.UnknownHTTPSourceError{StatusCode: serverResponseCode, RawResponse: []byte(serverResponse)}))
				})
			})

			When("response code is 5XX", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusBadGateway
				})

				It("returns an unknown http source error", func() {
					Expect(executeErr).To(MatchError(ccerror.UnknownHTTPSourceError{StatusCode: serverResponseCode, RawResponse: []byte(serverResponse)}))
				})
			})
		})

		When("the error is from the cloud controller", func() {
			When("the error is a 4XX error", func() {
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
							Expect(executeErr).To(MatchError(ccerror.BadRequestError{
								Message: "bad request",
							}))
						})
					})

					When("a not staged error is encountered", func() {
						BeforeEach(func() {
							serverResponse = `{
								"description": "App has not finished staging",
								"error_code": "CF-NotStaged"
							}`
						})

						It("returns a NotStagedError", func() {
							Expect(executeErr).To(MatchError(ccerror.NotStagedError{
								Message: "App has not finished staging",
							}))
						})
					})

					When("an instances error is encountered", func() {
						BeforeEach(func() {
							serverResponse = `{
								"description": "instances went bananas",
								"error_code": "CF-InstancesError"
							}`
						})

						It("returns an InstancesError", func() {
							Expect(executeErr).To(MatchError(ccerror.InstancesError{
								Message: "instances went bananas",
							}))
						})
					})

					When("creating a relation that is invalid", func() {
						BeforeEach(func() {
							serverResponse = `{
							"code": 1002,
							"description": "The requested app relation is invalid: the app and route must belong to the same space",
							"error_code": "CF-InvalidRelation"
						}`
						})

						It("returns an InvalidRelationError", func() {
							Expect(executeErr).To(MatchError(ccerror.InvalidRelationError{
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
							Expect(executeErr).To(MatchError(ccerror.ApplicationStoppedStatsError{
								Message: "Could not fetch stats for stopped app: some-app",
							}))
						})
					})

					When("creating a buildpack with nil stack that already exists", func() {
						BeforeEach(func() {
							serverResponse = `{
							 "description": "Buildpack is invalid: stack unique",
							 "error_code": "CF-BuildpackInvalid",
							 "code": 290003
						}`
						})

						It("returns an BuildpackAlreadyExistsWithoutStackError", func() {
							Expect(executeErr).To(MatchError(ccerror.BuildpackAlreadyExistsWithoutStackError{
								Message: "Buildpack is invalid: stack unique",
							}))
						})
					})

					When("creating a buildpack causes a name collision", func() {
						BeforeEach(func() {
							serverResponse = `{
							 "code": 290001,
							 "description": "The buildpack name is already in use: foo",
							 "error_code": "CF-BuildpackNameTaken"
						}`
						})

						It("returns an BuildpackNameTakenError", func() {
							Expect(executeErr).To(MatchError(ccerror.BuildpackNameTakenError{
								Message: "The buildpack name is already in use: foo",
							}))
						})
					})

					When("creating an organization fails because the name is taken", func() {
						BeforeEach(func() {
							serverResponse = `{
								"code": 30002,
								"description": "The organization name is taken: potato",
								"error_code": "CF-OrganizationNameTaken"
							  }`
						})

						It("returns a OrganizationNameTakenError", func() {
							Expect(executeErr).To(MatchError(ccerror.OrganizationNameTakenError{
								Message: "The organization name is taken: potato",
							}))
						})
					})

					When("creating a space fails because the name is taken", func() {
						BeforeEach(func() {
							serverResponse = `{
								"code": 40002,
								"description": "The app space name is taken: potato",
								"error_code": "CF-SpaceNameTaken"
							  }`
						})

						It("returns a SpaceNameTakenError", func() {
							if e, ok := executeErr.(ccerror.UnknownHTTPSourceError); ok {
								fmt.Printf("TV %s", string(e.RawResponse))
							}
							Expect(executeErr).To(MatchError(ccerror.SpaceNameTakenError{
								Message: "The app space name is taken: potato",
							}))
						})
					})

					When("creating a service instance fails because the name is taken", func() {
						BeforeEach(func() {
							serverResponse = `{
								"code": 40002,
								"description": "Service instance name is taken: potato",
								"error_code": "CF-ServiceInstanceNameTaken"
							  }`
						})

						It("returns a ServiceInstanceNameTakenError", func() {
							Expect(executeErr).To(MatchError(ccerror.ServiceInstanceNameTakenError{
								Message: "Service instance name is taken: potato",
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
							Expect(executeErr).To(MatchError(ccerror.UnauthorizedError{Message: "SomeCC Error Message"}))
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
							Expect(executeErr).To(MatchError(ccerror.InvalidAuthTokenError{Message: "Invalid Auth Token"}))
						})
					})
				})

				Context("(403) Forbidden", func() {
					BeforeEach(func() {
						serverResponseCode = http.StatusForbidden
					})

					It("returns a ForbiddenError", func() {
						Expect(executeErr).To(MatchError(ccerror.ForbiddenError{Message: "SomeCC Error Message"}))
					})
				})

				Context("(404) Not Found", func() {
					BeforeEach(func() {
						serverResponseCode = http.StatusNotFound
					})

					When("the error is a json response from the cloud controller", func() {
						It("returns a ResourceNotFoundError", func() {
							Expect(executeErr).To(MatchError(ccerror.ResourceNotFoundError{Message: "SomeCC Error Message"}))
						})
					})
				})

				Context("(422) Unprocessable Entity", func() {
					BeforeEach(func() {
						serverResponseCode = http.StatusUnprocessableEntity
					})

					Context("generic Unprocessable entity", func() {
						It("returns a UnprocessableEntityError", func() {
							Expect(executeErr).To(MatchError(ccerror.UnprocessableEntityError{Message: "SomeCC Error Message"}))
						})
					})

					When("creating a buildpack causes a name and stack collision", func() {
						BeforeEach(func() {
							serverResponse = `{
							 "code": 290000,
							 "description": "The buildpack name foo is already in use for the stack bar",
							 "error_code": "CF-BuildpackNameStackTaken"
						}`
						})

						It("returns an BuildpackAlreadyExistsForStackError", func() {
							Expect(executeErr).To(MatchError(ccerror.BuildpackAlreadyExistsForStackError{
								Message: "The buildpack name foo is already in use for the stack bar",
							}))
						})
					})

				})

				Context("unhandled Error Codes", func() {
					BeforeEach(func() {
						serverResponseCode = http.StatusTeapot
					})

					It("returns an UnexpectedResponseError", func() {
						Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
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

			When("the error is a 5XX error", func() {
				BeforeEach(func() {
					serverResponseCode = http.StatusInternalServerError
					serverResponse = "{}"
				})

				Context("(502) Bad Gateway", func() {
					BeforeEach(func() {
						serverResponseCode = http.StatusBadGateway
					})

					When("the service broker catalog is invalid", func() {
						BeforeEach(func() {
							serverResponse = `{
								"description": "Service broker catalog is invalid: \nService overview-service\n  Service dashboard client id must be unique\n",
								"error_code": "CF-ServiceBrokerCatalogInvalid",
								"code": 270012
							}`
						})

						It("returns a ServiceBrokerCatalogInvalidError", func() {
							Expect(executeErr).To(MatchError(ccerror.ServiceBrokerCatalogInvalidError{
								Message: "Service broker catalog is invalid: \nService overview-service\n  Service dashboard client id must be unique\n",
							}))
						})
					})

					When("the service broker rejected the request", func() {
						BeforeEach(func() {
							serverResponse = `{
								"description": "The service broker rejected the request to https://broker.example.com/v2/service_instances/1a3794e9-7ddf-4cae-b66a-6a0453c85a3e?accepts_incomplete=true. Status Code: 400 Bad Request, Body: instance requires property \"name\"",
								"error_code": "CF-ServiceBrokerRequestRejected",
								"code": 10001
							}`
						})

						It("returns a ServiceBrokerRequestRejectedError", func() {
							Expect(executeErr).To(MatchError(ccerror.ServiceBrokerRequestRejectedError{
								Message: `The service broker rejected the request to https://broker.example.com/v2/service_instances/1a3794e9-7ddf-4cae-b66a-6a0453c85a3e?accepts_incomplete=true. Status Code: 400 Bad Request, Body: instance requires property "name"`,
							}))
						})
					})

					When("the service broker responded with bad response", func() {
						BeforeEach(func() {
							serverResponse = `{
								"description": "Service broker error",
									"error_code": "CF-ServiceBrokerBadResponse",
									"code": 10001,
									"http": {
										"uri": "https://broker.url/v2/service_instances/593fe03a-3eda-4a53-93ed-aa30f309b120?accepts_incomplete=true",
										"method": "PUT",
										"status": 500
									}
								}`
						})

						It("returns a ServiceBrokerBadResponseError", func() {
							Expect(executeErr).To(MatchError(ccerror.ServiceBrokerBadResponseError{
								Message: `Service broker error`,
							}))
						})
					})

					When("the error_code is unknown", func() {
						It("returns a V2UnexpectedResponseError with no json", func() {
							Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
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

				Context("unhandled error codes", func() {
					It("returns a V2UnexpectedResponseError with no json", func() {
						Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
							ResponseCode: http.StatusInternalServerError,
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
})
