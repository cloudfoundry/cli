package api_test

import (
	"cf"
	. "cf/api"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testhelpers"
	"testing"
)

var findAllRoutesResponse = testhelpers.TestResponse{Status: http.StatusOK, Body: `
{
  "total_results": 1,
  "total_pages": 1,
  "prev_url": null,
  "next_url": null,
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
        }
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
        }
      }
    }
  ]
}`}

var findAllEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/routes?inline-relations-depth=1",
	nil,
	findAllRoutesResponse,
)

func TestRoutesFindAll(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(findAllEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	repo := NewCloudControllerRouteRepository(config)

	routes, err := repo.FindAll()
	assert.NoError(t, err)

	assert.Equal(t, len(routes), 2)

	route := routes[0]
	assert.Equal(t, route.Host, "route-1-host")
	assert.Equal(t, route.Guid, "route-1-guid")
	assert.Equal(t, route.Domain.Name, "cfapps.io")
	assert.Equal(t, route.Domain.Guid, "domain-1-guid")

	route = routes[1]
	assert.Equal(t, route.Guid, "route-2-guid")
}

var findRouteByHostResponse = testhelpers.TestResponse{Status: http.StatusCreated, Body: `
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

var findRouteByHostEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/routes?q=host%3Amy-cool-app",
	nil,
	findRouteByHostResponse,
)

func TestFindByHost(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(findRouteByHostEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	repo := NewCloudControllerRouteRepository(config)

	route, err := repo.FindByHost("my-cool-app")
	assert.NoError(t, err)
	assert.Equal(t, route, cf.Route{Host: "my-cool-app", Guid: "my-route-guid"})
}

var findRouteByHostNotFoundResponse = testhelpers.TestResponse{Status: http.StatusCreated, Body: `
{ "resources": [
]}`}

var findRouteByHostNotFoundEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/routes?q=host%3Amy-cool-app",
	nil,
	findRouteByHostNotFoundResponse,
)

func TestFindByHostWhenHostIsNotFound(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(findRouteByHostNotFoundEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	repo := NewCloudControllerRouteRepository(config)

	_, err := repo.FindByHost("my-cool-app")
	assert.Error(t, err)
}

var createRouteResponse = testhelpers.TestResponse{Status: http.StatusCreated, Body: `
{
    "metadata": {
        "guid": "my-route-guid"
    },
    "entity": {
        "host": "my-cool-app"
    }
}`}

var createRouteEndpoint = testhelpers.CreateEndpoint(
	"POST",
	"/v2/routes",
	testhelpers.RequestBodyMatcher(`{"host":"my-cool-app","domain_guid":"my-domain-guid","space_guid":"my-space-guid"}`),
	createRouteResponse,
)

func TestCreate(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(createRouteEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
	}
	repo := NewCloudControllerRouteRepository(config)

	domain := cf.Domain{Guid: "my-domain-guid"}
	newRoute := cf.Route{Host: "my-cool-app"}

	createdRoute, err := repo.Create(newRoute, domain)
	assert.NoError(t, err)

	assert.Equal(t, createdRoute, cf.Route{Host: "my-cool-app", Guid: "my-route-guid"})
}

var bindRouteEndpoint = testhelpers.CreateEndpoint(
	"PUT",
	"/v2/apps/my-cool-app-guid/routes/my-cool-route-guid",
	nil,
	testhelpers.TestResponse{Status: http.StatusCreated, Body: ""},
)

func TestBind(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(bindRouteEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	repo := NewCloudControllerRouteRepository(config)

	route := cf.Route{Guid: "my-cool-route-guid"}
	app := cf.Application{Guid: "my-cool-app-guid"}

	err := repo.Bind(route, app)
	assert.NoError(t, err)
}
