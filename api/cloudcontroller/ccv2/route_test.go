package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Route", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetSpaceRoutes", func() {
		Context("when there are routes in this space", func() {
			BeforeEach(func() {
				response1 := `{
				"next_url": "/v2/spaces/some-space-guid/routes?q=space_guid:some-space-guid&page=2",
				"resources": [
					{
						"metadata": {
							"guid": "route-guid-1",
							"updated_at": null
						},
						"entity": {
							"host": "host-1",
							"path": "path",
							"port": null,
							"domain_guid": "some-http-domain"
						}
					},
					{
						"metadata": {
							"guid": "route-guid-2",
							"updated_at": null
						},
						"entity": {
							"host": "host-2",
							"path": "",
							"port": 3333,
							"domain_guid": "some-tcp-domain"
						}
					}
				]
			}`
				response2 := `{
				"next_url": null,
				"resources": [
					{
						"metadata": {
							"guid": "route-guid-3",
							"updated_at": null
						},
						"entity": {
							"host": "host-3",
							"path": "path",
							"port": null,
							"domain_guid": "some-http-domain"
						}
					},
					{
						"metadata": {
							"guid": "route-guid-4",
							"updated_at": null
						},
						"entity": {
							"host": "host-4",
							"path": "",
							"port": 333,
							"domain_guid": "some-tcp-domain"
						}
					}
				]
			}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/spaces/some-space-guid/routes", "q=space_guid:some-space-guid"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/spaces/some-space-guid/routes", "q=space_guid:some-space-guid&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns all the routes and all warnings", func() {
				routes, warnings, err := client.GetSpaceRoutes("some-space-guid", []Query{{
					Filter:   SpaceGUIDFilter,
					Operator: EqualOperator,
					Value:    "some-space-guid",
				}})
				Expect(err).NotTo(HaveOccurred())
				Expect(routes).To(ConsistOf([]Route{
					{
						GUID:       "route-guid-1",
						Host:       "host-1",
						Path:       "path",
						Port:       0,
						DomainGUID: "some-http-domain",
					},
					{
						GUID:       "route-guid-2",
						Host:       "host-2",
						Path:       "",
						Port:       3333,
						DomainGUID: "some-tcp-domain",
					},
					{
						GUID:       "route-guid-3",
						Host:       "host-3",
						Path:       "path",
						Port:       0,
						DomainGUID: "some-http-domain",
					},
					{
						GUID:       "route-guid-4",
						Host:       "host-4",
						Path:       "",
						Port:       333,
						DomainGUID: "some-tcp-domain",
					},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})

		Context("when there are no routes in this space", func() {
			BeforeEach(func() {
				response := `{
				"next_url": "",
				"resources": []
			}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/spaces/some-space-guid/routes"),
						RespondWith(http.StatusOK, response),
					),
				)
			})

			It("returns an empty list of routes", func() {
				routes, _, err := client.GetSpaceRoutes("some-space-guid", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(routes).To(BeEmpty())
			})
		})

		Context("when the space is not found", func() {
			BeforeEach(func() {
				response := `{
					"code": 40004,
					"description": "The app space could not be found: some-space-guid",
					"error_code": "CF-SpaceNotFound"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/spaces/some-space-guid/routes"),
						RespondWith(http.StatusNotFound, response),
					),
				)
			})

			It("returns an error", func() {
				routes, _, err := client.GetSpaceRoutes("some-space-guid", nil)
				Expect(err).To(MatchError(cloudcontroller.ResourceNotFoundError{
					Message: "The app space could not be found: some-space-guid",
				}))
				Expect(routes).To(BeEmpty())
			})
		})
	})

	Describe("DeleteRoute", func() {
		Context("when the route exists", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v2/routes/some-route-guid"),
						RespondWith(http.StatusNoContent, "{}", http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("deletes the route and returns all warnings", func() {
				warnings, err := client.DeleteRoute("some-route-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})

		Context("when the route does not exist", func() {
			BeforeEach(func() {
				response := `{
				"code": 210002,
				"description": "The route could not be found: some-route-guid",
				"error_code": "CF-RouteNotFound"
			}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v2/routes/some-route-guid"),
						RespondWith(http.StatusNotFound, response),
					),
				)
			})

			It("returns an error", func() {
				_, err := client.DeleteRoute("some-route-guid")
				Expect(err).To(MatchError(cloudcontroller.ResourceNotFoundError{
					Message: "The route could not be found: some-route-guid",
				}))
			})
		})
	})
})
