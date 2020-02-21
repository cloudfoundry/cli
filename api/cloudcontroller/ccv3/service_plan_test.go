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
						VisibilityType:      "public",
						ServiceOfferingGUID: "79d428b9-75b4-44db-addf-19c85c7f0f1e",
					},
					ServicePlan{
						GUID:                "service-plan-2-guid",
						Name:                "service-plan-2-name",
						VisibilityType:      "admin",
						ServiceOfferingGUID: "69d428b9-75b4-44db-addf-19c85c7f0f1e",
					},
					ServicePlan{
						GUID:                "service-plan-3-guid",
						Name:                "service-plan-3-name",
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
})
