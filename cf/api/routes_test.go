package api_test

import (
	"net/http"
	"net/http/httptest"
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
		ts.Close()
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
			Expect(routes[0].ServiceInstance.Guid).To(Equal("service-guid"))
			Expect(routes[0].ServiceInstance.Name).To(Equal("test-service"))
			Expect(routes[1].Guid).To(Equal("route-2-guid"))
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
			Expect(routes[0].ServiceInstance.Guid).To(Equal("service-guid"))
			Expect(routes[0].ServiceInstance.Name).To(Equal("test-service"))
			Expect(routes[1].Guid).To(Equal("route-2-guid"))
			Expect(routes[1].Space.Guid).To(Equal("space-2-guid"))

			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})
		It("finds a route by host and domain", func() {
			ts, handler = testnet.NewServer([]testnet.TestRequest{
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     "/v2/routes?q=host%3Amy-cool-app%3Bdomain_guid%3Amy-domain-guid",
					Response: findRouteByHostResponse,
				}),
			})
			configRepo.SetApiEndpoint(ts.URL)

			domain := models.DomainFields{}
			domain.Guid = "my-domain-guid"

			route, apiErr := repo.FindByHostAndDomain("my-cool-app", domain)

			Expect(apiErr).NotTo(HaveOccurred())
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(route.Host).To(Equal("my-cool-app"))
			Expect(route.Guid).To(Equal("my-route-guid"))
			Expect(route.Domain.Guid).To(Equal(domain.Guid))
		})

		It("returns 'not found' response when there is no route w/ the given domain and host", func() {
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

			_, apiErr := repo.FindByHostAndDomain("my-cool-app", domain)

			Expect(handler).To(HaveAllRequestsCalled())

			Expect(apiErr.(*errors.ModelNotFoundError)).NotTo(BeNil())
		})
	})

	Describe("Create routes", func() {
		It("creates routes in a given space", func() {
			ts, handler = testnet.NewServer([]testnet.TestRequest{
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:  "POST",
					Path:    "/v2/routes?inline-relations-depth=1",
					Matcher: testnet.RequestBodyMatcher(`{"host":"my-cool-app","domain_guid":"my-domain-guid","space_guid":"my-space-guid"}`),
					Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
						{
							"metadata": { "guid": "my-route-guid" },
							"entity": { "host": "my-cool-app" }
						}
					`},
				}),
			})
			configRepo.SetApiEndpoint(ts.URL)

			createdRoute, apiErr := repo.CreateInSpace("my-cool-app", "my-domain-guid", "my-space-guid")

			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
			Expect(createdRoute.Guid).To(Equal("my-route-guid"))
		})

		It("creates routes", func() {
			ts, handler = testnet.NewServer([]testnet.TestRequest{
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:  "POST",
					Path:    "/v2/routes?inline-relations-depth=1",
					Matcher: testnet.RequestBodyMatcher(`{"host":"my-cool-app","domain_guid":"my-domain-guid","space_guid":"the-space-guid"}`),
					Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
						{
							"metadata": { "guid": "my-route-guid" },
							"entity": { "host": "my-cool-app" }
						}
					`},
				}),
			})
			configRepo.SetApiEndpoint(ts.URL)

			createdRoute, apiErr := repo.Create("my-cool-app", models.DomainFields{Guid: "my-domain-guid"})
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())

			Expect(createdRoute.Guid).To(Equal("my-route-guid"))
		})

	})

	Describe("Check routes", func() {
		It("checks if a route exists", func() {
			ts, handler = testnet.NewServer([]testnet.TestRequest{
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     "/v2/routes/reserved/domain/domain-guid/host/my-host",
					Response: testnet.TestResponse{Status: http.StatusNoContent},
				}),
			})
			configRepo.SetApiEndpoint(ts.URL)

			domain := models.DomainFields{}
			domain.Guid = "domain-guid"

			found, apiErr := repo.CheckIfExists("my-host", domain)

			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
		})
		Context("when the route is not found", func() {
			It("does not return the error", func() {
				ts, handler = testnet.NewServer([]testnet.TestRequest{
					testapi.NewCloudControllerTestRequest(testnet.TestRequest{
						Method:   "GET",
						Path:     "/v2/routes/reserved/domain/domain-guid/host/my-host",
						Response: testnet.TestResponse{Status: http.StatusNotFound},
					}),
				})
				configRepo.SetApiEndpoint(ts.URL)

				domain := models.DomainFields{}
				domain.Guid = "domain-guid"

				found, apiErr := repo.CheckIfExists("my-host", domain)

				Expect(handler).To(HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})
		Context("when there is a random httpError", func() {
			It("returns false and the error", func() {
				ts, handler = testnet.NewServer([]testnet.TestRequest{
					testapi.NewCloudControllerTestRequest(testnet.TestRequest{
						Method:   "GET",
						Path:     "/v2/routes/reserved/domain/domain-guid/host/my-host",
						Response: testnet.TestResponse{Status: http.StatusForbidden},
					}),
				})
				configRepo.SetApiEndpoint(ts.URL)

				domain := models.DomainFields{}
				domain.Guid = "domain-guid"

				found, apiErr := repo.CheckIfExists("my-host", domain)

				Expect(handler).To(HaveAllRequestsCalled())
				Expect(apiErr).To(HaveOccurred())
				Expect(found).To(BeFalse())
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

var findRouteByHostResponse = testnet.TestResponse{Status: http.StatusCreated, Body: `
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
       	     }
    	}
    }
]}`}

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
