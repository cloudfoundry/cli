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

var _ = Describe("RouteMappings", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetRouteMapping", func() {
		var (
			routeMapping RouteMapping
			warnings     Warnings
			executeErr   error
		)

		JustBeforeEach(func() {
			routeMapping, warnings, executeErr = client.GetRouteMapping("some-route-mapping-guid")
		})

		When("the cc returns an error", func() {
			BeforeEach(func() {
				response := `{
					"code": 1,
					"description": "some error description",
					"error_code": "CF-SomeError"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/route_mappings/some-route-mapping-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					),
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        1,
						Description: "some error description",
						ErrorCode:   "CF-SomeError",
					},
					ResponseCode: http.StatusTeapot,
				}))

				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		When("there are no errors", func() {
			BeforeEach(func() {
				response := `{
						"metadata": {
							"guid": "route-mapping-guid-1",
							"updated_at": null
						},
						"entity": {
							"app_port": 8888,
							"app_guid": "some-app-guid-1",
							"route_guid": "some-route-guid-1",
							"app_url": "/v2/apps/some-app-guid-1",
							"route_url": "/v2/routes/some-route-guid-1"
						}
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/route_mappings/some-route-mapping-guid"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					),
				)
			})

			It("returns the route mapping", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(routeMapping).To(Equal(RouteMapping{
					GUID:      "route-mapping-guid-1",
					AppGUID:   "some-app-guid-1",
					RouteGUID: "some-route-guid-1",
				}))
			})
		})
	})

	Describe("GetRouteMappings", func() {
		When("there are routes", func() {
			BeforeEach(func() {
				response1 := `{
				"next_url": "/v2/route_mappings?q=organization_guid:some-org-guid&page=2",
				"resources": [
					{
						"metadata": {
							"guid": "route-mapping-guid-1",
							"updated_at": null
						},
						"entity": {
							"app_port": 8888,
							"app_guid": "some-app-guid-1",
							"route_guid": "some-route-guid-1",
							"app_url": "/v2/apps/some-app-guid-1",
							"route_url": "/v2/routes/some-route-guid-1"
						}
					},
					{
						"metadata": {
							"guid": "route-mapping-guid-2",
							"updated_at": null
						},
						"entity": {
							"app_port": 8888,
							"app_guid": "some-app-guid-2",
							"route_guid": "some-route-guid-2",
							"app_url": "/v2/apps/some-app-guid-2",
							"route_url": "/v2/routes/some-route-guid-2"
						}
					}
				]
			}`
				response2 := `{
				"next_url": null,
				"resources": [
					{
						"metadata": {
							"guid": "route-mapping-guid-3",
							"updated_at": null
						},
						"entity": {
							"app_port": 8888,
							"app_guid": "some-app-guid-3",
							"route_guid": "some-route-guid-3",
							"app_url": "/v2/apps/some-app-guid-3",
							"route_url": "/v2/routes/some-route-guid-3"
						}
					},
					{
						"metadata": {
							"guid": "route-mapping-guid-4",
							"updated_at": null
						},
						"entity": {
							"app_port": 8888,
							"app_guid": "some-app-guid-4",
							"route_guid": "some-route-guid-4",
							"app_url": "/v2/apps/some-app-guid-4",
							"route_url": "/v2/routes/some-route-guid-4"
						}
					}
				]
			}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/route_mappings", "q=organization_guid:some-org-guid"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/route_mappings", "q=organization_guid:some-org-guid&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns all the route mappings and all warnings", func() {
				routeMappings, warnings, err := client.GetRouteMappings(Filter{
					Type:     constant.OrganizationGUIDFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"some-org-guid"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(routeMappings).To(ConsistOf([]RouteMapping{
					{
						GUID:      "route-mapping-guid-1",
						AppGUID:   "some-app-guid-1",
						RouteGUID: "some-route-guid-1",
					},
					{
						GUID:      "route-mapping-guid-2",
						AppGUID:   "some-app-guid-2",
						RouteGUID: "some-route-guid-2",
					},
					{
						GUID:      "route-mapping-guid-3",
						AppGUID:   "some-app-guid-3",
						RouteGUID: "some-route-guid-3",
					},
					{
						GUID:      "route-mapping-guid-4",
						AppGUID:   "some-app-guid-4",
						RouteGUID: "some-route-guid-4",
					},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})

		When("the cc returns an error", func() {
			BeforeEach(func() {
				response := `{
					"code": 10001,
					"description": "Some Error",
					"error_code": "CF-SomeError"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/route_mappings"), RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning, this is another warning"}})))
			})

			It("returns an error", func() {
				_, warnings, err := client.GetRouteMappings()
				Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))

				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})
	})
})
