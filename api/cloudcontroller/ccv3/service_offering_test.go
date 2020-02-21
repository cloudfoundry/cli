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

var _ = Describe("Service Offering", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("GetServiceOfferings", func() {
		var (
			query []Query

			offerings  []ServiceOffering
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			offerings, warnings, executeErr = client.GetServiceOfferings(query...)
		})

		When("service offerings exist", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`
					{
						"pagination": {
							"next": {
								"href": "%s/v3/service_offerings?names=myServiceOffering&service_broker_names=myServiceBroker&page=2"
							}
						},
						"resources": [
							{
								"guid": "service-offering-1-guid",
								"name": "service-offering-1-name",
								"relationships": {
									"service_broker": {
										"data": {
											"name": "overview-broker"
										}
									}
								}
							},
							{
								"guid": "service-offering-2-guid",
								"name": "service-offering-2-name",
								"relationships": {
									"service_broker": {
										"data": {
											"name": "overview-broker"
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
								"guid": "service-offering-3-guid",
								"name": "service-offering-3-name",
								"relationships": {
									"service_broker": {
										"data": {
											"name": "other-broker"
										}
									}
								}
							}
						]
					}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_offerings", "names=myServiceOffering&service_broker_names=myServiceBroker"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_offerings", "names=myServiceOffering&service_broker_names=myServiceBroker&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
					),
				)

				query = []Query{
					{
						Key:    NameFilter,
						Values: []string{"myServiceOffering"},
					},
					{
						Key:    ServiceBrokerNamesFilter,
						Values: []string{"myServiceBroker"},
					},
				}
			})

			It("returns a list of service offerings with their associated warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(offerings).To(ConsistOf(
					ServiceOffering{
						GUID:              "service-offering-1-guid",
						Name:              "service-offering-1-name",
						ServiceBrokerName: "overview-broker",
					},
					ServiceOffering{
						GUID:              "service-offering-2-guid",
						Name:              "service-offering-2-name",
						ServiceBrokerName: "overview-broker",
					},
					ServiceOffering{
						GUID:              "service-offering-3-guid",
						Name:              "service-offering-3-name",
						ServiceBrokerName: "other-broker",
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
						VerifyRequest(http.MethodGet, "/v3/service_offerings"),
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
