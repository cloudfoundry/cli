package api_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/api"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("route repository", func() {
	var (
		ts         *httptest.Server
		handler    *testnet.TestHandler
		configRepo core_config.Repository
		repo       CloudControllerRouteRepository
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		configRepo.SetSpaceFields(models.SpaceFields{
			Guid: "the-space-guid",
			Name: "the-space-name",
		})
		gateway := net.NewCloudControllerGateway(configRepo, time.Now, &testterm.FakeUI{})
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
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     "/v2/spaces/the-space-guid/routes?inline-relations-depth=1",
					Response: firstPageRoutesResponse,
				}),
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     "/v2/spaces/the-space-guid/routes?inline-relations-depth=1&page=2",
					Response: secondPageRoutesResponse,
				}),
			})
			configRepo.SetApiEndpoint(ts.URL)

			routes := []models.Route{}
			apiErr := repo.ListRoutes(func(route models.Route) bool {
				routes = append(routes, route)
				return true
			})

			Expect(len(routes)).To(Equal(2))
			Expect(routes[0].Guid).To(Equal("route-1-guid"))
			Expect(routes[0].Path).To(Equal(""))
			Expect(routes[1].Guid).To(Equal("route-2-guid"))
			Expect(routes[1].Path).To(Equal("/path-2"))
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})

		It("lists routes from all the spaces of current org", func() {
			ts, handler = testnet.NewServer([]testnet.TestRequest{
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     "/v2/routes?q=organization_guid:my-org-guid&inline-relations-depth=1",
					Response: firstPageRoutesOrgLvlResponse,
				}),
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     "/v2/routes?q=organization_guid:my-org-guid&inline-relations-depth=1&page=2",
					Response: secondPageRoutesResponse,
				}),
			})
			configRepo.SetApiEndpoint(ts.URL)

			routes := []models.Route{}
			apiErr := repo.ListAllRoutes(func(route models.Route) bool {
				routes = append(routes, route)
				return true
			})

			Expect(len(routes)).To(Equal(2))
			Expect(routes[0].Guid).To(Equal("route-1-guid"))
			Expect(routes[0].Space.Guid).To(Equal("space-1-guid"))
			Expect(routes[1].Guid).To(Equal("route-2-guid"))
			Expect(routes[1].Space.Guid).To(Equal("space-2-guid"))
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})

		Describe("Find", func() {
			var ccServer *ghttp.Server
			BeforeEach(func() {
				ccServer = ghttp.NewServer()
				configRepo.SetApiEndpoint(ccServer.URL())
			})

			AfterEach(func() {
				ccServer.Close()
			})

			Context("when the route is found", func() {
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
							ghttp.RespondWith(http.StatusCreated, findResponseBody),
						),
					)
				})

				It("returns the route", func() {
					domain := models.DomainFields{}
					domain.Guid = "my-domain-guid"

					route, apiErr := repo.Find("my-cool-app", domain, "somepath")

					Expect(apiErr).NotTo(HaveOccurred())
					Expect(route.Host).To(Equal("my-cool-app"))
					Expect(route.Guid).To(Equal("my-route-guid"))
					Expect(route.Path).To(Equal("/somepath"))
					Expect(route.Domain.Guid).To(Equal(domain.Guid))
				})
			})

			Context("when the route is not found", func() {
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
							ghttp.RespondWith(http.StatusOK, `{ "resources": [] }`),
						),
					)
				})

				It("returns 'not found'", func() {
					ts, handler = testnet.NewServer([]testnet.TestRequest{
						testapi.NewCloudControllerTestRequest(testnet.TestRequest{
							Method:   "GET",
							Path:     "/v2/routes?q=host%3Amy-cool-app%3Bdomain_guid%3Amy-domain-guid",
							Response: testnet.TestResponse{Status: http.StatusOK, Body: `{ "resources": [ ] }`},
						}),
					})
					configRepo.SetApiEndpoint(ts.URL)

					domain := models.DomainFields{}
					domain.Guid = "my-domain-guid"

					_, apiErr := repo.Find("my-cool-app", domain, "somepath")

					Expect(handler).To(HaveAllRequestsCalled())

					Expect(apiErr.(*errors.ModelNotFoundError)).NotTo(BeNil())
				})
			})
		})
	})

	Describe("Create routes", func() {
		It("creates routes in a given space", func() {
			ts, handler = testnet.NewServer([]testnet.TestRequest{
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:  "POST",
					Path:    "/v2/routes?inline-relations-depth=1",
					Matcher: testnet.RequestBodyMatcher(`{"host":"my-cool-app","path":"","domain_guid":"my-domain-guid","space_guid":"my-space-guid"}`),
					Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
						{
							"metadata": { "guid": "my-route-guid" },
							"entity": { "host": "my-cool-app" }
						}
					`},
				}),
			})
			configRepo.SetApiEndpoint(ts.URL)

			createdRoute, apiErr := repo.CreateInSpace("my-cool-app", "", "my-domain-guid", "my-space-guid")

			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
			Expect(createdRoute.Guid).To(Equal("my-route-guid"))
		})

		It("creates routes with a path in a given space", func() {
			ts, handler = testnet.NewServer([]testnet.TestRequest{
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:  "POST",
					Path:    "/v2/routes?inline-relations-depth=1",
					Matcher: testnet.RequestBodyMatcher(`{"host":"my-cool-app","path":"/this-is-a-path","domain_guid":"my-domain-guid","space_guid":"my-space-guid"}`),
					Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
						{
							"metadata": { "guid": "my-route-guid" },
							"entity": { "host": "my-cool-app", "path": "/this-is-a-path" }
						}
					`},
				}),
			})
			configRepo.SetApiEndpoint(ts.URL)

			createdRoute, apiErr := repo.CreateInSpace("my-cool-app", "this-is-a-path", "my-domain-guid", "my-space-guid")

			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
			Expect(createdRoute.Guid).To(Equal("my-route-guid"))
			Expect(createdRoute.Path).To(Equal("/this-is-a-path"))
		})

		It("creates routes", func() {
			ts, handler = testnet.NewServer([]testnet.TestRequest{
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:  "POST",
					Path:    "/v2/routes?inline-relations-depth=1",
					Matcher: testnet.RequestBodyMatcher(`{"host":"my-cool-app","path":"/the-path","domain_guid":"my-domain-guid","space_guid":"the-space-guid"}`),
					Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
						{
							"metadata": { "guid": "my-route-guid" },
							"entity": { "host": "my-cool-app" }
						}
					`},
				}),
			})
			configRepo.SetApiEndpoint(ts.URL)

			createdRoute, apiErr := repo.Create("my-cool-app", models.DomainFields{Guid: "my-domain-guid"}, "the-path")
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())

			Expect(createdRoute.Guid).To(Equal("my-route-guid"))
		})

	})

	Describe("Check routes", func() {
		var (
			ccServer *ghttp.Server
			domain   models.DomainFields
		)

		BeforeEach(func() {
			domain = models.DomainFields{
				Guid: "domain-guid",
			}
			ccServer = ghttp.NewServer()
			configRepo.SetApiEndpoint(ccServer.URL())
		})

		AfterEach(func() {
			ccServer.Close()
		})

		Context("when the route is found", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/routes/reserved/domain/domain-guid/host/my-host", "path=some-path"),
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
						ghttp.VerifyRequest("GET", "/v2/routes/reserved/domain/domain-guid/host/my-host", "path=some-path"),
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
						ghttp.VerifyRequest("GET", "/v2/routes/reserved/domain/domain-guid/host/my-host", "path=some-path"),
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
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "PUT",
					Path:     "/v2/apps/my-cool-app-guid/routes/my-cool-route-guid",
					Response: testnet.TestResponse{Status: http.StatusCreated, Body: ""},
				}),
			})
			configRepo.SetApiEndpoint(ts.URL)

			apiErr := repo.Bind("my-cool-route-guid", "my-cool-app-guid")
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})

		It("unbinds routes", func() {
			ts, handler = testnet.NewServer([]testnet.TestRequest{
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "DELETE",
					Path:     "/v2/apps/my-cool-app-guid/routes/my-cool-route-guid",
					Response: testnet.TestResponse{Status: http.StatusCreated, Body: ""},
				}),
			})
			configRepo.SetApiEndpoint(ts.URL)

			apiErr := repo.Unbind("my-cool-route-guid", "my-cool-app-guid")
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})

	})

	Describe("Delete routes", func() {
		It("deletes routes", func() {
			ts, handler = testnet.NewServer([]testnet.TestRequest{
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "DELETE",
					Path:     "/v2/routes/my-cool-route-guid",
					Response: testnet.TestResponse{Status: http.StatusCreated, Body: ""},
				}),
			})
			configRepo.SetApiEndpoint(ts.URL)

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
        ]
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
        ]
      }
    }
  ]
}`}
