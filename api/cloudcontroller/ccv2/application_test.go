package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Application", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetApplications", func() {
		BeforeEach(func() {
			response1 := `{
				"next_url": "/v2/apps?q=space_guid:some-space-guid&page=2",
				"resources": [
					{
						"metadata": {
							"guid": "app-guid-1",
							"updated_at": null
						},
						"entity": {
							"name": "app-name-1"
						}
					},
					{
						"metadata": {
							"guid": "app-guid-2",
							"updated_at": null
						},
						"entity": {
							"name": "app-name-2"
						}
					}
				]
			}`
			response2 := `{
				"next_url": null,
				"resources": [
					{
						"metadata": {
							"guid": "app-guid-3",
							"updated_at": null
						},
						"entity": {
							"name": "app-name-3"
						}
					},
					{
						"metadata": {
							"guid": "app-guid-4",
							"updated_at": null
						},
						"entity": {
							"name": "app-name-4"
						}
					}
				]
			}`
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v2/apps", "q=space_guid:some-space-guid"),
					RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
				),
			)
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v2/apps", "q=space_guid:some-space-guid&page=2"),
					RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
				),
			)
		})

		Context("when apps exist", func() {
			It("returns all the queried apps", func() {
				apps, warnings, err := client.GetApplications([]Query{{
					Filter:   SpaceGUIDFilter,
					Operator: EqualOperator,
					Value:    "some-space-guid",
				}})
				Expect(err).NotTo(HaveOccurred())
				Expect(apps).To(ConsistOf([]Application{
					{Name: "app-name-1", GUID: "app-guid-1"},
					{Name: "app-name-2", GUID: "app-guid-2"},
					{Name: "app-name-3", GUID: "app-guid-3"},
					{Name: "app-name-4", GUID: "app-guid-4"},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})
	})

	Describe("GetRouteApplications", func() {
		Context("when the route guid is not found", func() {
			BeforeEach(func() {
				response := `
{
  "code": 210002,
  "description": "The route could not be found: some-route-guid",
  "error_code": "CF-RouteNotFound"
}
			`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/routes/some-route-guid/apps"),
						RespondWith(http.StatusNotFound, response),
					),
				)
			})

			It("returns an error", func() {
				_, _, err := client.GetRouteApplications("some-route-guid", nil)
				Expect(err).To(MatchError(cloudcontroller.ResourceNotFoundError{
					Message: "The route could not be found: some-route-guid",
				}))
			})
		})

		Context("when there are applications associated with this route", func() {
			BeforeEach(func() {
				response1 := `{
				"next_url": "/v2/routes/some-route-guid/apps?q=space_guid:some-space-guid&page=2",
				"resources": [
					{
						"metadata": {
							"guid": "app-guid-1",
							"updated_at": null
						},
						"entity": {
							"name": "app-name-1"
						}
					},
					{
						"metadata": {
							"guid": "app-guid-2",
							"updated_at": null
						},
						"entity": {
							"name": "app-name-2"
						}
					}
				]
			}`
				response2 := `{
				"next_url": null,
				"resources": [
					{
						"metadata": {
							"guid": "app-guid-3",
							"updated_at": null
						},
						"entity": {
							"name": "app-name-3"
						}
					},
					{
						"metadata": {
							"guid": "app-guid-4",
							"updated_at": null
						},
						"entity": {
							"name": "app-name-4"
						}
					}
				]
			}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/routes/some-route-guid/apps", "q=space_guid:some-space-guid"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/routes/some-route-guid/apps", "q=space_guid:some-space-guid&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns all the applications and all warnings", func() {
				apps, warnings, err := client.GetRouteApplications("some-route-guid", []Query{{
					Filter:   SpaceGUIDFilter,
					Operator: EqualOperator,
					Value:    "some-space-guid",
				}})
				Expect(err).NotTo(HaveOccurred())
				Expect(apps).To(ConsistOf([]Application{
					{Name: "app-name-1", GUID: "app-guid-1"},
					{Name: "app-name-2", GUID: "app-guid-2"},
					{Name: "app-name-3", GUID: "app-guid-3"},
					{Name: "app-name-4", GUID: "app-guid-4"},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})

		Context("when there are no applications associated with this route", func() {
			BeforeEach(func() {
				response := `{
				"next_url": "",
				"resources": []
			}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/routes/some-route-guid/apps"),
						RespondWith(http.StatusOK, response),
					),
				)
			})

			It("returns an empty list of applications", func() {
				apps, _, err := client.GetRouteApplications("some-route-guid", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(apps).To(BeEmpty())
			})
		})
	})
})
