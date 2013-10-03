package api_test

import (
	. "cf/api"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testhelpers"
	"testing"
)

var singleStackResponse = testhelpers.TestResponse{Status: http.StatusOK, Body: `
{
"resources": [
    {
      "metadata": {
        "guid": "custom-linux-guid"
      },
      "entity": {
        "name": "custom-linux"
      }
    }
  ]
}`}

var singleStackEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/stacks?q=name%3Alinux",
	nil,
	singleStackResponse,
)

func TestStacksFindByName(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(singleStackEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerStackRepository(config, gateway)

	stack, apiStatus := repo.FindByName("linux")
	assert.False(t, apiStatus.NotSuccessful())
	assert.Equal(t, stack.Name, "custom-linux")
	assert.Equal(t, stack.Guid, "custom-linux-guid")

	stack, apiStatus = repo.FindByName("stack that does not exist")
	assert.True(t, apiStatus.NotSuccessful())
}

var allStacksResponse = testhelpers.TestResponse{Status: http.StatusOK, Body: `
{
  "total_results": 2,
  "total_pages": 1,
  "prev_url": null,
  "next_url": null,
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

var allStacksEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/stacks",
	nil,
	allStacksResponse,
)

func TestStacksFindAll(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(allStacksEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerStackRepository(config, gateway)

	stacks, apiStatus := repo.FindAll()
	assert.False(t, apiStatus.NotSuccessful())
	assert.Equal(t, len(stacks), 2)
	assert.Equal(t, stacks[0].Name, "lucid64")
	assert.Equal(t, stacks[0].Guid, "50688ae5-9bfc-4bf6-a4bf-caadb21a32c6")

}
