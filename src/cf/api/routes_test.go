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

var findAllRoutesResponse = testnet.TestResponse{Status: http.StatusOK, Body: `
{
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
    },
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

func TestRoutesFindAll(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/routes?inline-relations-depth=1",
		Response: findAllRoutesResponse,
	})

	ts, handler, repo, _ := createRoutesRepo(t, request)
	defer ts.Close()

	routes, apiResponse := repo.FindAll()

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, len(routes), 2)

	route := routes[0]
	assert.Equal(t, route.Host, "route-1-host")
	assert.Equal(t, route.Guid, "route-1-guid")
	assert.Equal(t, route.Domain.Name, "cfapps.io")
	assert.Equal(t, route.Domain.Guid, "domain-1-guid")
	assert.Equal(t, route.AppNames, []string{"app-1"})

	route = routes[1]
	assert.Equal(t, route.Guid, "route-2-guid")
	assert.Equal(t, route.AppNames, []string{"app-2", "app-3"})
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

	assert.Equal(t, createdRoute, cf.Route{Host: "my-cool-app", Guid: "my-route-guid"})
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

func createRoutesRepo(t *testing.T, request testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo CloudControllerRouteRepository, domainRepo *testapi.FakeDomainRepository) {
	ts, handler = testnet.NewTLSServer(t, []testnet.TestRequest{request})

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
