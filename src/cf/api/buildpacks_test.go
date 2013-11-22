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

func TestBuildpacksListBuildpacks(t *testing.T) {
	firstRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/buildpacks",
		Response: testnet.TestResponse{
			Status: http.StatusOK,
			Body: `{
			  "next_url": "/v2/buildpacks?page=2",
			  "resources": [
			    {
			      "metadata": {
			        "guid": "buildpack1-guid"
			      },
			      "entity": {
			        "name": "Buildpack1",
					"position" : 1
			      }
			    }
			  ]
			}`},
	})

	secondRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/buildpacks?page=2",
		Response: testnet.TestResponse{
			Status: http.StatusOK,
			Body: `{
			  "resources": [
			    {
			      "metadata": {
			        "guid": "buildpack2-guid"
			      },
			      "entity": {
			        "name": "Buildpack2",
					"position" : 2
			      }
			    }
			  ]
			}`},
	})

	ts, handler, repo := createBuildpackRepo(t, firstRequest, secondRequest)
	defer ts.Close()

	stopChan := make(chan bool)
	defer close(stopChan)
	buildpacksChan, statusChan := repo.ListBuildpacks(stopChan)

	one := 1
	buildpack := cf.Buildpack{}
	buildpack.Guid = "buildpack1-guid"
	buildpack.Name = "Buildpack1"
	buildpack.Position = &one

	two := 2
	buildpack2 := cf.Buildpack{}
	buildpack2.Guid = "buildpack2-guid"
	buildpack2.Name = "Buildpack2"
	buildpack2.Position = &two

	expectedBuildpacks := []cf.Buildpack{buildpack, buildpack2}

	buildpacks := []cf.Buildpack{}
	for chunk := range buildpacksChan {
		buildpacks = append(buildpacks, chunk...)
	}
	apiResponse := <-statusChan

	assert.Equal(t, buildpacks, expectedBuildpacks)
	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestBuildpacksListBuildpacksWithNoBuildpacks(t *testing.T) {
	emptyBuildpacksRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/buildpacks",
		Response: testnet.TestResponse{
			Status: http.StatusOK,
			Body:   `{"resources": []}`,
		},
	})

	ts, handler, repo := createBuildpackRepo(t, emptyBuildpacksRequest)
	defer ts.Close()

	stopChan := make(chan bool)
	defer close(stopChan)
	buildpacksChan, statusChan := repo.ListBuildpacks(stopChan)

	_, ok := <-buildpacksChan
	apiResponse := <-statusChan

	assert.False(t, ok)
	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

var singleBuildpackResponse = testnet.TestResponse{
	Status: http.StatusOK,
	Body: `{"resources": [
		  {
			  "metadata": {
				  "guid": "buildpack1-guid"
			  },
			  "entity": {
				  "name": "Buildpack1",
				  "position": 10
			  }
		  }
		  ]
	  }`}

var findBuildpackRequest = testnet.TestRequest{
	Method:   "GET",
	Path:     "/v2/buildpacks?q=name%3ABuildpack1",
	Response: singleBuildpackResponse,
}

func TestBuildpacksFindByName(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(findBuildpackRequest)

	ts, handler, repo := createBuildpackRepo(t, req)
	defer ts.Close()
	existingBuildpack := cf.Buildpack{}
	existingBuildpack.Guid = "buildpack1-guid"
	existingBuildpack.Name = "Buildpack1"

	buildpack, apiResponse := repo.FindByName("Buildpack1")

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())

	assert.Equal(t, buildpack.Name, existingBuildpack.Name)
	assert.Equal(t, buildpack.Guid, existingBuildpack.Guid)
	assert.Equal(t, *buildpack.Position, 10)
}

func TestFindByNameWhenBuildpackIsNotFound(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(findBuildpackRequest)
	req.Response = testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`}

	ts, handler, repo := createBuildpackRepo(t, req)
	defer ts.Close()

	_, apiResponse := repo.FindByName("Buildpack1")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsError())
	assert.True(t, apiResponse.IsNotFound())
}

func TestBuildpackCreateRejectsImproperNames(t *testing.T) {
	badRequest := testnet.TestRequest{
		Method: "POST",
		Path:   "/v2/buildpacks",
		Response: testnet.TestResponse{
			Status: http.StatusBadRequest,
			Body: `{
				"code":290003,
				"description":"Buildpack is invalid: [\"name name can only contain alphanumeric characters\"]",
				"error_code":"CF-BuildpackInvalid"
			}`,
		}}

	ts, _, repo := createBuildpackRepo(t, badRequest)
	defer ts.Close()
	one := 1
	createdBuildpack, apiResponse := repo.Create("name with space", &one)
	assert.True(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, createdBuildpack, cf.Buildpack{})
	assert.Equal(t, apiResponse.ErrorCode, "290003")
	assert.Contains(t, apiResponse.Message, "Buildpack is invalid")
}

func TestCreateBuildpackWithPosition(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:  "POST",
		Path:    "/v2/buildpacks",
		Matcher: testnet.RequestBodyMatcher(`{"name":"my-cool-buildpack","position":999}`),
		Response: testnet.TestResponse{
			Status: http.StatusCreated,
			Body: `{
				"metadata": {
					"guid": "my-cool-buildpack-guid"
				},
				"entity": {
					"name": "my-cool-buildpack",
					"position":999
				}
			}`},
	})

	ts, handler, repo := createBuildpackRepo(t, req)
	defer ts.Close()

	position := 999
	created, apiResponse := repo.Create("my-cool-buildpack", &position)

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())

	assert.NotNil(t, created.Guid)
	assert.Equal(t, "my-cool-buildpack", created.Name)
	assert.Equal(t, 999, *created.Position)
}

func TestDeleteBuildpack(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "DELETE",
		Path:   "/v2/buildpacks/my-cool-buildpack-guid",
		Response: testnet.TestResponse{
			Status: http.StatusNoContent,
		}})

	ts, handler, repo := createBuildpackRepo(t, req)
	defer ts.Close()

	apiResponse := repo.Delete("my-cool-buildpack-guid")

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestUpdateBuildpack(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:  "PUT",
		Path:    "/v2/buildpacks/my-cool-buildpack-guid",
		Matcher: testnet.RequestBodyMatcher(`{"name":"my-cool-buildpack","position":555}`),
		Response: testnet.TestResponse{
			Status: http.StatusCreated,
			Body: `{
				
				    "metadata": {
				        "guid": "my-cool-buildpack-guid"
				    },
				    "entity": {
				        "name": "my-cool-buildpack",
						"position":555
				    }
				}`},
	})

	ts, handler, repo := createBuildpackRepo(t, req)
	defer ts.Close()

	position := 555
	buildpack := cf.Buildpack{}
	buildpack.Name = "my-cool-buildpack"
	buildpack.Guid = "my-cool-buildpack-guid"
	buildpack.Position = &position
	updated, apiResponse := repo.Update(buildpack)

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())

	assert.Equal(t, buildpack, updated)
}

func createBuildpackRepo(t *testing.T, requests ...testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo BuildpackRepository) {
	ts, handler = testnet.NewTLSServer(t, requests)
	space := cf.SpaceFields{}
	space.Name = "my-space"
	space.Guid = "my-space-guid"
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		SpaceFields: space,
	}
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerBuildpackRepository(config, gateway)
	return
}
