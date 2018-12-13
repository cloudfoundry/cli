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

var _ = Describe("Service", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetService", func() {
		When("the service exists", func() {
			When("the value of the 'extra' json key is non-empty", func() {
				BeforeEach(func() {
					response := `{
						"metadata": {
							"guid": "some-service-guid"
						},
						"entity": {
							"label": "some-service",
							"description": "some-description",
							"extra": "{\"provider\":{\"name\":\"The name\"},\"listing\":{\"imageUrl\":\"http://catgifpage.com/cat.gif\",\"blurb\":\"fake broker that is fake\",\"longDescription\":\"A long time ago, in a galaxy far far away...\"},\"displayName\":\"The Fake Broker\",\"shareable\":true}"
						}
					}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/services/some-service-guid"),
							RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns the service and warnings", func() {
					service, warnings, err := client.GetService("some-service-guid")
					Expect(err).NotTo(HaveOccurred())

					Expect(service).To(Equal(Service{
						GUID:        "some-service-guid",
						Label:       "some-service",
						Description: "some-description",
						Extra: ServiceExtra{
							Shareable: true,
						},
					}))
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				})
			})

			When("the value of the 'extra' json key is null", func() {
				BeforeEach(func() {
					response := `{
						"metadata": {
							"guid": "some-service-guid"
						},
						"entity": {
							"extra": null
						}
					}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/services/some-service-guid"),
							RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns extra.shareable == 'false'", func() {
					service, _, err := client.GetService("some-service-guid")
					Expect(err).NotTo(HaveOccurred())

					Expect(service).To(Equal(Service{
						GUID:  "some-service-guid",
						Extra: ServiceExtra{Shareable: false},
					}))
				})
			})

			When("the value of the 'extra' json key is the empty string", func() {
				BeforeEach(func() {
					response := `{
						"metadata": {
							"guid": "some-service-guid"
						},
						"entity": {
							"extra": ""
						}
					}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/services/some-service-guid"),
							RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns extra.shareable == 'false'", func() {
					service, _, err := client.GetService("some-service-guid")
					Expect(err).NotTo(HaveOccurred())

					Expect(service).To(Equal(Service{
						GUID:  "some-service-guid",
						Extra: ServiceExtra{Shareable: false},
					}))
				})
			})

			When("the key 'extra' is not in the json response", func() {
				BeforeEach(func() {
					response := `{
						"metadata": {
							"guid": "some-service-guid"
						}
					}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/services/some-service-guid"),
							RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns extra.shareable == 'false'", func() {
					service, _, err := client.GetService("some-service-guid")
					Expect(err).NotTo(HaveOccurred())

					Expect(service).To(Equal(Service{
						GUID:  "some-service-guid",
						Extra: ServiceExtra{Shareable: false},
					}))
				})
			})

			When("the documentation url is set", func() {
				Context("in the entity structure", func() {
					BeforeEach(func() {
						response := `{
						"metadata": {
							"guid": "some-service-guid"
						},
						"entity": {
							"documentation_url": "some-url"
						}
					}`
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodGet, "/v2/services/some-service-guid"),
								RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
							),
						)
					})

					It("returns the documentation url correctly", func() {
						service, _, err := client.GetService("some-service-guid")
						Expect(err).NotTo(HaveOccurred())

						Expect(service).To(Equal(Service{
							GUID:             "some-service-guid",
							DocumentationURL: "some-url",
						}))
					})
				})

				Context("in the extra structure", func() {
					BeforeEach(func() {
						response := `{
						"metadata": {
							"guid": "some-service-guid"
						},
						"entity": {
							"extra": "{\"documentationUrl\":\"some-url\"}"
						}
					}`
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodGet, "/v2/services/some-service-guid"),
								RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
							),
						)
					})

					It("returns the documentation url correctly", func() {
						service, _, err := client.GetService("some-service-guid")
						Expect(err).NotTo(HaveOccurred())

						Expect(service).To(Equal(Service{
							GUID:             "some-service-guid",
							DocumentationURL: "some-url",
						}))
					})
				})

				Context("in both the entity and extra structures", func() {
					BeforeEach(func() {
						response := `{
						"metadata": {
							"guid": "some-service-guid"
						},
						"entity": {
							"documentation_url": "entity-url",
							"extra": "{\"documentationUrl\":\"some-url\"}"
						}
					}`
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodGet, "/v2/services/some-service-guid"),
								RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
							),
						)
					})

					It("prioritises the entity structure", func() {
						service, _, err := client.GetService("some-service-guid")
						Expect(err).NotTo(HaveOccurred())

						Expect(service).To(Equal(Service{
							GUID:             "some-service-guid",
							DocumentationURL: "entity-url",
						}))
					})
				})
			})
		})

		When("the service does not exist (testing general error case)", func() {
			BeforeEach(func() {
				response := `{
					"description": "The service could not be found: non-existant-service-guid",
					"error_code": "CF-ServiceNotFound",
					"code": 120003
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/services/non-existant-service-guid"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					))
			})

			It("returns an error and warnings", func() {
				_, warnings, err := client.GetService("non-existant-service-guid")
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{Message: "The service could not be found: non-existant-service-guid"}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})
	Describe("GetServices", func() {
		var (
			services   []Service
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			services, warnings, executeErr = client.GetServices(Filter{
				Type:     constant.LabelFilter,
				Operator: constant.EqualOperator,
				Values:   []string{"some-label"},
			})
		})

		When("the cc returns back services", func() {
			BeforeEach(func() {
				response1 := `{
					"next_url": "/v2/services?q=label:some-label&page=2",
					"resources": [
						{
							"metadata": {
								"guid": "some-service-guid-1"
							},
							"entity": {
								"label": "some-service",
								"service_broker_name": "broker-1"
							}
						},
						{
							"metadata": {
								"guid": "some-service-guid-2"
							},
							"entity": {
								"label": "other-service",
								"service_broker_name": "broker-2"
							}
						}
					]
				}`

				response2 := `{
					"next_url": null,
					"resources": [
						{
							"metadata": {
								"guid": "some-service-guid-3"
							},
							"entity": {
								"label": "some-service",
								"service_broker_name": "broker-3"
							}
						},
						{
							"metadata": {
								"guid": "some-service-guid-4"
							},
							"entity": {
								"label": "other-service",
								"service_broker_name": "broker-4"
							}
						}
					]
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/services", "q=label:some-label"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/services", "q=label:some-label&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns all the queried services", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(services).To(ConsistOf([]Service{
					{GUID: "some-service-guid-1", Label: "some-service", ServiceBrokerName: "broker-1"},
					{GUID: "some-service-guid-2", Label: "other-service", ServiceBrokerName: "broker-2"},
					{GUID: "some-service-guid-3", Label: "some-service", ServiceBrokerName: "broker-3"},
					{GUID: "some-service-guid-4", Label: "other-service", ServiceBrokerName: "broker-4"},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})

		When("the cc returns an error", func() {
			BeforeEach(func() {
				response := `{
					"description": "Some description.",
					"error_code": "CF-Error",
					"code": 90003
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/services"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns an error and warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        90003,
						Description: "Some description.",
						ErrorCode:   "CF-Error",
					},
					ResponseCode: http.StatusTeapot,
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
	Describe("GetSpaceServices", func() {
		var (
			services   []Service
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			services, warnings, executeErr = client.GetSpaceServices("some-space-guid", Filter{
				Type:     constant.LabelFilter,
				Operator: constant.EqualOperator,
				Values:   []string{"some-label"},
			})
		})

		When("the cc returns back services", func() {
			BeforeEach(func() {
				response1 := `{
					"next_url": "/v2/spaces/some-space-guid/services?q=label:some-label&page=2",
					"resources": [
						{
							"metadata": {
								"guid": "some-service-guid-1"
							},
							"entity": {
								"label": "some-service"
							}
						},
						{
							"metadata": {
								"guid": "some-service-guid-2"
							},
							"entity": {
								"label": "other-service"
							}
						}
					]
				}`

				response2 := `{
					"next_url": null,
					"resources": [
						{
							"metadata": {
								"guid": "some-service-guid-3"
							},
							"entity": {
								"label": "some-service"
							}
						},
						{
							"metadata": {
								"guid": "some-service-guid-4"
							},
							"entity": {
								"label": "other-service"
							}
						}
					]
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/spaces/some-space-guid/services", "q=label:some-label"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/spaces/some-space-guid/services", "q=label:some-label&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns all the queried services", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(services).To(ConsistOf([]Service{
					{GUID: "some-service-guid-1", Label: "some-service"},
					{GUID: "some-service-guid-2", Label: "other-service"},
					{GUID: "some-service-guid-3", Label: "some-service"},
					{GUID: "some-service-guid-4", Label: "other-service"},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})

		When("the cc returns an error", func() {
			BeforeEach(func() {
				response := `{
					"description": "Some description.",
					"error_code": "CF-Error",
					"code": 90003
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/spaces/some-space-guid/services"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns an error and warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        90003,
						Description: "Some description.",
						ErrorCode:   "CF-Error",
					},
					ResponseCode: http.StatusTeapot,
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})
