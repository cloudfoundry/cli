package api

import (
	"cf"
	"cf/configuration"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var createRouteEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	bodyBytes, err := ioutil.ReadAll(request.Body)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	acceptHeaderMatches := request.Header.Get("accept") == "application/json"
	methodMatches := request.Method == "POST"
	pathMatches := request.URL.Path == "/v2/routes"
	authMatches := request.Header.Get("authorization") == "BEARER my_access_token"
	expectedBody := `{"host":"my-cool-app","domain_guid":"my-domain-guid","space_guid":"my-space-guid"}`
	bodyMatches := string(bodyBytes) == expectedBody

	if !(acceptHeaderMatches && methodMatches && pathMatches && authMatches && bodyMatches) {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		writer.WriteHeader(http.StatusCreated)
		jsonResponse := `
{
    "metadata": {
        "guid": "my-route-guid"
    },
    "entity": {
        "host": "my-cool-app"
    }
}`
		fmt.Fprintln(writer, jsonResponse)

	}
}

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

var bindRouteEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	acceptHeaderMatches := request.Header.Get("accept") == "application/json"
	methodMatches := request.Method == "PUT"
	pathMatches := request.URL.Path == "/v2/apps/my-cool-app-guid/routes/my-cool-route-guid"
	authMatches := request.Header.Get("authorization") == "BEARER my_access_token"

	if !(acceptHeaderMatches && methodMatches && pathMatches && authMatches) {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		writer.WriteHeader(http.StatusCreated)
		jsonResponse := ``
		fmt.Fprintln(writer, jsonResponse)
	}
}

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
