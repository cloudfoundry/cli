package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Service Broker", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetServiceBroker", func() {
		var (
			serviceBroker ServiceBroker
			warnings      Warnings
			executeErr    error
		)

		JustBeforeEach(func() {
			serviceBroker, warnings, executeErr = client.GetServiceBroker("broker-guid")
		})

		When("the cc returns back a service broker", func() {
			BeforeEach(func() {
				response := `
					{
						"metadata": {
							"guid": "broker-guid"
						},
						"entity": {
							"name":"some-broker-name"
						}
					}
				`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_brokers/broker-guid"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the service broker", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(serviceBroker).To(Equal(ServiceBroker{
					GUID: "broker-guid",
					Name: "some-broker-name",
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})

		})

		When("the cc returns 404 not found", func() {
			BeforeEach(func() {
				response := `{
					"description": "The broker is not found.",
					"error_code": "CF-BrokenNotFound",
					"code": 90004
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_brokers/broker-guid"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns an error and warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.ResourceNotFoundError{Message: "The broker is not found."}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("the cc returns error", func() {
			BeforeEach(func() {
				response := `{
					"description": "The broker is broken.",
					"error_code": "CF-BrokenBroker",
					"code": 90003
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_brokers/broker-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns an error and warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        90003,
						Description: "The broker is broken.",
						ErrorCode:   "CF-BrokenBroker",
					},
					ResponseCode: http.StatusTeapot,
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("GetServiceBrokers", func() {
		var (
			serviceBrokers []ServiceBroker
			warnings       Warnings
			executeErr     error
		)

		JustBeforeEach(func() {
			serviceBrokers, warnings, executeErr = client.GetServiceBrokers(Filter{
				Type:     constant.NameFilter,
				Operator: constant.EqualOperator,
				Values:   []string{"some-name"},
			})
		})

		When("the cc returns back service brokers", func() {
			BeforeEach(func() {
				response1 := `{
				"next_url": "/v2/service_brokers?q=name:some-name&page=2",
				"resources": [
					{
						"metadata": {
							"guid": "service-broker-guid-1"
						},
						"entity": {
							"name":"some-broker-name"
						}
					},
					{
						"metadata": {
							"guid": "service-broker-guid-2"
						},
						"entity": {
							"name":"other-broker-name"
						}
					}
				]
			}`
				response2 := `{
				"next_url": null,
				"resources": [
					{
						"metadata": {
							"guid": "service-broker-guid-3"
						},
						"entity": {
							"name":"some-broker-name"
						}
					},
					{
						"metadata": {
							"guid": "service-broker-guid-4"
						},
						"entity": {
							"name":"other-broker-name"
						}
					}
				]
			}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_brokers", "q=name:some-name"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_brokers", "q=name:some-name&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns all the queried service brokers", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(serviceBrokers).To(ConsistOf([]ServiceBroker{
					{GUID: "service-broker-guid-1", Name: "some-broker-name"},
					{GUID: "service-broker-guid-2", Name: "other-broker-name"},
					{GUID: "service-broker-guid-3", Name: "some-broker-name"},
					{GUID: "service-broker-guid-4", Name: "other-broker-name"},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})

		When("the cc returns an error", func() {
			BeforeEach(func() {
				response := `{
					"description": "The broker is broken.",
					"error_code": "CF-BrokenBroker",
					"code": 90003
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_brokers"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns an error and warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        90003,
						Description: "The broker is broken.",
						ErrorCode:   "CF-BrokenBroker",
					},
					ResponseCode: http.StatusTeapot,
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("CreateServiceBroker", func() {
		var (
			serviceBroker ServiceBroker
			warnings      Warnings
			executeErr    error
			spaceGUID     string
		)

		JustBeforeEach(func() {
			serviceBroker, warnings, executeErr = client.CreateServiceBroker(
				"broker-name", "username", "password", "https://broker.com", spaceGUID,
			)
		})

		Context("with a spaceGUID", func() {
			When("service broker creation is successful", func() {
				BeforeEach(func() {
					spaceGUID = "a-space-guid"
					response := `{
					"metadata": {
						"guid": "service-broker-guid",
						"created_at": "2016-06-08T16:41:22Z",
						"updated_at": "2016-06-08T16:41:26Z",
						"url": "/v2/service_brokers/36931aaf-62a7-4019-a708-0e9abf7e7a8f"
					},
					"entity": {
						"name": "broker-name",
						"broker_url": "https://broker.com",
						"auth_username": "username",
						"space_guid": "a-space-guid"
					}
				}`

					requestBody := map[string]interface{}{
						"name":          "broker-name",
						"broker_url":    "https://broker.com",
						"auth_username": "username",
						"auth_password": "password",
						"space_guid":    spaceGUID,
					}

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v2/service_brokers"),
							VerifyJSONRepresenting(requestBody),
							RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"a-warning,another-warning"}}),
						),
					)
				})

				It("returns the service broker and all warnings", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf(Warnings{"a-warning", "another-warning"}))
					Expect(serviceBroker).To(Equal(ServiceBroker{
						GUID:         "service-broker-guid",
						Name:         "broker-name",
						BrokerURL:    "https://broker.com",
						AuthUsername: "username",
						SpaceGUID:    spaceGUID,
					}))
				})
			})
		})

		Context("without a spaceGUID", func() {
			When("service broker creation is successful", func() {
				BeforeEach(func() {
					spaceGUID = ""
					response := `{
					"metadata": {
						"guid": "service-broker-guid",
						"created_at": "2016-06-08T16:41:22Z",
						"updated_at": "2016-06-08T16:41:26Z",
						"url": "/v2/service_brokers/36931aaf-62a7-4019-a708-0e9abf7e7a8f"
					},
					"entity": {
						"name": "broker-name",
						"broker_url": "https://broker.com",
						"auth_username": "username",
						"space_guid": ""
					}
				}`

					requestBody := map[string]interface{}{
						"name":          "broker-name",
						"broker_url":    "https://broker.com",
						"auth_username": "username",
						"auth_password": "password",
					}

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v2/service_brokers"),
							VerifyJSONRepresenting(requestBody),
							RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"a-warning"}}),
						),
					)
				})

				It("returns the service broker and all warnings", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("a-warning"))
					Expect(serviceBroker).To(Equal(ServiceBroker{
						GUID:         "service-broker-guid",
						Name:         "broker-name",
						BrokerURL:    "https://broker.com",
						AuthUsername: "username",
						SpaceGUID:    spaceGUID,
					}))
				})
			})
		})

		When("an error is returned", func() {
			BeforeEach(func() {
				response := `{
					"code": 10001,
					"description": "I'm a teapot",
					"error_code": "CF-TeapotError"
				}`

				server.AppendHandlers(
					CombineHandlers(
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"a-warning, another-warning"}}),
					),
				)
			})

			It("returns an empty service broker with errors and warnings", func() {
				Expect(serviceBroker).To(Equal(ServiceBroker{}))
				Expect(warnings).To(ConsistOf(Warnings{"a-warning", "another-warning"}))
				Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Description: "I'm a teapot",
						ErrorCode:   "CF-TeapotError",
						Code:        10001,
					},
					RequestIDs:   nil,
					ResponseCode: http.StatusTeapot,
				}))
			})
		})
	})
})
