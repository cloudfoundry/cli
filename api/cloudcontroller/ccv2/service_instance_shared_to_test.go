package ccv2_test

import (
	"net/http"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Service Instance Shared To", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetServiceInstanceSharedTos", func() {
		Context("when the cc api returns a valid response", func() {
			BeforeEach(func() {
				response1 := `{
				"total_results": 3,
				"total_pages": 2,
				"prev_url": null,
				"next_url": "/v2/service_instances/some-service-instance-guid/shared_to?page=2",
				"resources": [{
						"space_guid": "some-space-guid",
						"space_name": "some-space-name",
						"organization_name": "some-org-name"
					},
					{
						"space_guid": "some-space-guid-2",
						"space_name": "some-space-name-2",
						"organization_name": "some-org-name-2"
					}
				]
			}`

				response2 := `{
				"total_results": 3,
				"total_pages": 2,
				"prev_url": "/v2/service_instances/some-service-instance-guid/shared_to?page=1",
				"next_url": null,
				"resources": [{
					"space_guid": "some-space-guid-3",
					"space_name": "some-space-name-3",
					"organization_name": "some-org-name-3"
				}]
			}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_instances/some-service-instance-guid/shared_to"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_instances/some-service-instance-guid/shared_to", "page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			Context("when the service instance exist", func() {
				It("returns all the shared-to resources", func() {
					serviceInstances, warnings, err := client.GetServiceInstanceSharedTos("some-service-instance-guid")
					Expect(err).NotTo(HaveOccurred())

					Expect(serviceInstances).To(ConsistOf([]ServiceInstanceSharedTo{
						{SpaceGUID: "some-space-guid", SpaceName: "some-space-name", OrganizationName: "some-org-name"},
						{SpaceGUID: "some-space-guid-2", SpaceName: "some-space-name-2", OrganizationName: "some-org-name-2"},
						{SpaceGUID: "some-space-guid-3", SpaceName: "some-space-name-3", OrganizationName: "some-org-name-3"},
					}))
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
				})
			})
		})

		Context("when the cc api returns an invalid response", func() {
			BeforeEach(func() {
				invalidResponse := `{"foo": "bar"}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_instances/some-service-instance-guid/shared_to"),
						RespondWith(http.StatusOK, invalidResponse, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns an error and warnings", func() {
				_, warnings, err := client.GetServiceInstanceSharedTos("some-service-instance-guid")
				Expect(err).To(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})
})
