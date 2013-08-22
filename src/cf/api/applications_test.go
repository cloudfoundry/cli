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

var multipleAppsEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	acceptHeaderMatches := request.Header.Get("accept") == "application/json"
	methodMatches := request.Method == "GET"
	pathMatches := request.URL.Path == "/v2/spaces/my-space-guid/apps"
	authMatches := request.Header.Get("authorization") == "BEARER my_access_token"

	if !(acceptHeaderMatches && methodMatches && pathMatches && authMatches) {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonResponse := `
{
  "resources": [
    {
      "metadata": {
        "guid": "app1-guid"
      },
      "entity": {
        "name": "App1"
      }
    },
    {
      "metadata": {
        "guid": "app2-guid"
      },
      "entity": {
        "name": "App2"
      }
    }
  ]
}`
	fmt.Fprintln(writer, jsonResponse)
}

func TestFindByName(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleAppsEndpoint))
	defer ts.Close()

	repo := CloudControllerApplicationRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Name: "my-space", Guid: "my-space-guid"},
	}

	existingApp := cf.Application{Guid: "app1-guid", Name: "App1"}

	app, err := repo.FindByName(config, "App1")
	assert.NoError(t, err)
	assert.Equal(t, app, existingApp)

	app, err = repo.FindByName(config, "app1")
	assert.NoError(t, err)
	assert.Equal(t, app, existingApp)

	app, err = repo.FindByName(config, "app that does not exist")
	assert.Error(t, err)
}

var setEnvEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	bodyBytes, err := ioutil.ReadAll(request.Body)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	acceptHeaderMatches := request.Header.Get("accept") == "application/json"
	methodMatches := request.Method == "PUT"
	pathMatches := request.URL.Path == "/v2/apps/app1-guid"
	authMatches := request.Header.Get("authorization") == "BEARER my_access_token"
	bodyMatches := string(bodyBytes) == `{"environment_json":{"DATABASE_URL":"mysql://example.com/my-db"}}`

	if !(acceptHeaderMatches && methodMatches && pathMatches && authMatches && bodyMatches) {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		writer.WriteHeader(http.StatusCreated)
	}
}

func TestSetEnv(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(setEnvEndpoint))
	defer ts.Close()

	repo := CloudControllerApplicationRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}

	app := cf.Application{Guid: "app1-guid", Name: "App1"}

	err := repo.SetEnv(config, app, "DATABASE_URL", "mysql://example.com/my-db")

	assert.NoError(t, err)
}
