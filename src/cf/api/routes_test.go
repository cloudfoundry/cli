package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testnet "testhelpers/net"
	"testing"
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

func TestRoutesListRoutes(t *testing.T) {
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

	ts, handler, repo, _ := createRoutesRepo(t, firstRequest, secondRequest)
	defer ts.Close()

	stopChan := make(chan bool)
	defer close(stopChan)
	routesChan, statusChan := repo.ListRoutes(stopChan)

	routes := []cf.Route{}
	for chunk := range routesChan {
		routes = append(routes, chunk...)
	}
	apiResponse := <-statusChan

	assert.Equal(t, len(routes), 2)
	assert.Equal(t, routes[0].Guid, "route-1-guid")
	assert.Equal(t, routes[1].Guid, "route-2-guid")
	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestRoutesListRoutesWithNoRoutes(t *testing.T) {
	emptyRoutesRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/routes?inline-relations-depth=1",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
	})

	ts, handler, repo, _ := createRoutesRepo(t, emptyRoutesRequest)
	defer ts.Close()

	stopChan := make(chan bool)
	defer close(stopChan)
	routesChan, statusChan := repo.ListRoutes(stopChan)

	_, ok := <-routesChan
	apiResponse := <-statusChan

	assert.False(t, ok)
	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

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

func TestFindByHost(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/routes?q=host%3Amy-cool-app",
		Response: findRouteByHostResponse,
	})

	ts, handler, repo, _ := createRoutesRepo(t, request)
	defer ts.Close()

	route, apiResponse := repo.FindByHost("my-cool-app")

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, route.Host, "my-cool-app")
	assert.Equal(t, route.Guid, "my-route-guid")
}

func TestFindByHostWhenHostIsNotFound(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/routes?q=host%3Amy-cool-app",
		Response: testnet.TestResponse{Status: http.StatusCreated, Body: ` { "resources": [ ]}`},
	})

	ts, handler, repo, _ := createRoutesRepo(t, request)
	defer ts.Close()

	_, apiResponse := repo.FindByHost("my-cool-app")

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsNotSuccessful())
}

func TestFindByHostAndDomain(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/routes?q=host%3Amy-cool-app%3Bdomain_guid%3Amy-domain-guid",
		Response: findRouteByHostResponse,
	})

	ts, handler, repo, domainRepo := createRoutesRepo(t, request)
	defer ts.Close()

	domain := cf.Domain{}
	domain.Guid = "my-domain-guid"
	domainRepo.FindByNameDomain = domain

	route, apiResponse := repo.FindByHostAndDomain("my-cool-app", "my-domain.com")

	assert.False(t, apiResponse.IsNotSuccessful())
	assert.True(t, handler.AllRequestsCalled())
	assert.Equal(t, domainRepo.FindByNameName, "my-domain.com")
	assert.Equal(t, route.Host, "my-cool-app")
	assert.Equal(t, route.Guid, "my-route-guid")
	assert.Equal(t, route.Domain.Guid, domain.Guid)
}

func TestFindByHostAndDomainWhenRouteIsNotFound(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/routes?q=host%3Amy-cool-app%3Bdomain_guid%3Amy-domain-guid",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{ "resources": [ ] }`},
	})

	ts, handler, repo, domainRepo := createRoutesRepo(t, request)
	defer ts.Close()

	domain := cf.Domain{}
	domain.Guid = "my-domain-guid"
	domainRepo.FindByNameDomain = domain

	_, apiResponse := repo.FindByHostAndDomain("my-cool-app", "my-domain.com")

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsError())
	assert.True(t, apiResponse.IsNotFound())
}

func TestCreateInSpace(t *testing.T) {
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

	ts, handler, repo, _ := createRoutesRepo(t, request)
	defer ts.Close()

	createdRoute, apiResponse := repo.CreateInSpace("my-cool-app", "my-domain-guid", "my-space-guid")

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, createdRoute.Guid, "my-route-guid")
}

func TestCreateRoute(t *testing.T) {
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

	ts, handler, repo, _ := createRoutesRepo(t, request)
	defer ts.Close()

	createdRoute, apiResponse := repo.Create("my-cool-app", "my-domain-guid")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())

	assert.Equal(t, createdRoute.Guid, "my-route-guid")
}

func TestBind(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     "/v2/apps/my-cool-app-guid/routes/my-cool-route-guid",
		Response: testnet.TestResponse{Status: http.StatusCreated, Body: ""},
	})

	ts, handler, repo, _ := createRoutesRepo(t, request)
	defer ts.Close()

	apiResponse := repo.Bind("my-cool-route-guid", "my-cool-app-guid")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestUnbind(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "DELETE",
		Path:     "/v2/apps/my-cool-app-guid/routes/my-cool-route-guid",
		Response: testnet.TestResponse{Status: http.StatusCreated, Body: ""},
	})

	ts, handler, repo, _ := createRoutesRepo(t, request)
	defer ts.Close()

	apiResponse := repo.Unbind("my-cool-route-guid", "my-cool-app-guid")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestDelete(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "DELETE",
		Path:     "/v2/routes/my-cool-route-guid",
		Response: testnet.TestResponse{Status: http.StatusCreated, Body: ""},
	})

	ts, handler, repo, _ := createRoutesRepo(t, request)
	defer ts.Close()

	apiResponse := repo.Delete("my-cool-route-guid")
	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

func createRoutesRepo(t *testing.T, requests ...testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo CloudControllerRouteRepository, domainRepo *testapi.FakeDomainRepository) {
	ts, handler = testnet.NewTLSServer(t, requests)
	space := cf.SpaceFields{}
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
