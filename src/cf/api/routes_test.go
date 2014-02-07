package api_test

import (
	. "cf/api"
	"cf/configuration"
	"cf/models"
	"cf/net"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testnet "testhelpers/net"
)

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
       	     "host": "my-cool-app"
    	}
    }
]}`}

func createRoutesRepo(t mr.TestingT, requests ...testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo CloudControllerRouteRepository, domainRepo *testapi.FakeDomainRepository) {
	ts, handler = testnet.NewTLSServer(t, requests)
	space := models.SpaceFields{}
	space.Guid = "my-space-guid"
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		SpaceFields: space,
	}

	gateway := net.NewCloudControllerGateway()
	domainRepo = &testapi.FakeDomainRepository{}

	repo = NewCloudControllerRouteRepository(config, gateway, domainRepo)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestRoutesListRoutes", func() {
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

			ts, handler, repo, _ := createRoutesRepo(mr.T(), firstRequest, secondRequest)
			defer ts.Close()

			stopChan := make(chan bool)
			defer close(stopChan)
			routesChan, statusChan := repo.ListRoutes(stopChan)

			routes := []models.Route{}
			for chunk := range routesChan {
				routes = append(routes, chunk...)
			}
			apiResponse := <-statusChan

			assert.Equal(mr.T(), len(routes), 2)
			assert.Equal(mr.T(), routes[0].Guid, "route-1-guid")
			assert.Equal(mr.T(), routes[1].Guid, "route-2-guid")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())
		})
		It("TestRoutesListRoutesWithNoRoutes", func() {

			emptyRoutesRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/routes?inline-relations-depth=1",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
			})

			ts, handler, repo, _ := createRoutesRepo(mr.T(), emptyRoutesRequest)
			defer ts.Close()

			stopChan := make(chan bool)
			defer close(stopChan)
			routesChan, statusChan := repo.ListRoutes(stopChan)

			_, ok := <-routesChan
			apiResponse := <-statusChan

			assert.False(mr.T(), ok)
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())
		})
		It("TestFindByHost", func() {

			request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/routes?q=host%3Amy-cool-app",
				Response: findRouteByHostResponse,
			})

			ts, handler, repo, _ := createRoutesRepo(mr.T(), request)
			defer ts.Close()

			route, apiResponse := repo.FindByHost("my-cool-app")

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
			assert.Equal(mr.T(), route.Host, "my-cool-app")
			assert.Equal(mr.T(), route.Guid, "my-route-guid")
		})
		It("TestFindByHostWhenHostIsNotFound", func() {

			request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/routes?q=host%3Amy-cool-app",
				Response: testnet.TestResponse{Status: http.StatusCreated, Body: ` { "resources": [ ]}`},
			})

			ts, handler, repo, _ := createRoutesRepo(mr.T(), request)
			defer ts.Close()

			_, apiResponse := repo.FindByHost("my-cool-app")

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsNotSuccessful())
		})
		It("TestFindByHostAndDomain", func() {

			request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/routes?q=host%3Amy-cool-app%3Bdomain_guid%3Amy-domain-guid",
				Response: findRouteByHostResponse,
			})

			ts, handler, repo, domainRepo := createRoutesRepo(mr.T(), request)
			defer ts.Close()

			domain := models.DomainFields{}
			domain.Guid = "my-domain-guid"
			domainRepo.FindByNameDomain = domain

			route, apiResponse := repo.FindByHostAndDomain("my-cool-app", "my-domain.com")

			assert.False(mr.T(), apiResponse.IsNotSuccessful())
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.Equal(mr.T(), domainRepo.FindByNameName, "my-domain.com")
			assert.Equal(mr.T(), route.Host, "my-cool-app")
			assert.Equal(mr.T(), route.Guid, "my-route-guid")
			assert.Equal(mr.T(), route.Domain.Guid, domain.Guid)
		})
		It("TestFindByHostAndDomainWhenRouteIsNotFound", func() {

			request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/routes?q=host%3Amy-cool-app%3Bdomain_guid%3Amy-domain-guid",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{ "resources": [ ] }`},
			})

			ts, handler, repo, domainRepo := createRoutesRepo(mr.T(), request)
			defer ts.Close()

			domain := models.DomainFields{}
			domain.Guid = "my-domain-guid"
			domainRepo.FindByNameDomain = domain

			_, apiResponse := repo.FindByHostAndDomain("my-cool-app", "my-domain.com")

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsError())
			assert.True(mr.T(), apiResponse.IsNotFound())
		})
		It("TestCreateInSpace", func() {

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

			ts, handler, repo, _ := createRoutesRepo(mr.T(), request)
			defer ts.Close()

			createdRoute, apiResponse := repo.CreateInSpace("my-cool-app", "my-domain-guid", "my-space-guid")

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
			assert.Equal(mr.T(), createdRoute.Guid, "my-route-guid")
		})
		It("TestCreateRoute", func() {

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

			ts, handler, repo, _ := createRoutesRepo(mr.T(), request)
			defer ts.Close()

			createdRoute, apiResponse := repo.Create("my-cool-app", "my-domain-guid")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())

			assert.Equal(mr.T(), createdRoute.Guid, "my-route-guid")
		})
		It("TestBind", func() {

			request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/apps/my-cool-app-guid/routes/my-cool-route-guid",
				Response: testnet.TestResponse{Status: http.StatusCreated, Body: ""},
			})

			ts, handler, repo, _ := createRoutesRepo(mr.T(), request)
			defer ts.Close()

			apiResponse := repo.Bind("my-cool-route-guid", "my-cool-app-guid")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
		})
		It("TestUnbind", func() {

			request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "DELETE",
				Path:     "/v2/apps/my-cool-app-guid/routes/my-cool-route-guid",
				Response: testnet.TestResponse{Status: http.StatusCreated, Body: ""},
			})

			ts, handler, repo, _ := createRoutesRepo(mr.T(), request)
			defer ts.Close()

			apiResponse := repo.Unbind("my-cool-route-guid", "my-cool-app-guid")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
		})
		It("TestDelete", func() {

			request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "DELETE",
				Path:     "/v2/routes/my-cool-route-guid",
				Response: testnet.TestResponse{Status: http.StatusCreated, Body: ""},
			})

			ts, handler, repo, _ := createRoutesRepo(mr.T(), request)
			defer ts.Close()

			apiResponse := repo.Delete("my-cool-route-guid")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())
		})
	})
}
