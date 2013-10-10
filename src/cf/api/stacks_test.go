package api_test

import (
	. "cf/api"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	"testing"
)

func TestStacksFindByName(t *testing.T) {
	response := testapi.TestResponse{Status: http.StatusOK, Body: `
{
"resources": [
    {
      "metadata": { "guid": "custom-linux-guid" },
      "entity": { "name": "custom-linux" }
    }
  ]
}`}

	endpoint, status := testapi.CreateCheckableEndpoint(
		"GET",
		"/v2/stacks?q=name%3Alinux",
		nil,
		response,
	)

	ts, repo := createStackRepo(endpoint)
	defer ts.Close()

	stack, apiResponse := repo.FindByName("linux")
	assert.True(t, status.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, stack.Name, "custom-linux")
	assert.Equal(t, stack.Guid, "custom-linux-guid")

	stack, apiResponse = repo.FindByName("stack that does not exist")
	assert.True(t, apiResponse.IsNotSuccessful())
}

var allStacksResponse = testapi.TestResponse{Status: http.StatusOK, Body: `
{
  "resources": [
    {
      "metadata": {
        "guid": "50688ae5-9bfc-4bf6-a4bf-caadb21a32c6",
        "url": "/v2/stacks/50688ae5-9bfc-4bf6-a4bf-caadb21a32c6",
        "created_at": "2013-08-31 01:32:40 +0000",
        "updated_at": "2013-08-31 01:32:40 +0000"
      },
      "entity": {
        "name": "lucid64",
        "description": "Ubuntu 10.04"
      }
    },
    {
      "metadata": {
        "guid": "e8cda251-7ce8-44b9-becb-ba5f5913d8ba",
        "url": "/v2/stacks/e8cda251-7ce8-44b9-becb-ba5f5913d8ba",
        "created_at": "2013-08-31 01:32:40 +0000",
        "updated_at": "2013-08-31 01:32:40 +0000"
      },
      "entity": {
        "name": "lucid64custom",
        "description": "Fake Ubuntu 10.04"
      }
    }
  ]
}`}

func TestStacksFindAll(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"GET",
		"/v2/stacks",
		nil,
		allStacksResponse,
	)

	ts, repo := createStackRepo(endpoint)
	defer ts.Close()

	stacks, apiResponse := repo.FindAll()
	assert.True(t, status.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, len(stacks), 2)
	assert.Equal(t, stacks[0].Name, "lucid64")
	assert.Equal(t, stacks[0].Guid, "50688ae5-9bfc-4bf6-a4bf-caadb21a32c6")
}

func createStackRepo(endpoint http.HandlerFunc) (ts *httptest.Server, repo StackRepository) {
	ts = httptest.NewTLSServer(endpoint)

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerStackRepository(config, gateway)
	return
}
