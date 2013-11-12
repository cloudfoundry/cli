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

	expectedRoutes := []cf.Route{
		{
			Guid: "route-1-guid",
			Host: "route-1-host",
			Domain: cf.Domain{
				Name: "cfapps.io",
				Guid: "domain-1-guid",
			},
			Space: cf.Space{
				Name: "space-1",
				Guid: "space-1-guid",
			},
			AppNames: []string{"app-1"},
		},
		{
			Guid: "route-2-guid",
			Host: "route-2-host",
			Domain: cf.Domain{
				Name: "example.com",
				Guid: "domain-2-guid",
			},
			Space: cf.Space{
				Name: "space-2",
				Guid: "space-2-guid",
			},
			AppNames: []string{"app-2", "app-3"},
		},
	}

	routes := []cf.Route{}
	for chunk := range routesChan {
		routes = append(routes, chunk...)
	}
	apiResponse := <-statusChan

	assert.Equal(t, routes, expectedRoutes)
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

	domainRepo.FindByNameDomain = cf.Domain{Guid: "my-domain-guid"}
	route, apiResponse := repo.FindByHostAndDomain("my-cool-app", "my-domain.com")

	assert.False(t, apiResponse.IsNotSuccessful())
	assert.True(t, handler.AllRequestsCalled())
	assert.Equal(t, domainRepo.FindByNameName, "my-domain.com")
	assert.Equal(t, route.Host, "my-cool-app")
	assert.Equal(t, route.Guid, "my-route-guid")
	assert.Equal(t, route.Domain, domainRepo.FindByNameDomain)
}

func TestFindByHostAndDomainWhenRouteIsNotFound(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/routes?q=host%3Amy-cool-app%3Bdomain_guid%3Amy-domain-guid",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{ "resources": [ ] }`},
	})

	ts, handler, repo, domainRepo := createRoutesRepo(t, request)
	defer ts.Close()

	domainRepo.FindByNameDomain = cf.Domain{Guid: "my-domain-guid"}
	_, apiResponse := repo.FindByHostAndDomain("my-cool-app", "my-domain.com")

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsError())
	assert.True(t, apiResponse.IsNotFound())
}

func TestCreateInSpace(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:  "POST",
		Path:    "/v2/routes",
		Matcher: testnet.RequestBodyMatcher(`{"host":"my-cool-app","domain_guid":"my-domain-guid","space_guid":"my-space-guid"}`),
		Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
{
  "metadata": { "guid": "my-route-guid" },
  "entity": { "host": "my-cool-app" }
}`},
	})

	ts, handler, repo, _ := createRoutesRepo(t, request)
	defer ts.Close()

	domain := cf.Domain{Guid: "my-domain-guid"}
	newRoute := cf.Route{Host: "my-cool-app"}
	space := cf.Space{Guid: "my-space-guid"}

	createdRoute, apiResponse := repo.CreateInSpace(newRoute, domain, space)
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())

	assert.Equal(t, createdRoute, cf.Route{Host: "my-cool-app", Guid: "my-route-guid", Domain: domain})
}

func TestCreateRoute(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:  "POST",
		Path:    "/v2/routes",
		Matcher: testnet.RequestBodyMatcher(`{"host":"my-cool-app","domain_guid":"my-domain-guid","space_guid":"my-space-guid"}`),
		Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
{
  "metadata": { "guid": "my-route-guid" },
  "entity": { "host": "my-cool-app" }
}`},
	})

	ts, handler, repo, _ := createRoutesRepo(t, request)
	defer ts.Close()

	domain := cf.Domain{Guid: "my-domain-guid"}
	newRoute := cf.Route{Host: "my-cool-app"}

	createdRoute, apiResponse := repo.Create(newRoute, domain)
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())

	assert.Equal(t, createdRoute, cf.Route{Host: "my-cool-app", Guid: "my-route-guid", Domain: domain})
}

func TestBind(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     "/v2/apps/my-cool-app-guid/routes/my-cool-route-guid",
		Response: testnet.TestResponse{Status: http.StatusCreated, Body: ""},
	})

	ts, handler, repo, _ := createRoutesRepo(t, request)
	defer ts.Close()

	route := cf.Route{Guid: "my-cool-route-guid"}
	app := cf.Application{Guid: "my-cool-app-guid"}

	apiResponse := repo.Bind(route, app)
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

	route := cf.Route{Guid: "my-cool-route-guid"}
	app := cf.Application{Guid: "my-cool-app-guid"}

	apiResponse := repo.Unbind(route, app)
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

	route := cf.Route{Guid: "my-cool-route-guid"}

	apiResponse := repo.Delete(route)
	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

func createRoutesRepo(t *testing.T, requests ...testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo CloudControllerRouteRepository, domainRepo *testapi.FakeDomainRepository) {
	ts, handler = testnet.NewTLSServer(t, requests)

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
	}

	gateway := net.NewCloudControllerGateway()
	domainRepo = &testapi.FakeDomainRepository{}

	repo = NewCloudControllerRouteRepository(config, gateway, domainRepo)
	return
}
