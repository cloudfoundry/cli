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

	repo := CloudControllerRouteRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}

	route, err := repo.FindByHost(config, "my-cool-app")
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

	repo := CloudControllerRouteRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}

	_, err := repo.FindByHost(config, "my-cool-app")
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

	repo := CloudControllerRouteRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
	}

	domain := cf.Domain{Guid: "my-domain-guid"}
	newRoute := cf.Route{Host: "my-cool-app"}

	createdRoute, err := repo.Create(config, newRoute, domain)
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

	repo := CloudControllerRouteRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}

	route := cf.Route{Guid: "my-cool-route-guid"}
	app := cf.Application{Guid: "my-cool-app-guid"}

	err := repo.Bind(config, route, app)
	assert.NoError(t, err)
}
