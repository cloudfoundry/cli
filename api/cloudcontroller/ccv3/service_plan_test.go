package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Service Plan", func() {
	var client *Client
	var query []Query

	BeforeEach(func() {
		client, _ = NewTestClient()
		query = []Query{}
	})

	Describe("GetServicePlans", func() {
		var (
			plans      []ServicePlan
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			plans, warnings, executeErr = client.GetServicePlans(query...)
		})

		When("service plans exist", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`
					{
						"pagination": {
							"next": {
								"href": "%s/v3/service_plans?names=myServicePlan&service_broker_names=myServiceBroker&service_offering_names=someOffering&page=2"
							}
						},
						"resources": [
							{
								"guid": "service-plan-1-guid",
								"name": "service-plan-1-name",
								"description": "service-plan-1-description",
								"free": true,
								"visibility_type": "public",
								"relationships": {
									"service_offering": {
									   "data": {
										  "guid": "79d428b9-75b4-44db-addf-19c85c7f0f1e"
									   }
									}
								}
							},
							{
								"guid": "service-plan-2-guid",
								"name": "service-plan-2-name",
								"visibility_type": "admin",
								"description": "service-plan-2-description",
								"free": false,
								"relationships": {
									"service_offering": {
									   "data": {
										  "guid": "69d428b9-75b4-44db-addf-19c85c7f0f1e"
									   }
									}
								}
							}
						]
					}`,
					server.URL())

				response2 := `
					{
						"pagination": {
							"next": {
								"href": null
							}
						},
						"resources": [
							{
								"guid": "service-plan-3-guid",
								"name": "service-plan-3-name",
								"visibility_type": "organization",
								"description": "service-plan-3-description",
								"free": true,
								"relationships": {
									"service_offering": {
									   "data": {
										  "guid": "59d428b9-75b4-44db-addf-19c85c7f0f1e"
									   }
									}
								}
							}
						]
					}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_plans", "names=myServicePlan&service_broker_names=myServiceBroker&service_offering_names=someOffering"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_plans", "names=myServicePlan&service_broker_names=myServiceBroker&service_offering_names=someOffering&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
					),
				)

				query = []Query{
					{
						Key:    NameFilter,
						Values: []string{"myServicePlan"},
					},
					{
						Key:    ServiceBrokerNamesFilter,
						Values: []string{"myServiceBroker"},
					},
					{
						Key:    ServiceOfferingNamesFilter,
						Values: []string{"someOffering"},
					},
				}
			})

			It("returns a list of service plans with their associated warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(plans).To(ConsistOf(
					ServicePlan{
						GUID:                "service-plan-1-guid",
						Name:                "service-plan-1-name",
						Description:         "service-plan-1-description",
						Free:                true,
						VisibilityType:      "public",
						ServiceOfferingGUID: "79d428b9-75b4-44db-addf-19c85c7f0f1e",
					},
					ServicePlan{
						GUID:                "service-plan-2-guid",
						Name:                "service-plan-2-name",
						Description:         "service-plan-2-description",
						Free:                false,
						VisibilityType:      "admin",
						ServiceOfferingGUID: "69d428b9-75b4-44db-addf-19c85c7f0f1e",
					},
					ServicePlan{
						GUID:                "service-plan-3-guid",
						Name:                "service-plan-3-name",
						Description:         "service-plan-3-description",
						Free:                true,
						VisibilityType:      "organization",
						ServiceOfferingGUID: "59d428b9-75b4-44db-addf-19c85c7f0f1e",
					},
				))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		When("the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"code": 42424,
							"detail": "Some detailed error message",
							"title": "CF-SomeErrorTitle"
						},
						{
							"code": 11111,
							"detail": "Some other detailed error message",
							"title": "CF-SomeOtherErrorTitle"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_plans"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   42424,
							Detail: "Some detailed error message",
							Title:  "CF-SomeErrorTitle",
						},
						{
							Code:   11111,
							Detail: "Some other detailed error message",
							Title:  "CF-SomeOtherErrorTitle",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("GetServicePlansWithSpaceAndOrganization", func() {
		var (
			plans      []ServicePlanWithSpaceAndOrganization
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			plans, warnings, executeErr = client.GetServicePlansWithSpaceAndOrganization(query...)
		})

		When("when the query succeeds", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`
					{
						"pagination": {
							"next": {
								"href": "%s/v3/service_plans?include=space.organization&service_offering_names=someOffering&page=2"
							}
						},
						"resources": [
							{
								"guid": "service-plan-1-guid",
								"name": "service-plan-1-name",
								"visibility_type": "public",
								"relationships": {
									"service_offering": {
									   "data": {
										  "guid": "79d428b9-75b4-44db-addf-19c85c7f0f1e"
									   }
									}
								}
							},
							{
								"guid": "service-plan-2-guid",
								"name": "service-plan-2-name",
								"visibility_type": "space",
								"relationships": {
									"service_offering": {
									   "data": {
										  "guid": "69d428b9-75b4-44db-addf-19c85c7f0f1e"
									   }
									},
									"space": {
										"data": {
											"guid": "fake-space-guid"
										}
									}
								}
							}
						],
						"included": {
							"spaces": [
								{
									"name": "matching-space",
									"guid": "fake-space-guid",
									"relationships": {
										"organization": {
											"data": {
												"guid": "matching-org-guid"
											}
										}
									}
								},
								{
									"name": "non-matching-space",
									"guid": "fake-other-space-guid",
									"relationships": {
										"organization": {
											"data": {
												"guid": "other-org-guid"
											}
										}
									}
								}
							],
							"organizations": [
								{
									"name": "matching-org",
									"guid": "matching-org-guid"
								},
								{
									"name": "non-matching-org",
									"guid": "other-org-guid"
								}
							]
						}
					}`,
					server.URL())

				response2 := `
					{
						"pagination": {
							"next": {
								"href": null
							}
						},
						"resources": [
							{
								"guid": "service-plan-3-guid",
								"name": "service-plan-3-name",
								"visibility_type": "organization",
								"relationships": {
									"service_offering": {
									   "data": {
										  "guid": "59d428b9-75b4-44db-addf-19c85c7f0f1e"
									   }
									}
								}
							}
						]
					}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_plans", "include=space.organization&service_offering_names=someOffering"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_plans", "include=space.organization&service_offering_names=someOffering&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
					),
				)

				query = []Query{
					{
						Key:    ServiceOfferingNamesFilter,
						Values: []string{"someOffering"},
					},
				}
			})

			It("returns space and org name for space-scoped plans", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(plans).To(ConsistOf(
					ServicePlanWithSpaceAndOrganization{
						GUID:                "service-plan-1-guid",
						Name:                "service-plan-1-name",
						VisibilityType:      "public",
						ServiceOfferingGUID: "79d428b9-75b4-44db-addf-19c85c7f0f1e",
					},
					ServicePlanWithSpaceAndOrganization{
						GUID:                "service-plan-2-guid",
						Name:                "service-plan-2-name",
						VisibilityType:      "space",
						ServiceOfferingGUID: "69d428b9-75b4-44db-addf-19c85c7f0f1e",
						SpaceGUID:           "fake-space-guid",
						SpaceName:           "matching-space",
						OrganizationName:    "matching-org",
					},
					ServicePlanWithSpaceAndOrganization{
						GUID:                "service-plan-3-guid",
						Name:                "service-plan-3-name",
						VisibilityType:      "organization",
						ServiceOfferingGUID: "59d428b9-75b4-44db-addf-19c85c7f0f1e",
					},
				))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		When("the query fails", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"code": 42424,
							"detail": "Some detailed error message",
							"title": "CF-SomeErrorTitle"
						},
						{
							"code": 11111,
							"detail": "Some other detailed error message",
							"title": "CF-SomeOtherErrorTitle"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_plans", "include=space.organization"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   42424,
							Detail: "Some detailed error message",
							Title:  "CF-SomeErrorTitle",
						},
						{
							Code:   11111,
							Detail: "Some other detailed error message",
							Title:  "CF-SomeOtherErrorTitle",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("GetServicePlansWithOfferings", func() {
		var (
			offerings  []ServiceOfferingWithPlans
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			offerings, warnings, executeErr = client.GetServicePlansWithOfferings(query...)
		})

		When("when the query succeeds", func() {
			BeforeEach(func() {

				response1Template := `
					{
						"pagination": {
							"next": {
								"href": "%s/v3/service_plans?include=service_offering&space_guids=some-space-guid&organization_guids=some-org-guid&fields[service_offering.service_broker]=name,guid&page=2"
							}
						},
						"resources": [
							{
								"guid": "service-plan-1-guid",
								"name": "service-plan-1-name",
								"description": "service-plan-1-description",
								"free": true,
								"relationships": {
									"service_offering": {
									   "data": {
										  "guid": "79d428b9-75b4-44db-addf-19c85c7f0f1e"
									   }
									}
								}
							},
							{
								"guid": "service-plan-2-guid",
								"name": "service-plan-2-name",
								"description": "service-plan-2-description",
								"free": false,
								"relationships": {
									"service_offering": {
									   "data": {
										  "guid": "69d428b9-75b4-44db-addf-19c85c7f0f1e"
									   }
									}
								}
							}
						],
						"included": {
							"service_offerings": [
								{
									"name": "service-offering-1",
									"guid": "79d428b9-75b4-44db-addf-19c85c7f0f1e",
									"description": "something about service offering 1",
									"relationships": {
										"service_broker": {
											"data": {
												"guid": "service-broker-1-guid"
											}
										}
									}
								},
								{
									"name": "service-offering-2",
									"guid": "69d428b9-75b4-44db-addf-19c85c7f0f1e",
									"description": "something about service offering 2",
									"relationships": {
										"service_broker": {
											"data": {
												"guid": "service-broker-2-guid"
											}
										}
									}
								}
							],
							"service_brokers": [
								{
									"name": "service-broker-1",
									"guid": "service-broker-1-guid"
								},
								{
									"name": "service-broker-2",
									"guid": "service-broker-2-guid"
								}
							]
						}
					}`

				response1 := fmt.Sprintf(response1Template, server.URL())
				response2 := `
					{
						"pagination": {
							"next": {
								"href": null
							}
						},
						"resources": [
							{
								"guid": "service-plan-3-guid",
								"name": "service-plan-3-name",
								"description": "service-plan-3-description",
								"free": true,
								"relationships": {
									"service_offering": {
									   "data": {
										  "guid": "79d428b9-75b4-44db-addf-19c85c7f0f1e"
									   }
									}
								}
							}
						],
						"included": {
							"service_offerings": [
								{
									"name": "service-offering-1",
									"guid": "79d428b9-75b4-44db-addf-19c85c7f0f1e",
									"description": "something about service offering 1",
									"relationships": {
										"service_broker": {
											"data": {
												"guid": "service-broker-1-guid"
											}
										}
									}
								}
							],
							"service_brokers": [
								{
									"name": "service-broker-1",
									"guid": "service-broker-1-guid"
								}
							]
						}
					}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_plans", "include=service_offering&space_guids=some-space-guid&organization_guids=some-org-guid&fields[service_offering.service_broker]=name,guid"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_plans", "include=service_offering&space_guids=some-space-guid&organization_guids=some-org-guid&fields[service_offering.service_broker]=name,guid&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
					),
				)

				query = []Query{
					{
						Key:    OrganizationGUIDFilter,
						Values: []string{"some-org-guid"},
					},
					{
						Key:    SpaceGUIDFilter,
						Values: []string{"some-space-guid"},
					},
				}
			})

			It("returns service offerings and service plans", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(offerings).To(Equal([]ServiceOfferingWithPlans{
					{
						GUID:              "79d428b9-75b4-44db-addf-19c85c7f0f1e",
						Name:              "service-offering-1",
						Description:       "something about service offering 1",
						ServiceBrokerName: "service-broker-1",
						Plans: []ServicePlan{
							{
								GUID:                "service-plan-1-guid",
								Name:                "service-plan-1-name",
								Description:         "service-plan-1-description",
								Free:                true,
								ServiceOfferingGUID: "79d428b9-75b4-44db-addf-19c85c7f0f1e",
							},
							{
								GUID:                "service-plan-3-guid",
								Name:                "service-plan-3-name",
								Description:         "service-plan-3-description",
								Free:                true,
								ServiceOfferingGUID: "79d428b9-75b4-44db-addf-19c85c7f0f1e",
							},
						},
					},
					{
						GUID:              "69d428b9-75b4-44db-addf-19c85c7f0f1e",
						Name:              "service-offering-2",
						Description:       "something about service offering 2",
						ServiceBrokerName: "service-broker-2",
						Plans: []ServicePlan{
							{
								GUID:                "service-plan-2-guid",
								Name:                "service-plan-2-name",
								Description:         "service-plan-2-description",
								Free:                false,
								ServiceOfferingGUID: "69d428b9-75b4-44db-addf-19c85c7f0f1e",
							},
						},
					},
				}))
			})
		})

		When("the query fails", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"code": 42424,
							"detail": "Some detailed error message",
							"title": "CF-SomeErrorTitle"
						},
						{
							"code": 11111,
							"detail": "Some other detailed error message",
							"title": "CF-SomeOtherErrorTitle"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_plans", "include=service_offering&fields[service_offering.service_broker]=name,guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   42424,
							Detail: "Some detailed error message",
							Title:  "CF-SomeErrorTitle",
						},
						{
							Code:   11111,
							Detail: "Some other detailed error message",
							Title:  "CF-SomeOtherErrorTitle",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

	})
})
