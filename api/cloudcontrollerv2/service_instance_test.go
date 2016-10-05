package cloudcontrollerv2_test

import (
	"net/http"

	. "code.cloudfoundry.org/cli/api/cloudcontrollerv2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Service Instance", func() {
	var client *CloudControllerClient

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetServiceInstances", func() {
		BeforeEach(func() {
			response1 := `{
				"next_url": "/v2/service_instances?q=space_guid:some-space-guid&page=2",
				"resources": [
					{
						"metadata": {
							"guid": "some-service-guid-1"
						},
						"entity": {
							"name": "some-service-name-1"
						}
					},
					{
						"metadata": {
							"guid": "some-service-guid-2"
						},
						"entity": {
							"name": "some-service-name-2"
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
							"name": "some-service-name-3"
						}
					},
					{
						"metadata": {
							"guid": "some-service-guid-4"
						},
						"entity": {
							"name": "some-service-name-4"
						}
					}
				]
			}`
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest("GET", "/v2/service_instances", "q=space_guid:some-space-guid"),
					RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
				),
			)
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest("GET", "/v2/service_instances", "q=space_guid:some-space-guid&page=2"),
					RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
				),
			)
		})

		Context("when service instances exist", func() {
			It("returns all the queried service instances", func() {
				serviceInstances, warnings, err := client.GetServiceInstances([]Query{{
					Filter:   SpaceGUIDFilter,
					Operator: EqualOperator,
					Value:    "some-space-guid",
				}})
				Expect(err).NotTo(HaveOccurred())
				Expect(serviceInstances).To(ConsistOf([]ServiceInstance{
					{Name: "some-service-name-1", GUID: "some-service-guid-1"},
					{Name: "some-service-name-2", GUID: "some-service-guid-2"},
					{Name: "some-service-name-3", GUID: "some-service-guid-3"},
					{Name: "some-service-name-4", GUID: "some-service-guid-4"},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})
	})
})
