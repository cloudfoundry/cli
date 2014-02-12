package api_test

import (
	. "cf/api"
	"cf/models"
	"cf/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
)

var _ = Describe("route repository", func() {
	It("lists routes", func() {
		firstRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "GET",
			Path:     "/v2/routes?inline-relations-depth=1",
			Response: firstPageRoutesResponse,
		})

		secondRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "GET",
			Path:     "/v2/routes?inline-relations-depth=1&page=2",
			Response: secondPageRoutesResponse,
		})

		ts, handler, repo, _ := createRoutesRepo(firstRequest, secondRequest)
		defer ts.Close()

		routes := []models.Route{}
		apiResponse := repo.ListRoutes(func(route models.Route) bool {
			routes = append(routes, route)
			return true
		})

		Expect(len(routes)).To(Equal(2))
		Expect(routes[0].Guid).To(Equal("route-1-guid"))
		Expect(routes[1].Guid).To(Equal("route-2-guid"))
		Expect(handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsSuccessful()).To(BeTrue())
	})

	It("finds routes by host", func() {
		request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "GET",
			Path:     "/v2/routes?q=host%3Amy-cool-app",
			Response: findRouteByHostResponse,
		})

		ts, handler, repo, _ := createRoutesRepo(request)
		defer ts.Close()

		route, apiResponse := repo.FindByHost("my-cool-app")

		Expect(handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsNotSuccessful()).To(BeFalse())
		Expect(route.Host).To(Equal("my-cool-app"))
		Expect(route.Guid).To(Equal("my-route-guid"))
	})

	It("returns an error when a route is not found with the given host", func() {
		request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "GET",
			Path:     "/v2/routes?q=host%3Amy-cool-app",
			Response: testnet.TestResponse{Status: http.StatusCreated, Body: ` { "resources": [ ]}`},
		})

		ts, handler, repo, _ := createRoutesRepo(request)
		defer ts.Close()

		_, apiResponse := repo.FindByHost("my-cool-app")

		Expect(handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsNotSuccessful()).To(BeTrue())
	})

	It("finds a route by host and domain", func() {
		request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "GET",
			Path:     "/v2/routes?q=host%3Amy-cool-app%3Bdomain_guid%3Amy-domain-guid",
			Response: findRouteByHostResponse,
		})

		ts, handler, repo, domainRepo := createRoutesRepo(request)
		defer ts.Close()

		domain := models.DomainFields{}
		domain.Guid = "my-domain-guid"
		domainRepo.FindByNameDomain = domain

		route, apiResponse := repo.FindByHostAndDomain("my-cool-app", "my-domain.com")

		Expect(apiResponse.IsNotSuccessful()).To(BeFalse())
		Expect(handler.AllRequestsCalled()).To(BeTrue())
		Expect(domainRepo.FindByNameName).To(Equal("my-domain.com"))
		Expect(route.Host).To(Equal("my-cool-app"))
		Expect(route.Guid).To(Equal("my-route-guid"))
		Expect(route.Domain.Guid).To(Equal(domain.Guid))
	})

	It("returns 'not found' response when there is no route w/ the given domain and host", func() {
		request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "GET",
			Path:     "/v2/routes?q=host%3Amy-cool-app%3Bdomain_guid%3Amy-domain-guid",
			Response: testnet.TestResponse{Status: http.StatusOK, Body: `{ "resources": [ ] }`},
		})

		ts, handler, repo, domainRepo := createRoutesRepo(request)
		defer ts.Close()

		domain := models.DomainFields{}
		domain.Guid = "my-domain-guid"
		domainRepo.FindByNameDomain = domain

		_, apiResponse := repo.FindByHostAndDomain("my-cool-app", "my-domain.com")

		Expect(handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsError()).To(BeFalse())
		Expect(apiResponse.IsNotFound()).To(BeTrue())
	})

	It("creates routes in a given space", func() {
		request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:  "POST",
			Path:    "/v2/routes?inline-relations-depth=1",
			Matcher: testnet.RequestBodyMatcher(`{"host":"my-cool-app","domain_guid":"my-domain-guid","space_guid":"my-space-guid"}`),
			Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
{
"metadata": { "guid": "my-route-guid" },
"entity": { "host": "my-cool-app" }
}`},
		})

		ts, handler, repo, _ := createRoutesRepo(request)
		defer ts.Close()

		createdRoute, apiResponse := repo.CreateInSpace("my-cool-app", "my-domain-guid", "my-space-guid")

		Expect(handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsNotSuccessful()).To(BeFalse())
		Expect(createdRoute.Guid).To(Equal("my-route-guid"))
	})

	It("creates routes", func() {
		request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:  "POST",
			Path:    "/v2/routes?inline-relations-depth=1",
			Matcher: testnet.RequestBodyMatcher(`{"host":"my-cool-app","domain_guid":"my-domain-guid","space_guid":"my-space-guid"}`),
			Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
{
"metadata": { "guid": "my-route-guid" },
"entity": { "host": "my-cool-app" }
}`},
		})

		ts, handler, repo, _ := createRoutesRepo(request)
		defer ts.Close()

		createdRoute, apiResponse := repo.Create("my-cool-app", "my-domain-guid")
		Expect(handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsNotSuccessful()).To(BeFalse())

		Expect(createdRoute.Guid).To(Equal("my-route-guid"))
	})

	It("binds routes", func() {
		request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "PUT",
			Path:     "/v2/apps/my-cool-app-guid/routes/my-cool-route-guid",
			Response: testnet.TestResponse{Status: http.StatusCreated, Body: ""},
		})

		ts, handler, repo, _ := createRoutesRepo(request)
		defer ts.Close()

		apiResponse := repo.Bind("my-cool-route-guid", "my-cool-app-guid")
		Expect(handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsNotSuccessful()).To(BeFalse())
	})

	It("unbinds routes", func() {
		request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "DELETE",
			Path:     "/v2/apps/my-cool-app-guid/routes/my-cool-route-guid",
			Response: testnet.TestResponse{Status: http.StatusCreated, Body: ""},
		})

		ts, handler, repo, _ := createRoutesRepo(request)
		defer ts.Close()

		apiResponse := repo.Unbind("my-cool-route-guid", "my-cool-app-guid")
		Expect(handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsNotSuccessful()).To(BeFalse())
	})

	It("deletes routes", func() {
		request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "DELETE",
			Path:     "/v2/routes/my-cool-route-guid",
			Response: testnet.TestResponse{Status: http.StatusCreated, Body: ""},
		})

		ts, handler, repo, _ := createRoutesRepo(request)
		defer ts.Close()

		apiResponse := repo.Delete("my-cool-route-guid")
		Expect(handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsSuccessful()).To(BeTrue())
	})
})

var firstPageRoutesResponse = testnet.TestResponse{Status: http.StatusOK, Body: `
{
  "next_url": "/v2/routes?inline-relations-depth=1&page=2",
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

func createRoutesRepo(requests ...testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo CloudControllerRouteRepository, domainRepo *testapi.FakeDomainRepository) {
	ts, handler = testnet.NewTLSServer(requests)

	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway()
	domainRepo = &testapi.FakeDomainRepository{}

	repo = NewCloudControllerRouteRepository(configRepo, gateway, domainRepo)
	return
}
