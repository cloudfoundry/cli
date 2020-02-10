package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Route", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("CreateRoute", func() {
		var (
			route      resources.Route
			warnings   Warnings
			executeErr error
			spaceGUID  string
			domainGUID string
			host       string
			path       string
			ccv3Route  resources.Route
		)

		BeforeEach(func() {
			host = ""
			path = ""
		})

		JustBeforeEach(func() {
			spaceGUID = "space-guid"
			domainGUID = "domain-guid"
			ccv3Route = resources.Route{SpaceGUID: spaceGUID, DomainGUID: domainGUID, Host: host, Path: path}
			route, warnings, executeErr = client.CreateRoute(ccv3Route)
		})

		When("the request succeeds", func() {
			When("no additional flags", func() {
				BeforeEach(func() {
					host = ""
					response := `{
  "guid": "some-route-guid",
  "relationships": {
    "space": {
	  "data": { "guid": "space-guid" }
    },
    "domain": {
	  "data": { "guid": "domain-guid" }
    }
  },
	"host": ""
}`

					expectedBody := `{
  "relationships": {
  	"space": {
      "data": { "guid": "space-guid" }
    },
    "domain": {
	  "data": { "guid": "domain-guid" }
    }
  }
}`

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v3/routes"),
							VerifyJSON(expectedBody),
							RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
						),
					)
				})

				It("returns the given route and all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1"))

					Expect(route).To(Equal(resources.Route{
						GUID:       "some-route-guid",
						SpaceGUID:  "space-guid",
						DomainGUID: "domain-guid",
					}))
				})
			})

			When("hostname is passed in", func() {

				BeforeEach(func() {
					host = "cheesecake"
					response := `{
  "guid": "some-route-guid",
  "relationships": {
    "space": {
			"data": { "guid": "space-guid" }
    },
    "domain": {
			"data": { "guid": "domain-guid" }
    }
  },
	"host": "cheesecake"
}`

					expectedBody := `{
  "relationships": {
  	"space": {
      "data": { "guid": "space-guid" }
    },
    "domain": {
			"data": { "guid": "domain-guid" }
    }
  },
	"host": "cheesecake"
}`

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v3/routes"),
							VerifyJSON(expectedBody),
							RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
						),
					)
				})

				It("returns the given route and all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1"))

					Expect(route).To(Equal(resources.Route{
						GUID:       "some-route-guid",
						SpaceGUID:  "space-guid",
						DomainGUID: "domain-guid",
						Host:       "cheesecake",
					}))
				})
			})

			When("path is passed in", func() {
				BeforeEach(func() {
					path = "lion"

					response := `{
	"guid": "this-route-guid",
	"relationships": {
		"space": {
			"data": {
				"guid": "space-guid"
			}
		},
		"domain": {
			"data": {
				"guid": "domain-guid"
			}
		}
	},
	"path": "lion"
}`
					expectedRequestBody := `{
	"relationships": {
		"space": {
			"data": {
				"guid": "space-guid"
			}
		},
		"domain": {
			"data": {
				"guid": "domain-guid"
			}
		}
	},
	"path": "lion"
}`

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v3/routes"),
							VerifyJSON(expectedRequestBody),
							RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
						),
					)
				})
				When("the request succeeds", func() {
					It("returns the given route and all warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("warning-1"))

						Expect(route).To(Equal(resources.Route{
							GUID:       "this-route-guid",
							SpaceGUID:  "space-guid",
							DomainGUID: "domain-guid",
							Path:       "lion",
						}))
					})
				})
			})

		})

		When("the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				response := `{
  "errors": [
    {
      "code": 10008,
      "detail": "The request is semantically invalid: command presence",
      "title": "CF-UnprocessableEntity"
    },
		{
      "code": 10010,
      "detail": "Isolation segment not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/routes"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   10008,
							Detail: "The request is semantically invalid: command presence",
							Title:  "CF-UnprocessableEntity",
						},
						{
							Code:   10010,
							Detail: "Isolation segment not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("GetRoutes", func() {
		var (
			query      Query
			routes     []resources.Route
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			routes, warnings, executeErr = client.GetRoutes(query)
		})

		When("the request succeeds", func() {
			var (
				response1 string
				response2 string
			)

			BeforeEach(func() {
				response1 = fmt.Sprintf(`
				{
					"pagination": {
						"next": {
							"href": "%s/v3/routes?page=2"
						}
					},
					"resources": [
						{
							"guid": "route-1-guid",
							"url": "hello",
							"metadata": {
								"labels": {
									"key1": "value1"
								}
							}
						},
						{
							"guid": "route-2-guid",
							"url": "bye"
						}
					]
				}`, server.URL())

				response2 = `
				{
					"pagination": {
						"next": null
					},
					"resources": [
						{
							"guid": "route-3-guid"
						}
					]
				}`
			})

			When("not passing any filters", func() {
				BeforeEach(func() {
					query = Query{}

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v3/routes"),
							RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
						),
					)
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v3/routes", "page=2"),
							RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
						),
					)
				})

				It("returns the given route and all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

					Expect(routes).To(Equal([]resources.Route{
						resources.Route{
							GUID: "route-1-guid",
							URL:  "hello",
							Metadata: &resources.Metadata{
								Labels: map[string]types.NullString{
									"key1": types.NewNullString("value1"),
								},
							},
						},
						resources.Route{
							GUID: "route-2-guid",
							URL:  "bye",
						},
						resources.Route{
							GUID: "route-3-guid",
						},
					}))
				})
			})

			When("passing in a query", func() {
				BeforeEach(func() {
					query = Query{Key: "space_guids", Values: []string{"guid1", "guid2"}}

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v3/routes", "space_guids=guid1,guid2"),
							RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
						),
					)
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v3/routes", "page=2", "space_guids=guid1,guid2"),
							RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
						),
					)
				})

				It("passes query params", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

					Expect(routes).To(Equal([]resources.Route{
						resources.Route{
							GUID: "route-1-guid",
							URL:  "hello",
							Metadata: &resources.Metadata{
								Labels: map[string]types.NullString{
									"key1": types.NewNullString("value1"),
								},
							},
						},
						resources.Route{
							GUID: "route-2-guid",
							URL:  "bye",
						},
						resources.Route{
							GUID: "route-3-guid",
						},
					}))
				})
			})
		})
	})

	Describe("DeleteRoute", func() {
		var (
			routeGUID    string
			jobURLString string
			jobURL       JobURL
			warnings     Warnings
			executeErr   error
		)

		JustBeforeEach(func() {
			jobURL, warnings, executeErr = client.DeleteRoute(routeGUID)
		})

		When("route exists", func() {
			routeGUID = "route-guid"
			jobURLString = "https://api.test.com/v3/jobs/job-guid"

			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/routes/route-guid"),
						RespondWith(http.StatusAccepted, nil, http.Header{
							"X-Cf-Warnings": {"this is a warning"},
							"Location":      {jobURLString},
						}),
					),
				)
			})

			It("returns all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(jobURL).To(Equal(JobURL(jobURLString)))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				response := `{
	  "errors": [
	    {
	      "code": 10008,
	      "detail": "The request is semantically invalid: command presence",
	      "title": "CF-UnprocessableEntity"
	    },
			{
	      "code": 10010,
	      "detail": "Isolation segment not found",
	      "title": "CF-ResourceNotFound"
	    }
	  ]
	}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/routes/route-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   10008,
							Detail: "The request is semantically invalid: command presence",
							Title:  "CF-UnprocessableEntity",
						},
						{
							Code:   10010,
							Detail: "Isolation segment not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("MapRoute", func() {
		var (
			routeGUID  = "route-guid"
			appGUID    = "app-guid"
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			warnings, executeErr = client.MapRoute(routeGUID, appGUID)
		})

		When("the request is successful", func() {
			BeforeEach(func() {
				expectedBody := fmt.Sprintf(`
					{
						"destinations": [
						 {
							"app": {
								"guid": "%s"
							}
						 }
						]
					}
				`, appGUID)

				response := `{}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/routes/route-guid/destinations"),
						VerifyJSON(expectedBody),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
			})

			It("returns the warnings and no error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})

		When("the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				response := `{
  "errors": [
    {
      "code": 10008,
      "detail": "The request is semantically invalid: command presence",
      "title": "CF-UnprocessableEntity"
    },
		{
      "code": 10010,
      "detail": "Isolation segment not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/routes/route-guid/destinations"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   10008,
							Detail: "The request is semantically invalid: command presence",
							Title:  "CF-UnprocessableEntity",
						},
						{
							Code:   10010,
							Detail: "Isolation segment not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("GetRouteDestinations", func() {
		var (
			routeGUID    = "some-route-guid"
			destinations []resources.RouteDestination
			warnings     Warnings
			executeErr   error
		)

		JustBeforeEach(func() {
			destinations, warnings, executeErr = client.GetRouteDestinations(routeGUID)
		})

		When("the request succeeds", func() {
			var (
				response string
			)

			BeforeEach(func() {
				response = `
				{
					"destinations": [
						{
							"guid": "destination-1-guid",
							"app": {
								"guid": "app-1-guid",
								"process": {
									"type": "web"
								}
							}
						},
						{
							"guid": "destination-2-guid",
							"app": {
								"guid": "app-2-guid",
								"process": {
									"type": "worker"
								}
							}
						}
					]
				}`
			})

			When("the request succeeds", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v3/routes/some-route-guid/destinations"),
							RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
						),
					)
				})

				It("returns destinations and all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1"))

					Expect(destinations).To(Equal([]resources.RouteDestination{
						{
							GUID: "destination-1-guid",
							App:  resources.RouteDestinationApp{GUID: "app-1-guid", Process: struct{ Type string }{Type: "web"}},
						},
						{
							GUID: "destination-2-guid",
							App:  resources.RouteDestinationApp{GUID: "app-2-guid", Process: struct{ Type string }{Type: "worker"}},
						},
					}))
				})
			})
		})
	})

	Describe("UnmapRoute", func() {
		var (
			routeGUID       string
			destinationGUID string
			warnings        Warnings
			executeErr      error
		)

		JustBeforeEach(func() {
			warnings, executeErr = client.UnmapRoute(routeGUID, destinationGUID)
		})

		When("route exists", func() {
			routeGUID = "route-guid"
			destinationGUID = "destination-guid"

			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/routes/route-guid/destinations/destination-guid"),
						RespondWith(http.StatusNoContent, nil, http.Header{
							"X-Cf-Warnings": {"this is a warning"},
						}),
					),
				)
			})

			It("returns all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				response := `{
	  "errors": [
	    {
	      "code": 10008,
	      "detail": "The request is semantically invalid: command presence",
	      "title": "CF-UnprocessableEntity"
	    },
			{
	      "code": 10010,
	      "detail": "Isolation segment not found",
	      "title": "CF-ResourceNotFound"
	    }
	  ]
	}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/routes/route-guid/destinations/destination-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{
							"X-Cf-Warnings": {"this is a warning"},
						}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   10008,
							Detail: "The request is semantically invalid: command presence",
							Title:  "CF-UnprocessableEntity",
						},
						{
							Code:   10010,
							Detail: "Isolation segment not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("DeleteOrphanedRoutes", func() {
		var (
			spaceGUID  string
			warnings   Warnings
			executeErr error
			jobURL     JobURL
		)
		JustBeforeEach(func() {
			jobURL, warnings, executeErr = client.DeleteOrphanedRoutes(spaceGUID)
		})

		When("the API succeeds", func() {
			BeforeEach(func() {
				spaceGUID = "space-guid"
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/spaces/space-guid/routes", "unmapped=true"),
						RespondWith(
							http.StatusAccepted,
							nil,
							http.Header{"X-Cf-Warnings": {"orphaned-warning"}, "Location": {"job-url"}},
						),
					),
				)
			})

			It("returns the warnings and a job", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("orphaned-warning"))
				Expect(jobURL).To(Equal(JobURL("job-url")))
			})
		})

		When("the API fails", func() {
			BeforeEach(func() {
				spaceGUID = "space-guid"
				response := `{
	  "errors": [
	    {
	      "code": 10008,
	      "detail": "The request is semantically invalid: command presence",
	      "title": "CF-UnprocessableEntity"
	    },
			{
	      "code": 10010,
	      "detail": "Isolation segment not found",
	      "title": "CF-ResourceNotFound"
	    }
	  ]
	}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/spaces/space-guid/routes", "unmapped=true"),
						RespondWith(
							http.StatusTeapot,
							response,
							http.Header{"X-Cf-Warnings": {"orphaned-warning"}},
						),
					),
				)
			})

			It("returns the warnings and a job", func() {

				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   10008,
							Detail: "The request is semantically invalid: command presence",
							Title:  "CF-UnprocessableEntity",
						},
						{
							Code:   10010,
							Detail: "Isolation segment not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("orphaned-warning"))
			})
		})
	})

	Describe("GetApplicationRoutes", func() {
		var (
			appGUID string

			routes     []resources.Route
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			appGUID = "some-app-guid"
		})

		JustBeforeEach(func() {
			routes, warnings, executeErr = client.GetApplicationRoutes(appGUID)
		})

		When("the request succeeds", func() {
			BeforeEach(func() {
				body := `{
	"resources": [
		{
			"guid": "route-guid",
			"host": "host",
			"path": "/path",
			"url": "host.domain.com/path",
			"relationships": {
				"space": {
					"data": {
						"guid": "space-guid"
					}
				},
				"domain": {
					"data": {
						"guid": "domain-guid"
					}
				}
			}
		}, {
			"guid": "route2-guid",
			"host": "",
			"path": "",
			"url": "domain.com",
			"relationships": {
				"space": {
					"data": {
						"guid": "space-guid"
					}
				},
				"domain": {
					"data": {
						"guid": "domain2-guid"
					}
				}
			}
		}
	]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/routes"),
						RespondWith(
							http.StatusOK,
							body,
							http.Header{"X-Cf-Warnings": {"get-app-routes-warning"}, "Location": {"job-url"}},
						),
					),
				)
			})

			It("returns an array of routes", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-app-routes-warning"))

				Expect(routes).To(ConsistOf(
					resources.Route{
						GUID:       "route-guid",
						DomainGUID: "domain-guid",
						SpaceGUID:  "space-guid",
						Host:       "host",
						Path:       "/path",
						URL:        "host.domain.com/path",
					},
					resources.Route{
						GUID:       "route2-guid",
						DomainGUID: "domain2-guid",
						SpaceGUID:  "space-guid",
						Host:       "",
						Path:       "",
						URL:        "domain.com",
					},
				))
			})
		})

		When("there is a cc error", func() {
			BeforeEach(func() {
				response := `{
  "errors": [
    {
      "code": 10008,
      "detail": "The request is semantically invalid: command presence",
      "title": "CF-UnprocessableEntity"
    },
		{
      "code": 10010,
      "detail": "Isolation segment not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/routes"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"get-app-routes-warning"}, "Location": {"job-url"}}),
					),
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   10008,
							Detail: "The request is semantically invalid: command presence",
							Title:  "CF-UnprocessableEntity",
						},
						{
							Code:   10010,
							Detail: "Isolation segment not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("get-app-routes-warning"))
			})
		})
	})
})
