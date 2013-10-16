package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	"testing"
)

var multipleBuildpacksResponse = testapi.TestResponse{Status: http.StatusOK, Body: `
{
  "total_results": 2,
  "total_pages": 1,
  "prev_url": null,
  "next_url": null,
  "resources": [
    {
      "metadata": {
        "guid": "buildpack1-guid"
      },
      "entity": {
        "name": "Buildpack1",
	"priority" : 1
      }
    },
    {
      "metadata": {
        "guid": "buildpack2-guid"
      },
      "entity": {
        "name": "Buildpack2",
	"priority" : 2
      }
    }
  ]
}`}

var multipleBuildpacksEndpoint, multipleBuildpacksEndpointStatus = testapi.CreateCheckableEndpoint(
	"GET",
	"/v2/buildpacks",
	nil,
	multipleBuildpacksResponse,
)

func TestBuildpacksFindAll(t *testing.T) {
	multipleBuildpacksEndpointStatus.Reset()
	ts, repo := createBuildpackRepo(multipleBuildpacksEndpoint)
	defer ts.Close()

	buildpacks, apiResponse := repo.FindAll()
	assert.True(t, multipleBuildpacksEndpointStatus.Called())
	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, 2, len(buildpacks))

	firstBuildpack := buildpacks[0]
	assert.Equal(t, firstBuildpack.Name, "Buildpack1")
	assert.Equal(t, firstBuildpack.Guid, "buildpack1-guid")

	secondBuildpack := buildpacks[1]
	assert.Equal(t, secondBuildpack.Name, "Buildpack2")
	assert.Equal(t, secondBuildpack.Guid, "buildpack2-guid")
}

func TestBuildpacksFindByName(t *testing.T) {
	response := testapi.TestResponse{Status: http.StatusOK, Body: `
{
  "resources": [
    {
      "metadata": {
        "guid": "buildpack1-guid"
      },
      "entity": {
        "name": "Buildpack1",
		"priority": 10
      }
    }
  ]
}`}

	endpoint, status := testapi.CreateCheckableEndpoint(
		"GET",
		"/v2/buildpacks?name=Buildpack1",
		nil,
		response,
	)

	ts, repo := createBuildpackRepo(endpoint)
	defer ts.Close()

	existingBuildpack := cf.Buildpack{Guid: "buildpack1-guid", Name: "Buildpack1"}

	buildpack, apiResponse := repo.FindByName("Buildpack1")
	assert.True(t, status.Called())
	assert.True(t, apiResponse.IsSuccessful())

	assert.Equal(t, buildpack.Name, existingBuildpack.Name)
	assert.Equal(t, buildpack.Guid, existingBuildpack.Guid)

	buildpack, apiResponse = repo.FindByName("buildpack1")
	assert.True(t, apiResponse.IsNotSuccessful())
}

func createBuildpackRepo(endpoint http.HandlerFunc) (ts *httptest.Server, repo BuildpackRepository) {
	ts = httptest.NewTLSServer(endpoint)

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Name: "my-space", Guid: "my-space-guid"},
	}
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerBuildpackRepository(config, gateway)
	return
}
