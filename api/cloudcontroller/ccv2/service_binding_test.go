package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Service Binding", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetServiceBindings", func() {
		BeforeEach(func() {
			response1 := `{
				"next_url": "/v2/service_bindings?q=app_guid:some-app-guid&page=2",
				"resources": [
					{
						"metadata": {
							"guid": "service-binding-guid-1"
						}
					},
					{
						"metadata": {
							"guid": "service-binding-guid-2"
						}
					}
				]
			}`
			response2 := `{
				"next_url": null,
				"resources": [
					{
						"metadata": {
							"guid": "service-binding-guid-3"
						}
					},
					{
						"metadata": {
							"guid": "service-binding-guid-4"
						}
					}
				]
			}`
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v2/service_bindings", "q=app_guid:some-app-guid"),
					RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
				),
			)
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v2/service_bindings", "q=app_guid:some-app-guid&page=2"),
					RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
				),
			)
		})

		Context("when service bindings exist", func() {
			It("returns all the queried service bindings", func() {
				serviceBindings, warnings, err := client.GetServiceBindings([]Query{{
					Filter:   AppGUIDFilter,
					Operator: EqualOperator,
					Value:    "some-app-guid",
				}})
				Expect(err).NotTo(HaveOccurred())
				Expect(serviceBindings).To(ConsistOf([]ServiceBinding{
					{GUID: "service-binding-guid-1"},
					{GUID: "service-binding-guid-2"},
					{GUID: "service-binding-guid-3"},
					{GUID: "service-binding-guid-4"},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})
	})

	Describe("DeleteServiceBinding", func() {
		Context("when the service binding exist", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v2/service_bindings/some-service-binding-guid"),
						RespondWith(http.StatusNoContent, "{}", http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("deletes the service binding", func() {
				warnings, err := client.DeleteServiceBinding("some-service-binding-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})

	Context("when the service binding does not exist", func() {
		BeforeEach(func() {
			response := `{
				"code": 90004,
				"description": "The service binding could not be found: some-service-binding-guid",
				"error_code": "CF-ServiceBindingNotFound"
			}`
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodDelete, "/v2/service_bindings/some-service-binding-guid"),
					RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
				),
			)
		})

		It("returns a not found error", func() {
			warnings, err := client.DeleteServiceBinding("some-service-binding-guid")
			Expect(err).To(MatchError(cloudcontroller.ResourceNotFoundError{
				Message: "The service binding could not be found: some-service-binding-guid",
			}))
			Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
		})
	})
})
