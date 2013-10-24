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

func TestBuildpacksFindAll(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/buildpacks",
		Response: testnet.TestResponse{
			Status: http.StatusOK,
			Body: `{
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
				"position" : 1
			      }
			    },
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

	ts, handler, repo := createBuildpackRepo(t, []testnet.TestRequest{req})
	defer ts.Close()

	buildpacks, apiResponse := repo.FindAll()

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, 2, len(buildpacks))

	firstBuildpack := buildpacks[0]
	assert.Equal(t, firstBuildpack.Name, "Buildpack1")
	assert.Equal(t, firstBuildpack.Guid, "buildpack1-guid")
	assert.Equal(t, *firstBuildpack.Position, 1)

	secondBuildpack := buildpacks[1]
	assert.Equal(t, secondBuildpack.Name, "Buildpack2")
	assert.Equal(t, secondBuildpack.Guid, "buildpack2-guid")
	assert.Equal(t, *secondBuildpack.Position, 2)
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

	ts, handler, repo := createBuildpackRepo(t, []testnet.TestRequest{req})
	defer ts.Close()

	existingBuildpack := cf.Buildpack{Guid: "buildpack1-guid", Name: "Buildpack1"}

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

	ts, handler, repo := createBuildpackRepo(t, []testnet.TestRequest{req})
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

	ts, _, repo := createBuildpackRepo(t, []testnet.TestRequest{badRequest})
	defer ts.Close()

	createdBuildpack, apiResponse := repo.Create(cf.Buildpack{Name: "name with space"})
	assert.True(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, createdBuildpack, cf.Buildpack{})
	assert.Equal(t, apiResponse.ErrorCode, "290003")
	assert.Contains(t, apiResponse.Message, "Buildpack is invalid")
}

func TestCreateBuildpack(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:  "POST",
		Path:    "/v2/buildpacks",
		Matcher: testnet.RequestBodyMatcher(`{"name":"my-cool-buildpack"}`),
		Response: testnet.TestResponse{
			Status: http.StatusCreated,
			Body: `{
			    "metadata": {
			        "guid": "my-cool-buildpack-guid"
			    },
			    "entity": {
			        "name": "my-cool-buildpack",
					"position":10
			    }
			}`},
	})

	ts, handler, repo := createBuildpackRepo(t, []testnet.TestRequest{req})
	defer ts.Close()

	buildpack := cf.Buildpack{Name: "my-cool-buildpack"}

	created, apiResponse := repo.Create(buildpack)

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())

	assert.NotNil(t, created.Guid)
	assert.Equal(t, buildpack.Name, created.Name)
	assert.NotEqual(t, *created.Position, 0)
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

	ts, handler, repo := createBuildpackRepo(t, []testnet.TestRequest{req})
	defer ts.Close()

	position := 999
	buildpack := cf.Buildpack{Name: "my-cool-buildpack", Position: &position}
	created, apiResponse := repo.Create(buildpack)

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())

	assert.NotNil(t, created.Guid)
	assert.Equal(t, buildpack.Name, created.Name)
	assert.Equal(t, *buildpack.Position, 999)
}

func TestDeleteBuildpack(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "DELETE",
		Path:   "/v2/buildpacks/my-cool-buildpack-guid",
		Response: testnet.TestResponse{
			Status: http.StatusNoContent,
		}})

	ts, handler, repo := createBuildpackRepo(t, []testnet.TestRequest{req})
	defer ts.Close()

	buildpack := cf.Buildpack{Name: "my-cool-buildpack", Guid: "my-cool-buildpack-guid"}
	apiResponse := repo.Delete(buildpack)

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

	ts, handler, repo := createBuildpackRepo(t, []testnet.TestRequest{req})
	defer ts.Close()

	position := 555
	buildpack := cf.Buildpack{Name: "my-cool-buildpack", Guid: "my-cool-buildpack-guid", Position: &position}
	updated, apiResponse := repo.Update(buildpack)

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())

	assert.Equal(t, buildpack, updated)
}

func createBuildpackRepo(t *testing.T, requests []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo BuildpackRepository) {
	ts, handler = testnet.NewTLSServer(t, requests)

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Name: "my-space", Guid: "my-space-guid"},
	}
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerBuildpackRepository(config, gateway)
	return
}
