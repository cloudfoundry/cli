package api_test

import (
	. "cf/api"
	"cf/configuration"
	"cf/models"
	"cf/net"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testnet "testhelpers/net"
)

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

func createBuildpackRepo(t mr.TestingT, requests ...testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo BuildpackRepository) {
	ts, handler = testnet.NewTLSServer(t, requests)
	space := models.SpaceFields{}
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
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestBuildpacksListBuildpacks", func() {
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
					"position" : 1,
					"filename" : "firstbp.zip"
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

			ts, handler, repo := createBuildpackRepo(mr.T(), firstRequest, secondRequest)
			defer ts.Close()

			stopChan := make(chan bool)
			defer close(stopChan)
			buildpacksChan, statusChan := repo.ListBuildpacks(stopChan)

			one := 1
			buildpack := models.Buildpack{}
			buildpack.Guid = "buildpack1-guid"
			buildpack.Name = "Buildpack1"
			buildpack.Position = &one
			buildpack.Filename = "firstbp.zip"

			two := 2
			buildpack2 := models.Buildpack{}
			buildpack2.Guid = "buildpack2-guid"
			buildpack2.Name = "Buildpack2"
			buildpack2.Position = &two

			expectedBuildpacks := []models.Buildpack{buildpack, buildpack2}

			buildpacks := []models.Buildpack{}
			for chunk := range buildpacksChan {
				buildpacks = append(buildpacks, chunk...)
			}
			apiResponse := <-statusChan

			assert.Equal(mr.T(), buildpacks, expectedBuildpacks)
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())
		})
		It("TestBuildpacksListBuildpacksWithNoBuildpacks", func() {

			emptyBuildpacksRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/buildpacks",
				Response: testnet.TestResponse{
					Status: http.StatusOK,
					Body:   `{"resources": []}`,
				},
			})

			ts, handler, repo := createBuildpackRepo(mr.T(), emptyBuildpacksRequest)
			defer ts.Close()

			stopChan := make(chan bool)
			defer close(stopChan)
			buildpacksChan, statusChan := repo.ListBuildpacks(stopChan)

			_, ok := <-buildpacksChan
			apiResponse := <-statusChan

			assert.False(mr.T(), ok)
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())
		})
		It("TestBuildpacksFindByName", func() {

			req := testapi.NewCloudControllerTestRequest(findBuildpackRequest)

			ts, handler, repo := createBuildpackRepo(mr.T(), req)
			defer ts.Close()
			existingBuildpack := models.Buildpack{}
			existingBuildpack.Guid = "buildpack1-guid"
			existingBuildpack.Name = "Buildpack1"

			buildpack, apiResponse := repo.FindByName("Buildpack1")

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())

			assert.Equal(mr.T(), buildpack.Name, existingBuildpack.Name)
			assert.Equal(mr.T(), buildpack.Guid, existingBuildpack.Guid)
			assert.Equal(mr.T(), *buildpack.Position, 10)
		})
		It("TestFindByNameWhenBuildpackIsNotFound", func() {

			req := testapi.NewCloudControllerTestRequest(findBuildpackRequest)
			req.Response = testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`}

			ts, handler, repo := createBuildpackRepo(mr.T(), req)
			defer ts.Close()

			_, apiResponse := repo.FindByName("Buildpack1")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsError())
			assert.True(mr.T(), apiResponse.IsNotFound())
		})
		It("TestBuildpackCreateRejectsImproperNames", func() {

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

			ts, _, repo := createBuildpackRepo(mr.T(), badRequest)
			defer ts.Close()
			one := 1
			createdBuildpack, apiResponse := repo.Create("name with space", &one, nil, nil)
			assert.True(mr.T(), apiResponse.IsNotSuccessful())
			assert.Equal(mr.T(), createdBuildpack, models.Buildpack{})
			assert.Equal(mr.T(), apiResponse.ErrorCode, "290003")
			assert.Contains(mr.T(), apiResponse.Message, "Buildpack is invalid")
		})
		It("TestCreateBuildpackWithPosition", func() {

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

			ts, handler, repo := createBuildpackRepo(mr.T(), req)
			defer ts.Close()

			position := 999
			created, apiResponse := repo.Create("my-cool-buildpack", &position, nil, nil)

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())

			assert.NotNil(mr.T(), created.Guid)
			assert.Equal(mr.T(), "my-cool-buildpack", created.Name)
			assert.Equal(mr.T(), 999, *created.Position)
		})
		It("TestCreateBuildpackEnabled", func() {

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:  "POST",
				Path:    "/v2/buildpacks",
				Matcher: testnet.RequestBodyMatcher(`{"name":"my-cool-buildpack","position":999, "enabled":true}`),
				Response: testnet.TestResponse{
					Status: http.StatusCreated,
					Body: `{
				"metadata": {
					"guid": "my-cool-buildpack-guid"
				},
				"entity": {
					"name": "my-cool-buildpack",
					"position":999,
					"enabled":true
				}
			}`},
			})

			ts, handler, repo := createBuildpackRepo(mr.T(), req)
			defer ts.Close()

			position := 999
			enabled := true
			created, apiResponse := repo.Create("my-cool-buildpack", &position, &enabled, nil)

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())

			assert.NotNil(mr.T(), created.Guid)
			assert.Equal(mr.T(), "my-cool-buildpack", created.Name)
			assert.Equal(mr.T(), 999, *created.Position)
		})
		It("TestDeleteBuildpack", func() {

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "DELETE",
				Path:   "/v2/buildpacks/my-cool-buildpack-guid",
				Response: testnet.TestResponse{
					Status: http.StatusNoContent,
				}})

			ts, handler, repo := createBuildpackRepo(mr.T(), req)
			defer ts.Close()

			apiResponse := repo.Delete("my-cool-buildpack-guid")

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
		})
		It("TestUpdateBuildpack", func() {

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:  "PUT",
				Path:    "/v2/buildpacks/my-cool-buildpack-guid",
				Matcher: testnet.RequestBodyMatcher(`{"name":"my-cool-buildpack","position":555,"enabled":false}`),
				Response: testnet.TestResponse{
					Status: http.StatusCreated,
					Body: `{
				    "metadata": {
				        "guid": "my-cool-buildpack-guid"
				    },
				    "entity": {
				        "name": "my-cool-buildpack",
						"position":555,
						"enabled":false
				    }
				}`},
			})

			ts, handler, repo := createBuildpackRepo(mr.T(), req)
			defer ts.Close()

			position := 555
			enabled := false
			buildpack := models.Buildpack{}
			buildpack.Name = "my-cool-buildpack"
			buildpack.Guid = "my-cool-buildpack-guid"
			buildpack.Position = &position
			buildpack.Enabled = &enabled
			updated, apiResponse := repo.Update(buildpack)

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())

			assert.Equal(mr.T(), buildpack, updated)
		})
		It("TestLockBuildpack", func() {

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:  "PUT",
				Path:    "/v2/buildpacks/my-cool-buildpack-guid",
				Matcher: testnet.RequestBodyMatcher(`{"name":"my-cool-buildpack","locked":true}`),
				Response: testnet.TestResponse{
					Status: http.StatusCreated,
					Body: `{
				
				    "metadata": {
				        "guid": "my-cool-buildpack-guid"
				    },
				    "entity": {
				        "name": "my-cool-buildpack",
						"position":123,
						"locked": true
				    }
				}`},
			})

			ts, handler, repo := createBuildpackRepo(mr.T(), req)
			defer ts.Close()

			position := 123
			locked := true

			buildpack := models.Buildpack{}
			buildpack.Name = "my-cool-buildpack"
			buildpack.Guid = "my-cool-buildpack-guid"
			buildpack.Locked = &locked

			expectedBuildpack := models.Buildpack{}
			expectedBuildpack.Name = "my-cool-buildpack"
			expectedBuildpack.Guid = "my-cool-buildpack-guid"
			expectedBuildpack.Position = &position
			expectedBuildpack.Locked = &locked

			updated, apiResponse := repo.Update(buildpack)

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())

			assert.Equal(mr.T(), expectedBuildpack, updated)
		})
	})
}
