package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Route", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("DeleteRouteApplication", func() {
		Context("when the delete is successful", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v2/routes/some-route-guid/apps/some-app-guid"),
						RespondWith(http.StatusNoContent, nil, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the route and warnings", func() {
				warnings, err := client.DeleteRouteApplication("some-route-guid", "some-app-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
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
						VerifyRequest(http.MethodDelete, "/v2/routes/some-route-guid/apps/some-app-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns an error", func() {
				warnings, err := client.DeleteRouteApplication("some-route-guid", "some-app-guid")
				Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("UpdateRouteApplication", func() {
		Context("when route mapping is successful", func() {
			BeforeEach(func() {
				response := `
						{
							"metadata": {
								"guid": "some-route-guid"
							},
							"entity": {
								"domain_guid": "some-domain-guid",
								"host": "some-host",
								"path": "some-path",
								"port": 42,
								"space_guid": "some-space-guid"
							}
						}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/routes/some-route-guid/apps/some-app-guid"),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the route and warnings", func() {
				route, warnings, err := client.UpdateRouteApplication("some-route-guid", "some-app-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				Expect(route).To(Equal(Route{
					DomainGUID: "some-domain-guid",
					GUID:       "some-route-guid",
					Host:       "some-host",
					Path:       "some-path",
					Port:       types.NullInt{IsSet: true, Value: 42},
					SpaceGUID:  "some-space-guid",
				}))
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
						VerifyRequest(http.MethodPut, "/v2/routes/some-route-guid/apps/some-app-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns an error", func() {
				_, warnings, err := client.UpdateRouteApplication("some-route-guid", "some-app-guid")
				Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("CreateRoute", func() {
		Context("when route creation is successful", func() {
			Context("when generate port is true", func() {
				BeforeEach(func() {
					response := `
						{
							"metadata": {
								"guid": "some-route-guid"
							},
							"entity": {
								"domain_guid": "some-domain-guid",
								"host": "some-host",
								"path": "some-path",
								"port": 100000,
								"space_guid": "some-space-guid"
							}
						}`
					requestBody := map[string]interface{}{
						"domain_guid": "some-domain-guid",
						"host":        "some-host",
						"path":        "some-path",
						"port":        42,
						"space_guid":  "some-space-guid",
					}
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v2/routes", "generate_port=true"),
							VerifyJSONRepresenting(requestBody),
							RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("creates the route with a random port", func() {
					route, warnings, err := client.CreateRoute(Route{
						DomainGUID: "some-domain-guid",
						Host:       "some-host",
						Path:       "some-path",
						Port:       types.NullInt{IsSet: true, Value: 42},
						SpaceGUID:  "some-space-guid",
					}, true)

					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))
					Expect(route).To(Equal(Route{
						DomainGUID: "some-domain-guid",
						GUID:       "some-route-guid",
						Host:       "some-host",
						Path:       "some-path",
						Port:       types.NullInt{IsSet: true, Value: 100000},
						SpaceGUID:  "some-space-guid",
					}))
				})
			})

			Context("when generate route is false", func() {
				BeforeEach(func() {
					response := `
						{
							"metadata": {
								"guid": "some-route-guid"
							},
							"entity": {
								"domain_guid": "some-domain-guid",
								"host": "some-host",
								"path": "some-path",
								"port": 42,
								"space_guid": "some-space-guid"
							}
						}`
					requestBody := map[string]interface{}{
						"domain_guid": "some-domain-guid",
						"host":        "some-host",
						"path":        "some-path",
						"port":        42,
						"space_guid":  "some-space-guid",
					}
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v2/routes"),
							VerifyJSONRepresenting(requestBody),
							RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("creates the route with the given port", func() {
					route, warnings, err := client.CreateRoute(Route{
						DomainGUID: "some-domain-guid",
						Host:       "some-host",
						Path:       "some-path",
						Port:       types.NullInt{IsSet: true, Value: 42},
						SpaceGUID:  "some-space-guid",
					}, false)

					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))
					Expect(route).To(Equal(Route{
						DomainGUID: "some-domain-guid",
						GUID:       "some-route-guid",
						Host:       "some-host",
						Path:       "some-path",
						Port:       types.NullInt{IsSet: true, Value: 42},
						SpaceGUID:  "some-space-guid",
					}))
				})
			})

			Context("when sending a basic route", func() {
				BeforeEach(func() {
					response := `
						{
							"metadata": {
								"guid": "some-route-guid"
							},
							"entity": {
								"domain_guid": "some-domain-guid",
								"space_guid": "some-space-guid"
							}
						}`
					requestBody := map[string]interface{}{
						"port":        nil,
						"domain_guid": "some-domain-guid",
						"space_guid":  "some-space-guid",
					}
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v2/routes"),
							VerifyJSONRepresenting(requestBody),
							RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("creates the route with only the space and domain guids", func() {
					route, warnings, err := client.CreateRoute(Route{
						DomainGUID: "some-domain-guid",
						SpaceGUID:  "some-space-guid",
					}, false)

					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))
					Expect(route).To(Equal(Route{
						DomainGUID: "some-domain-guid",
						GUID:       "some-route-guid",
						SpaceGUID:  "some-space-guid",
					}))
				})
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
						VerifyRequest(http.MethodPost, "/v2/routes"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns an error", func() {
				_, warnings, err := client.CreateRoute(Route{}, false)
				Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
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
				routes, warnings, err := client.GetRoutes(QQuery{
					Filter:   OrganizationGUIDFilter,
					Operator: EqualOperator,
					Values:   []string{"some-org-guid"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(routes).To(ConsistOf([]Route{
					{
						GUID:       "route-guid-1",
						Host:       "host-1",
						Path:       "path",
						Port:       types.NullInt{IsSet: false},
						DomainGUID: "some-http-domain",
						SpaceGUID:  "some-space-guid-1",
					},
					{
						GUID:       "route-guid-2",
						Host:       "host-2",
						Path:       "",
						Port:       types.NullInt{IsSet: true, Value: 3333},
						DomainGUID: "some-tcp-domain",
						SpaceGUID:  "some-space-guid-1",
					},
					{
						GUID:       "route-guid-3",
						Host:       "host-3",
						Path:       "path",
						Port:       types.NullInt{IsSet: false},
						DomainGUID: "some-http-domain",
						SpaceGUID:  "some-space-guid-2",
					},
					{
						GUID:       "route-guid-4",
						Host:       "host-4",
						Path:       "",
						Port:       types.NullInt{IsSet: true, Value: 333},
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
				_, _, err := client.GetRoutes()
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
				routes, warnings, err := client.GetApplicationRoutes("some-app-guid", QQuery{
					Filter:   OrganizationGUIDFilter,
					Operator: EqualOperator,
					Values:   []string{"some-org-guid"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(routes).To(ConsistOf([]Route{
					{
						GUID:       "route-guid-1",
						Host:       "host-1",
						Path:       "path",
						Port:       types.NullInt{IsSet: false},
						DomainGUID: "some-http-domain",
						SpaceGUID:  "some-space-guid-1",
					},
					{
						GUID:       "route-guid-2",
						Host:       "host-2",
						Path:       "",
						Port:       types.NullInt{IsSet: true, Value: 3333},
						DomainGUID: "some-tcp-domain",
						SpaceGUID:  "some-space-guid-1",
					},
					{
						GUID:       "route-guid-3",
						Host:       "host-3",
						Path:       "path",
						Port:       types.NullInt{IsSet: false},
						DomainGUID: "some-http-domain",
						SpaceGUID:  "some-space-guid-1",
					},
					{
						GUID:       "route-guid-4",
						Host:       "host-4",
						Path:       "",
						Port:       types.NullInt{IsSet: true, Value: 333},
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
				routes, _, err := client.GetApplicationRoutes("some-app-guid")
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
				routes, _, err := client.GetApplicationRoutes("some-app-guid")
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
				routes, warnings, err := client.GetSpaceRoutes("some-space-guid", QQuery{
					Filter:   SpaceGUIDFilter,
					Operator: EqualOperator,
					Values:   []string{"some-space-guid"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(routes).To(ConsistOf([]Route{
					{
						GUID:       "route-guid-1",
						Host:       "host-1",
						Path:       "path",
						Port:       types.NullInt{IsSet: false},
						DomainGUID: "some-http-domain",
						SpaceGUID:  "some-space-guid-1",
					},
					{
						GUID:       "route-guid-2",
						Host:       "host-2",
						Path:       "",
						Port:       types.NullInt{IsSet: true, Value: 3333},
						DomainGUID: "some-tcp-domain",
						SpaceGUID:  "some-space-guid-1",
					},
					{
						GUID:       "route-guid-3",
						Host:       "host-3",
						Path:       "path",
						Port:       types.NullInt{IsSet: false},
						DomainGUID: "some-http-domain",
						SpaceGUID:  "some-space-guid-1",
					},
					{
						GUID:       "route-guid-4",
						Host:       "host-4",
						Path:       "",
						Port:       types.NullInt{IsSet: true, Value: 333},
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
				routes, _, err := client.GetSpaceRoutes("some-space-guid")
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
				routes, _, err := client.GetSpaceRoutes("some-space-guid")
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
		var (
			route      Route
			exists     bool
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			exists, warnings, executeErr = client.CheckRoute(route)
		})

		Context("API Version < MinVersionHTTPRoutePath", func() {
			BeforeEach(func() {
				client = NewClientWithCustomAPIVersion("2.35.0")
			})

			Context("with minimum params", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/routes/reserved/domain/some-domain-guid/host/some-host"),
							RespondWith(http.StatusNoContent, "{}", http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)

					route = Route{DomainGUID: "some-domain-guid", Host: "some-host"}
				})

				It("does not contain any params", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
					Expect(exists).To(BeTrue())
				})
			})

			Context("with all the params", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/routes/reserved/domain/some-domain-guid/host/some-host", "&"),
							RespondWith(http.StatusNoContent, "{}", http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
					route = Route{
						Host:       "some-host",
						DomainGUID: "some-domain-guid",
						Path:       "some-path",
						Port:       types.NullInt{IsSet: true, Value: 42},
					}
				})

				It("contains all requested parameters", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
					Expect(exists).To(BeTrue())
				})
			})
		})

		Context("MinVersionHTTPRoutePath <= API Version < MinVersionNoHostInReservedRouteEndpoint", func() {
			BeforeEach(func() {
				client = NewClientWithCustomAPIVersion("2.36.0")
			})

			Context("when the route exists", func() {
				Context("with minimum params", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodGet, "/v2/routes/reserved/domain/some-domain-guid/host/some-host"),
								RespondWith(http.StatusNoContent, "{}", http.Header{"X-Cf-Warnings": {"this is a warning"}}),
							),
						)
						route = Route{DomainGUID: "some-domain-guid", Host: "some-host"}
					})

					It("does not contain any params", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
					})
				})

				Context("with all the params", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodGet, "/v2/routes/reserved/domain/some-domain-guid/host/some-host", "path=some-path"),
								RespondWith(http.StatusNoContent, "{}", http.Header{"X-Cf-Warnings": {"this is a warning"}}),
							),
						)
						route = Route{
							Host:       "some-host",
							DomainGUID: "some-domain-guid",
							Path:       "some-path",
						}
					})

					It("contains all requested parameters", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
						Expect(exists).To(BeTrue())
					})
				})
			})

			Context("when the route does not exist", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/routes/reserved/domain/some-domain-guid/host/some-host"),
							RespondWith(http.StatusNotFound, "{}", http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
					route = Route{Host: "some-host", DomainGUID: "some-domain-guid"}
				})

				It("returns false", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
					Expect(exists).To(BeFalse())
				})
			})

			Context("when the CC executeErrors", func() {
				BeforeEach(func() {
					response := `{
						"code": 777,
						"description": "The route could not be found: some-route-guid",
						"error_code": "CF-WUT"
					}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/routes/reserved/domain/some-domain-guid/host/some-host"),
							RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
					route = Route{Host: "some-host", DomainGUID: "some-domain-guid"}
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
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

		Context("MinVersionNoHostInReservedRouteEndpoint <= API Version", func() {
			BeforeEach(func() {
				client = NewClientWithCustomAPIVersion("2.55.0")
			})

			Context("when the route exists", func() {
				Context("with minimum params", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodGet, "/v2/routes/reserved/domain/some-domain-guid"),
								RespondWith(http.StatusNoContent, "{}", http.Header{"X-Cf-Warnings": {"this is a warning"}}),
							),
						)
						route = Route{DomainGUID: "some-domain-guid"}
					})

					It("does not contain any params", func() {
						Expect(executeErr).NotTo(HaveOccurred())
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
						route = Route{
							Host:       "some-host",
							DomainGUID: "some-domain-guid",
							Path:       "some-path",
							Port:       types.NullInt{IsSet: true, Value: 42},
						}
					})

					It("contains all requested parameters", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
						Expect(exists).To(BeTrue())
					})
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
					route = Route{Host: "some-host", DomainGUID: "some-domain-guid"}
				})

				It("returns false", func() {
					Expect(executeErr).NotTo(HaveOccurred())
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
					route = Route{Host: "some-host", DomainGUID: "some-domain-guid"}
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
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
})
