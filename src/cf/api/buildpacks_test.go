package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
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

var createBuildpackResponse = `
{
    "metadata": {
        "guid": "my-cool-buildpack-guid"
    },
    "entity": {
        "name": "my-cool-buildpack",
		"priority":10
    }
}`

var createBuildpackResponseWithPriority = `
{
    "metadata": {
        "guid": "my-cool-buildpack-guid"
    },
    "entity": {
        "name": "my-cool-buildpack",
		"priority":999
    }
}`

var updateBuildpackResponseWithPriority = `
{
    "metadata": {
        "guid": "my-cool-buildpack-guid"
    },
    "entity": {
        "name": "my-cool-buildpack",
		"priority":555
    }
}`

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

func TestBuildpackCreateRejectsInproperNames(t *testing.T) {
	endpoint := func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintln(writer, "{}")
	}

	ts, repo := createBuildpackRepo(endpoint)
	defer ts.Close()

	createdBuildpack, apiResponse := repo.Create(cf.Buildpack{Name: "name with space"})
	assert.Equal(t, createdBuildpack, cf.Buildpack{})
	assert.Contains(t, apiResponse.Message, "Buildpack name is invalid")

	_, apiResponse = repo.Create(cf.Buildpack{Name: "name-with-inv@lid-chars!"})
	assert.True(t, apiResponse.IsNotSuccessful())

	_, apiResponse = repo.Create(cf.Buildpack{Name: "Valid-Name"})
	assert.True(t, apiResponse.IsSuccessful())

	_, apiResponse = repo.Create(cf.Buildpack{Name: "name_with_numbers_2"})
	assert.True(t, apiResponse.IsSuccessful())
}

func TestCreateBuildpack(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"POST",
		"/v2/buildpacks",
		testapi.RequestBodyMatcher(`{"name":"my-cool-buildpack"}`),
		testapi.TestResponse{Status: http.StatusCreated, Body: createBuildpackResponse},
	)

	ts, repo := createBuildpackRepo(endpoint)
	defer ts.Close()

	buildpack := cf.Buildpack{Name: "my-cool-buildpack"}

	created, apiResponse := repo.Create(buildpack)
	assert.True(t, status.Called())
	assert.True(t, apiResponse.IsSuccessful())

	assert.NotNil(t, created.Guid)
	assert.Equal(t, buildpack.Name, created.Name)
	assert.NotEqual(t, created.Priority, 0)
}

func TestCreateBuildpackWithPriority(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"POST",
		"/v2/buildpacks",
		testapi.RequestBodyMatcher(`{"name":"my-cool-buildpack","priority":999}`),
		testapi.TestResponse{Status: http.StatusCreated, Body: createBuildpackResponseWithPriority},
	)

	ts, repo := createBuildpackRepo(endpoint)
	defer ts.Close()

	priority := 999
	buildpack := cf.Buildpack{Name: "my-cool-buildpack", Priority: &priority}

	created, apiResponse := repo.Create(buildpack)
	assert.True(t, status.Called())
	assert.True(t, apiResponse.IsSuccessful())

	assert.NotNil(t, created.Guid)
	assert.Equal(t, buildpack.Name, created.Name)
	assert.Equal(t, *buildpack.Priority, 999)
}

func TestDeleteBuildpack(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"DELETE",
		"/v2/buildpacks/my-cool-buildpack-guid",
		nil,
		testapi.TestResponse{Status: http.StatusNoContent, Body: ""},
	)

	ts, repo := createBuildpackRepo(endpoint)
	defer ts.Close()

	buildpack := cf.Buildpack{Name: "my-cool-buildpack", Guid: "my-cool-buildpack-guid"}

	apiResponse := repo.Delete(buildpack)
	assert.True(t, status.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestUpdateBuildpack(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"PUT",
		"/v2/buildpacks/my-cool-buildpack-guid",
		testapi.RequestBodyMatcher(`{"name":"my-cool-buildpack","priority":555}`),
		testapi.TestResponse{Status: http.StatusCreated, Body: updateBuildpackResponseWithPriority},
	)

	ts, repo := createBuildpackRepo(endpoint)
	defer ts.Close()

	priority := 555
	buildpack := cf.Buildpack{Name: "my-cool-buildpack", Guid: "my-cool-buildpack-guid", Priority: &priority}

	updated, apiResponse := repo.Update(buildpack)
	assert.True(t, status.Called())
	assert.False(t, apiResponse.IsNotSuccessful())

	assert.Equal(t, buildpack, updated)
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
