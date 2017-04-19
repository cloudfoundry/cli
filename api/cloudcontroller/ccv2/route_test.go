package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
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

	Describe("GetRoutes", func() {
		Context("when there are routes", func() {
			BeforeEach(func() {
				response1 := `{
				"next_url": "/v2/routes?q=organization_guid:some-org-guid&page=2",
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
							"domain_guid": "some-http-domain",
							"space_guid": "some-space-guid-1"
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
							"domain_guid": "some-tcp-domain",
							"space_guid": "some-space-guid-1"
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
							"domain_guid": "some-http-domain",
							"space_guid": "some-space-guid-2"
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
							"domain_guid": "some-tcp-domain",
							"space_guid": "some-space-guid-2"
						}
					}
				]
			}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/routes", "q=organization_guid:some-org-guid"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/routes", "q=organization_guid:some-org-guid&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns all the routes and all warnings", func() {
				routes, warnings, err := client.GetRoutes([]Query{{
					Filter:   OrganizationGUIDFilter,
					Operator: EqualOperator,
					Value:    "some-org-guid",
				}})
				Expect(err).NotTo(HaveOccurred())
				Expect(routes).To(ConsistOf([]Route{
					{
						GUID:       "route-guid-1",
						Host:       "host-1",
						Path:       "path",
						Port:       0,
						DomainGUID: "some-http-domain",
						SpaceGUID:  "some-space-guid-1",
					},
					{
						GUID:       "route-guid-2",
						Host:       "host-2",
						Path:       "",
						Port:       3333,
						DomainGUID: "some-tcp-domain",
						SpaceGUID:  "some-space-guid-1",
					},
					{
						GUID:       "route-guid-3",
						Host:       "host-3",
						Path:       "path",
						Port:       0,
						DomainGUID: "some-http-domain",
						SpaceGUID:  "some-space-guid-2",
					},
					{
						GUID:       "route-guid-4",
						Host:       "host-4",
						Path:       "",
						Port:       333,
						DomainGUID: "some-tcp-domain",
						SpaceGUID:  "some-space-guid-2",
					},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})

		Context("when the cc returns an error", func() {
			BeforeEach(func() {
				response := `{
					"code": 10001,
					"description": "Some Error",
					"error_code": "CF-SomeError"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/routes"),
						RespondWith(http.StatusTeapot, response),
					),
				)
			})

			It("returns an error", func() {
				_, _, err := client.GetRoutes(nil)
				Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
			})
		})
	})

	Describe("GetApplicationRoutes", func() {
		Context("when there are routes in this space", func() {
			BeforeEach(func() {
				response1 := `{
				"next_url": "/v2/apps/some-app-guid/routes?q=organization_guid:some-org-guid&page=2",
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
							"domain_guid": "some-http-domain",
							"space_guid": "some-space-guid-1"
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
							"domain_guid": "some-tcp-domain",
							"space_guid": "some-space-guid-1"
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
							"domain_guid": "some-http-domain",
							"space_guid": "some-space-guid-1"
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
							"domain_guid": "some-tcp-domain",
							"space_guid": "some-space-guid-1"
						}
					}
				]
			}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/apps/some-app-guid/routes", "q=organization_guid:some-org-guid"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/apps/some-app-guid/routes", "q=organization_guid:some-org-guid&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns all the routes and all warnings", func() {
				routes, warnings, err := client.GetApplicationRoutes("some-app-guid", []Query{{
					Filter:   OrganizationGUIDFilter,
					Operator: EqualOperator,
					Value:    "some-org-guid",
				}})
				Expect(err).NotTo(HaveOccurred())
				Expect(routes).To(ConsistOf([]Route{
					{
						GUID:       "route-guid-1",
						Host:       "host-1",
						Path:       "path",
						Port:       0,
						DomainGUID: "some-http-domain",
						SpaceGUID:  "some-space-guid-1",
					},
					{
						GUID:       "route-guid-2",
						Host:       "host-2",
						Path:       "",
						Port:       3333,
						DomainGUID: "some-tcp-domain",
						SpaceGUID:  "some-space-guid-1",
					},
					{
						GUID:       "route-guid-3",
						Host:       "host-3",
						Path:       "path",
						Port:       0,
						DomainGUID: "some-http-domain",
						SpaceGUID:  "some-space-guid-1",
					},
					{
						GUID:       "route-guid-4",
						Host:       "host-4",
						Path:       "",
						Port:       333,
						DomainGUID: "some-tcp-domain",
						SpaceGUID:  "some-space-guid-1",
					},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})

		Context("when there are no routes bound to the app", func() {
			BeforeEach(func() {
				response := `{
				"next_url": "",
				"resources": []
			}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/apps/some-app-guid/routes"),
						RespondWith(http.StatusOK, response),
					),
				)
			})

			It("returns an empty list of routes", func() {
				routes, _, err := client.GetApplicationRoutes("some-app-guid", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(routes).To(BeEmpty())
			})
		})

		Context("when the app is not found", func() {
			BeforeEach(func() {
				response := `{
					"code": 10000,
					"description": "The app could not be found: some-app-guid",
					"error_code": "CF-AppNotFound"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/apps/some-app-guid/routes"),
						RespondWith(http.StatusNotFound, response),
					),
				)
			})

			It("returns an error", func() {
				routes, _, err := client.GetApplicationRoutes("some-app-guid", nil)
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "The app could not be found: some-app-guid",
				}))
				Expect(routes).To(BeEmpty())
			})
		})
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
							"domain_guid": "some-http-domain",
							"space_guid": "some-space-guid-1"
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
							"domain_guid": "some-tcp-domain",
							"space_guid": "some-space-guid-1"
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
							"domain_guid": "some-http-domain",
							"space_guid": "some-space-guid-1"
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
							"domain_guid": "some-tcp-domain",
							"space_guid": "some-space-guid-1"
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
						SpaceGUID:  "some-space-guid-1",
					},
					{
						GUID:       "route-guid-2",
						Host:       "host-2",
						Path:       "",
						Port:       3333,
						DomainGUID: "some-tcp-domain",
						SpaceGUID:  "some-space-guid-1",
					},
					{
						GUID:       "route-guid-3",
						Host:       "host-3",
						Path:       "path",
						Port:       0,
						DomainGUID: "some-http-domain",
						SpaceGUID:  "some-space-guid-1",
					},
					{
						GUID:       "route-guid-4",
						Host:       "host-4",
						Path:       "",
						Port:       333,
						DomainGUID: "some-tcp-domain",
						SpaceGUID:  "some-space-guid-1",
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
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{
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
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "The route could not be found: some-route-guid",
				}))
			})
		})
	})

	Describe("CheckRoute", func() {
		Context("with minimum params", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/routes/reserved/domain/some-domain-guid"),
						RespondWith(http.StatusNoContent, "{}", http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("does not contain any params", func() {
				_, warnings, err := client.CheckRoute(Route{DomainGUID: "some-domain-guid"})
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})

		Context("with all the params", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/routes/reserved/domain/some-domain-guid", "host=some-host&path=some-path&port=42"),
						RespondWith(http.StatusNoContent, "{}", http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns true", func() {
				exists, warnings, err := client.CheckRoute(Route{
					Host:       "some-host",
					DomainGUID: "some-domain-guid",
					Path:       "some-path",
					Port:       42,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				Expect(exists).To(BeTrue())
			})
		})

		Context("when the route exists", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/routes/reserved/domain/some-domain-guid", "host=some-host"),
						RespondWith(http.StatusNoContent, "{}", http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns true", func() {
				exists, warnings, err := client.CheckRoute(Route{Host: "some-host", DomainGUID: "some-domain-guid"})
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				Expect(exists).To(BeTrue())
			})
		})

		Context("when the route does not exist", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/routes/reserved/domain/some-domain-guid", "host=some-host"),
						RespondWith(http.StatusNotFound, "{}", http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns false", func() {
				exists, warnings, err := client.CheckRoute(Route{Host: "some-host", DomainGUID: "some-domain-guid"})
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				Expect(exists).To(BeFalse())
			})
		})

		Context("when the CC errors", func() {
			BeforeEach(func() {
				response := `{
				"code": 777,
				"description": "The route could not be found: some-route-guid",
				"error_code": "CF-WUT"
			}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/routes/reserved/domain/some-domain-guid", "host=some-host"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error", func() {
				_, warnings, err := client.CheckRoute(Route{Host: "some-host", DomainGUID: "some-domain-guid"})
				Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        777,
						Description: "The route could not be found: some-route-guid",
						ErrorCode:   "CF-WUT",
					},
					ResponseCode: http.StatusTeapot,
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})
})
