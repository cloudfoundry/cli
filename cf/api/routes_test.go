package api_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testnet "code.cloudfoundry.org/cli/util/testhelpers/net"

	. "code.cloudfoundry.org/cli/cf/api"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("route repository", func() {
	var (
		ts         *httptest.Server
		handler    *testnet.TestHandler
		configRepo coreconfig.Repository
		repo       CloudControllerRouteRepository
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		configRepo.SetSpaceFields(models.SpaceFields{
			GUID: "the-space-guid",
			Name: "the-space-name",
		})
		gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
		repo = NewCloudControllerRouteRepository(configRepo, gateway)
	})

	AfterEach(func() {
		if ts != nil {
			ts.Close()
		}
	})

	Describe("List routes", func() {
		It("lists routes in the current space", func() {
			ts, handler = testnet.NewServer([]testnet.TestRequest{
				apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     "/v2/spaces/the-space-guid/routes?inline-relations-depth=1",
					Response: firstPageRoutesResponse,
				}),
				apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     "/v2/spaces/the-space-guid/routes?inline-relations-depth=1&page=2",
					Response: secondPageRoutesResponse,
				}),
			})
			configRepo.SetAPIEndpoint(ts.URL)

			routes := []models.Route{}
			apiErr := repo.ListRoutes(func(route models.Route) bool {
				routes = append(routes, route)
				return true
			})

			Expect(len(routes)).To(Equal(2))
			Expect(routes[0].GUID).To(Equal("route-1-guid"))
			Expect(routes[0].Path).To(Equal(""))
			Expect(routes[0].ServiceInstance.GUID).To(Equal("service-guid"))
			Expect(routes[0].ServiceInstance.Name).To(Equal("test-service"))
			Expect(routes[1].GUID).To(Equal("route-2-guid"))
			Expect(routes[1].Path).To(Equal("/path-2"))
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})

		It("lists routes from all the spaces of current org", func() {
			ts, handler = testnet.NewServer([]testnet.TestRequest{
				apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     "/v2/routes?q=organization_guid:my-org-guid&inline-relations-depth=1",
					Response: firstPageRoutesOrgLvlResponse,
				}),
				apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     "/v2/routes?q=organization_guid:my-org-guid&inline-relations-depth=1&page=2",
					Response: secondPageRoutesResponse,
				}),
			})
			configRepo.SetAPIEndpoint(ts.URL)

			routes := []models.Route{}
			apiErr := repo.ListAllRoutes(func(route models.Route) bool {
				routes = append(routes, route)
				return true
			})

			Expect(len(routes)).To(Equal(2))
			Expect(routes[0].GUID).To(Equal("route-1-guid"))
			Expect(routes[0].Space.GUID).To(Equal("space-1-guid"))
			Expect(routes[0].ServiceInstance.GUID).To(Equal("service-guid"))
			Expect(routes[0].ServiceInstance.Name).To(Equal("test-service"))
			Expect(routes[1].GUID).To(Equal("route-2-guid"))
			Expect(routes[1].Space.GUID).To(Equal("space-2-guid"))

			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})
	})

	Describe("Find", func() {
		var ccServer *ghttp.Server
		BeforeEach(func() {
			ccServer = ghttp.NewServer()
			configRepo.SetAPIEndpoint(ccServer.URL())
		})

		AfterEach(func() {
			ccServer.Close()
		})

		Context("when the port is not specified", func() {
			BeforeEach(func() {
				v := url.Values{}
				v.Set("inline-relations-depth", "1")
				v.Set("q", "host:my-cool-app;domain_guid:my-domain-guid;path:/somepath")

				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/routes", v.Encode()),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusCreated, findResponseBodyForHostAndDomainAndPath),
					),
				)
			})

			It("returns the route", func() {
				domain := models.DomainFields{}
				domain.GUID = "my-domain-guid"

				route, apiErr := repo.Find("my-cool-app", domain, "somepath", 0)

				Expect(apiErr).NotTo(HaveOccurred())
				Expect(route.Host).To(Equal("my-cool-app"))
				Expect(route.GUID).To(Equal("my-route-guid"))
				Expect(route.Path).To(Equal("/somepath"))
				Expect(route.Port).To(Equal(0))
				Expect(route.Domain.GUID).To(Equal(domain.GUID))
			})
		})

		Context("when the path is empty", func() {
			BeforeEach(func() {
				v := url.Values{}
				v.Set("inline-relations-depth", "1")
				v.Set("q", "host:my-cool-app;domain_guid:my-domain-guid")

				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/routes", v.Encode()),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusCreated, findResponseBodyForHostAndDomain),
					),
				)
			})

			It("returns the route", func() {
				domain := models.DomainFields{}
				domain.GUID = "my-domain-guid"

				route, apiErr := repo.Find("my-cool-app", domain, "", 0)

				Expect(apiErr).NotTo(HaveOccurred())
				Expect(route.Host).To(Equal("my-cool-app"))
				Expect(route.GUID).To(Equal("my-route-guid"))
				Expect(route.Path).To(Equal(""))
				Expect(route.Domain.GUID).To(Equal(domain.GUID))
			})
		})

		Context("when the route is found", func() {
			BeforeEach(func() {
				v := url.Values{}
				v.Set("inline-relations-depth", "1")
				v.Set("q", "host:my-cool-app;domain_guid:my-domain-guid;path:/somepath;port:8148")

				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/routes", v.Encode()),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusCreated, findResponseBody),
					),
				)
			})

			It("returns the route", func() {
				domain := models.DomainFields{}
				domain.GUID = "my-domain-guid"

				route, apiErr := repo.Find("my-cool-app", domain, "somepath", 8148)

				Expect(apiErr).NotTo(HaveOccurred())
				Expect(route.Host).To(Equal("my-cool-app"))
				Expect(route.GUID).To(Equal("my-route-guid"))
				Expect(route.Path).To(Equal("/somepath"))
				Expect(route.Port).To(Equal(8148))
				Expect(route.Domain.GUID).To(Equal(domain.GUID))
			})
		})

		Context("when the route is not found", func() {
			BeforeEach(func() {
				v := url.Values{}
				v.Set("inline-relations-depth", "1")
				v.Set("q", "host:my-cool-app;domain_guid:my-domain-guid;path:/somepath;port:1478")

				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/routes", v.Encode()),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{ "resources": [] }`),
					),
				)
			})

			It("returns 'not found'", func() {
				ts, handler = testnet.NewServer([]testnet.TestRequest{
					apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
						Method:   "GET",
						Path:     "/v2/routes?q=host%3Amy-cool-app%3Bdomain_guid%3Amy-domain-guid",
						Response: testnet.TestResponse{Status: http.StatusOK, Body: `{ "resources": [ ] }`},
					}),
				})
				configRepo.SetAPIEndpoint(ts.URL)

				domain := models.DomainFields{}
				domain.GUID = "my-domain-guid"

				_, apiErr := repo.Find("my-cool-app", domain, "somepath", 1478)

				Expect(handler).To(HaveAllRequestsCalled())

				Expect(apiErr.(*errors.ModelNotFoundError)).NotTo(BeNil())
			})
		})
	})

	Describe("CreateInSpace", func() {
		var ccServer *ghttp.Server
		BeforeEach(func() {
			ccServer = ghttp.NewServer()
			configRepo.SetAPIEndpoint(ccServer.URL())
		})

		AfterEach(func() {
			if ccServer != nil {
				ccServer.Close()
			}
		})

		Context("when no host, path, or port are given", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v2/routes", "inline-relations-depth=1&async=true"),
						ghttp.VerifyJSON(`
							{
								"domain_guid":"my-domain-guid",
								"space_guid":"my-space-guid"
							}
						`),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
					),
				)
			})

			It("tries to create a route", func() {
				repo.CreateInSpace("", "", "my-domain-guid", "my-space-guid", 0, false)
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})

			Context("when creating the route succeeds", func() {
				BeforeEach(func() {
					h := ccServer.GetHandler(0)
					ccServer.SetHandler(0, ghttp.CombineHandlers(
						h,
						ghttp.RespondWith(http.StatusCreated, `
								{
									"metadata": { "guid": "my-route-guid" },
									"entity": { "host": "my-cool-app" }
								}
							`),
					))
				})

				It("returns the created route", func() {
					createdRoute, err := repo.CreateInSpace("", "", "my-domain-guid", "my-space-guid", 0, false)
					Expect(err).NotTo(HaveOccurred())
					Expect(createdRoute.GUID).To(Equal("my-route-guid"))
				})
			})
		})

		Context("when a host is given", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v2/routes", "inline-relations-depth=1&async=true"),
						ghttp.VerifyJSON(`
							{
								"host":"the-host",
								"domain_guid":"my-domain-guid",
								"space_guid":"my-space-guid"
							}
						`),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
					),
				)
			})

			It("tries to create a route", func() {
				repo.CreateInSpace("the-host", "", "my-domain-guid", "my-space-guid", 0, false)
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})

			Context("when creating the route succeeds", func() {
				BeforeEach(func() {
					h := ccServer.GetHandler(0)
					ccServer.SetHandler(0, ghttp.CombineHandlers(
						h,
						ghttp.RespondWith(http.StatusCreated, `
								{
									"metadata": { "guid": "my-route-guid" },
									"entity": { "host": "the-host" }
								}
							`),
					))
				})

				It("returns the created route", func() {
					createdRoute, err := repo.CreateInSpace("the-host", "", "my-domain-guid", "my-space-guid", 0, false)
					Expect(err).NotTo(HaveOccurred())
					Expect(createdRoute.Host).To(Equal("the-host"))
				})
			})
		})

		Context("when a path is given", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v2/routes", "inline-relations-depth=1&async=true"),
						ghttp.VerifyJSON(`
							{
								"domain_guid":"my-domain-guid",
								"space_guid":"my-space-guid",
								"path":"/the-path"
							}
						`),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
					),
				)
			})

			It("tries to create a route", func() {
				repo.CreateInSpace("", "the-path", "my-domain-guid", "my-space-guid", 0, false)
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})

			Context("when creating the route succeeds", func() {
				BeforeEach(func() {
					h := ccServer.GetHandler(0)
					ccServer.SetHandler(0, ghttp.CombineHandlers(
						h,
						ghttp.RespondWith(http.StatusCreated, `
								{
									"metadata": { "guid": "my-route-guid" },
									"entity": { "path": "the-path" }
								}
							`),
					))
				})

				It("returns the created route", func() {
					createdRoute, err := repo.CreateInSpace("", "the-path", "my-domain-guid", "my-space-guid", 0, false)
					Expect(err).NotTo(HaveOccurred())
					Expect(createdRoute.Path).To(Equal("the-path"))
				})
			})

			Context("when creating the route fails", func() {
				BeforeEach(func() {
					ccServer.Close()
					ccServer = nil
				})

				It("returns an error", func() {
					_, err := repo.CreateInSpace("", "the-path", "my-domain-guid", "my-space-guid", 0, false)
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("when a port is given", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v2/routes", "inline-relations-depth=1&async=true"),
						ghttp.VerifyJSON(`
							{
								"port":9090,
								"domain_guid":"my-domain-guid",
								"space_guid":"my-space-guid"
							}
						`),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
					),
				)
			})

			It("tries to create a route", func() {
				repo.CreateInSpace("", "", "my-domain-guid", "my-space-guid", 9090, false)
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})

			Context("when creating the route succeeds", func() {
				BeforeEach(func() {
					h := ccServer.GetHandler(0)
					ccServer.SetHandler(0, ghttp.CombineHandlers(
						h,
						ghttp.RespondWith(http.StatusCreated, `
							{
								"metadata": { "guid": "my-route-guid" },
								"entity": { "port": 9090 }
							}
						`),
					))
				})

				It("returns the created route", func() {
					createdRoute, err := repo.CreateInSpace("", "", "my-domain-guid", "my-space-guid", 9090, false)
					Expect(err).NotTo(HaveOccurred())
					Expect(createdRoute.Port).To(Equal(9090))
				})
			})
		})

		Context("when random-port is true", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v2/routes", "inline-relations-depth=1&async=true&generate_port=true"),
						ghttp.VerifyJSON(`
							{
								"domain_guid":"my-domain-guid",
								"space_guid":"my-space-guid"
							}
						`),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
					),
				)
			})

			It("tries to create a route", func() {
				repo.CreateInSpace("", "", "my-domain-guid", "my-space-guid", 0, true)
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})

			Context("when creating the route succeeds", func() {
				BeforeEach(func() {
					h := ccServer.GetHandler(0)
					ccServer.SetHandler(0, ghttp.CombineHandlers(
						h,
						ghttp.RespondWith(http.StatusCreated, `
							{
								"metadata": { "guid": "my-route-guid" },
								"entity": { "port": 50321 }
							}
						`),
					))
				})

				It("returns the created route", func() {
					createdRoute, err := repo.CreateInSpace("", "", "my-domain-guid", "my-space-guid", 0, true)
					Expect(err).NotTo(HaveOccurred())
					Expect(createdRoute.Port).To(Equal(50321))
				})
			})
		})
	})

	Describe("Check routes", func() {
		var (
			ccServer *ghttp.Server
			domain   models.DomainFields
		)

		BeforeEach(func() {
			domain = models.DomainFields{
				GUID: "domain-guid",
			}
			ccServer = ghttp.NewServer()
			configRepo.SetAPIEndpoint(ccServer.URL())
		})

		AfterEach(func() {
			ccServer.Close()
		})

		Context("when the route is found", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/routes/reserved/domain/domain-guid/host/my-host", "path=/some-path"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
				)
			})

			It("returns true", func() {
				found, err := repo.CheckIfExists("my-host", domain, "some-path")
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
			})
		})

		Context("when the route is not found", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/routes/reserved/domain/domain-guid/host/my-host", "path=/some-path"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusNotFound, nil),
					),
				)
			})

			It("returns false", func() {
				found, err := repo.CheckIfExists("my-host", domain, "some-path")
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("when finding the route fails", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/routes/reserved/domain/domain-guid/host/my-host", "path=/some-path"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusForbidden, nil),
					),
				)
			})

			It("returns an error", func() {
				_, err := repo.CheckIfExists("my-host", domain, "some-path")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the path is empty", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
				)
			})

			It("should not add a path query param", func() {
				_, err := repo.CheckIfExists("my-host", domain, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(len(ccServer.ReceivedRequests())).To(Equal(1))
				req := ccServer.ReceivedRequests()[0]
				vals := req.URL.Query()
				_, ok := vals["path"]
				Expect(ok).To(BeFalse())
			})
		})
	})

	Describe("Bind routes", func() {
		It("binds routes", func() {
			ts, handler = testnet.NewServer([]testnet.TestRequest{
				apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "PUT",
					Path:     "/v2/apps/my-cool-app-guid/routes/my-cool-route-guid",
					Response: testnet.TestResponse{Status: http.StatusCreated, Body: ""},
				}),
			})
			configRepo.SetAPIEndpoint(ts.URL)

			apiErr := repo.Bind("my-cool-route-guid", "my-cool-app-guid")
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})

		It("unbinds routes", func() {
			ts, handler = testnet.NewServer([]testnet.TestRequest{
				apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "DELETE",
					Path:     "/v2/apps/my-cool-app-guid/routes/my-cool-route-guid",
					Response: testnet.TestResponse{Status: http.StatusCreated, Body: ""},
				}),
			})
			configRepo.SetAPIEndpoint(ts.URL)

			apiErr := repo.Unbind("my-cool-route-guid", "my-cool-app-guid")
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})

	})

	Describe("Delete routes", func() {
		It("deletes routes", func() {
			ts, handler = testnet.NewServer([]testnet.TestRequest{
				apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "DELETE",
					Path:     "/v2/routes/my-cool-route-guid",
					Response: testnet.TestResponse{Status: http.StatusCreated, Body: ""},
				}),
			})
			configRepo.SetAPIEndpoint(ts.URL)

			apiErr := repo.Delete("my-cool-route-guid")
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})
	})

})

var firstPageRoutesResponse = testnet.TestResponse{Status: http.StatusOK, Body: `
{
  "next_url": "/v2/spaces/the-space-guid/routes?inline-relations-depth=1&page=2",
  "resources": [
    {
      "metadata": {
        "guid": "route-1-guid"
      },
      "entity": {
        "host": "route-1-host",
        "path": "",
        "domain": {
          "metadata": {
            "guid": "domain-1-guid"
          },
          "entity": {
            "name": "cfapps.io"
          }
        },
        "space": {
          "metadata": {
            "guid": "space-1-guid"
          },
          "entity": {
            "name": "space-1"
          }
        },
        "apps": [
       	  {
       	    "metadata": {
              "guid": "app-1-guid"
            },
            "entity": {
              "name": "app-1"
       	    }
       	  }
        ],
         "service_instance_url": "/v2/service_instances/service-guid",
        "service_instance": {
           "metadata": {
              "guid": "service-guid",
              "url": "/v2/service_instances/service-guid"
           },
           "entity": {
              "name": "test-service",
              "credentials": {
                 "username": "user",
                 "password": "password"
              },
              "type": "managed_service_instance",
              "route_service_url": "https://something.awesome.com",
              "space_url": "/v2/spaces/space-1-guid"
           }
        }
      }
    }
  ]
}`}

var secondPageRoutesResponse = testnet.TestResponse{Status: http.StatusOK, Body: `
{
  "resources": [
    {
      "metadata": {
        "guid": "route-2-guid"
      },
      "entity": {
        "host": "route-2-host",
        "path": "/path-2",
        "domain": {
          "metadata": {
            "guid": "domain-2-guid"
          },
          "entity": {
            "name": "example.com"
          }
        },
        "space": {
          "metadata": {
            "guid": "space-2-guid"
          },
          "entity": {
            "name": "space-2"
          }
        },
        "apps": [
       	  {
       	    "metadata": {
              "guid": "app-2-guid"
            },
            "entity": {
              "name": "app-2"
       	    }
       	  },
       	  {
       	    "metadata": {
              "guid": "app-3-guid"
            },
            "entity": {
              "name": "app-3"
       	    }
       	  }
        ]
      }
    }
  ]
}`}

var findResponseBody = `
{ "resources": [
	{
		"metadata": {
			"guid": "my-route-guid"
		},
		"entity": {
			"host": "my-cool-app",
			"domain": {
				"metadata": {
					"guid": "my-domain-guid"
				}
			},
			"port": 8148,
			"path": "/somepath"
		}
	}
]}`

var findResponseBodyForHostAndDomain = `
{ "resources": [
	{
		"metadata": {
			"guid": "my-second-route-guid"
		},
		"entity": {
			"host": "my-cool-app",
			"domain": {
				"metadata": {
					"guid": "my-domain-guid"
				}
			},
			"path": "/somepath"
		}
	},
	{
		"metadata": {
			"guid": "my-route-guid"
		},
		"entity": {
			"host": "my-cool-app",
			"domain": {
				"metadata": {
					"guid": "my-domain-guid"
				}
			}
		}
	}
]}`

var findResponseBodyForHostAndDomainAndPath = `
{ "resources": [
	{
		"metadata": {
			"guid": "my-second-route-guid"
		},
		"entity": {
			"host": "my-cool-app",
			"domain": {
				"metadata": {
					"guid": "my-domain-guid"
				}
			},
			"port": 8148,
			"path": "/somepath"
		}
	},
	{
		"metadata": {
			"guid": "my-route-guid"
		},
		"entity": {
			"host": "my-cool-app",
			"domain": {
				"metadata": {
					"guid": "my-domain-guid"
				}
			},
			"path": "/somepath"
		}
	}
]}`

var firstPageRoutesOrgLvlResponse = testnet.TestResponse{Status: http.StatusOK, Body: `
{
  "next_url": "/v2/routes?q=organization_guid:my-org-guid&inline-relations-depth=1&page=2",
  "resources": [
    {
      "metadata": {
        "guid": "route-1-guid"
      },
      "entity": {
        "host": "route-1-host",
        "domain": {
          "metadata": {
            "guid": "domain-1-guid"
          },
          "entity": {
            "name": "cfapps.io"
          }
        },
        "space": {
          "metadata": {
            "guid": "space-1-guid"
          },
          "entity": {
            "name": "space-1"
          }
        },
        "apps": [
          {
            "metadata": {
              "guid": "app-1-guid"
            },
            "entity": {
              "name": "app-1"
            }
          }
        ],
        "service_instance_url": "/v2/service_instances/service-guid",
        "service_instance": {
           "metadata": {
              "guid": "service-guid",
              "url": "/v2/service_instances/service-guid"
           },
           "entity": {
              "name": "test-service",
              "credentials": {
                 "username": "user",
                 "password": "password"
              },
              "type": "managed_service_instance",
              "route_service_url": "https://something.awesome.com",
              "space_url": "/v2/spaces/space-1-guid"
           }
        }
      }
    }
  ]
}`}
