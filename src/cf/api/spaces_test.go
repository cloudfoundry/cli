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
	testnet "testhelpers/net"
	"testing"
)

func TestSpacesListSpaces(t *testing.T) {
	firstPageSpacesRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/organizations/some-org-guid/spaces",
		Response: testnet.TestResponse{
			Status: http.StatusOK,
			Body: `{
			"next_url": "/v2/organizations/some-org-guid/spaces?page=2",
			"resources": [
				{
			  		"metadata": {
				  		"guid": "acceptance-space-guid"
			  		},
			  		"entity": {
				  		"name": "acceptance"
			  		}
			  	}
			]
		}`}})

	secondPageSpacesRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/organizations/some-org-guid/spaces?page=2",
		Response: testnet.TestResponse{
			Status: http.StatusOK,
			Body: `{
			"resources": [
			  	{
			  		"metadata": {
				      	"guid": "staging-space-guid"
				  	},
			    	"entity": {
						"name": "staging"
				    }
				}
			]
		}`}})

	ts, handler, repo := createSpacesRepo(t, firstPageSpacesRequest, secondPageSpacesRequest)
	defer ts.Close()

	stopChan := make(chan bool)
	spacesChan, statusChan := repo.ListSpaces(stopChan)

	expectedSpaces := []cf.Space{
		{
			Guid:             "acceptance-space-guid",
			Name:             "acceptance",
			ServiceInstances: []cf.ServiceInstance{},
			Organization:     cf.Organization{},
			Applications:     []cf.Application{},
			Domains:          []cf.Domain{},
		},
		{
			Guid:             "staging-space-guid",
			Name:             "staging",
			ServiceInstances: []cf.ServiceInstance{},
			Organization:     cf.Organization{},
			Applications:     []cf.Application{},
			Domains:          []cf.Domain{},
		},
	}

	spaces := []cf.Space{}
	for chunk := range spacesChan {
		spaces = append(spaces, chunk...)
	}
	apiResponse := <-statusChan

	assert.Equal(t, spaces, expectedSpaces)
	assert.True(t, apiResponse.IsSuccessful())
	assert.True(t, handler.AllRequestsCalled())
}

func TestSpacesFindByName(t *testing.T) {
	testSpacesFindByNameWithOrg(t,
		"some-org-guid",
		func(repo SpaceRepository, spaceName string) (cf.Space, net.ApiResponse) {
			return repo.FindByName(spaceName)
		},
	)
}

func TestSpacesFindByNameInOrg(t *testing.T) {
	org := cf.Organization{Guid: "another-org-guid"}

	testSpacesFindByNameWithOrg(t,
		"another-org-guid",
		func(repo SpaceRepository, spaceName string) (cf.Space, net.ApiResponse) {
			return repo.FindByNameInOrg(spaceName, org)
		},
	)
}

func testSpacesFindByNameWithOrg(t *testing.T, orgGuid string, findByName func(SpaceRepository, string) (cf.Space, net.ApiResponse)) {
	findSpaceByNameResponse := testnet.TestResponse{
		Status: http.StatusOK,
		Body: `
{
  "resources": [
    {
      "metadata": {
        "guid": "space1-guid"
      },
      "entity": {
        "name": "Space1",
        "organization_guid": "org1-guid",
        "organization": {
          "metadata": {
            "guid": "org1-guid"
          },
          "entity": {
            "name": "Org1"
          }
        },
        "apps": [
          {
            "metadata": {
              "guid": "app1-guid"
            },
            "entity": {
              "name": "app1"
            }
          },
          {
            "metadata": {
              "guid": "app2-guid"
            },
            "entity": {
              "name": "app2"
            }
          }
        ],
        "domains": [
          {
            "metadata": {
              "guid": "domain1-guid"
            },
            "entity": {
              "name": "domain1"
            }
          }
        ],
        "service_instances": [
          {
			"metadata": {
              "guid": "service1-guid"
            },
            "entity": {
              "name": "service1"
            }
          }
        ]
      }
    }
  ]
}`}
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     fmt.Sprintf("/v2/organizations/%s/spaces?q=name%%3Aspace1&inline-relations-depth=1", orgGuid),
		Response: findSpaceByNameResponse,
	})

	ts, handler, repo := createSpacesRepo(t, request)
	defer ts.Close()

	existingOrg := cf.Organization{Guid: "org1-guid", Name: "Org1"}
	apps := []cf.Application{
		cf.Application{Name: "app1", Guid: "app1-guid"},
		cf.Application{Name: "app2", Guid: "app2-guid"},
	}
	domains := []cf.Domain{
		cf.Domain{Name: "domain1", Guid: "domain1-guid"},
	}
	services := []cf.ServiceInstance{
		cf.ServiceInstance{Name: "service1", Guid: "service1-guid"},
	}

	space, apiResponse := findByName(repo, "Space1")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, space.Name, "Space1")
	assert.Equal(t, space.Guid, "space1-guid")

	assert.Equal(t, space.Organization, existingOrg)
	assert.Equal(t, space.Applications, apps)
	assert.Equal(t, space.Domains, domains)
	assert.Equal(t, space.ServiceInstances, services)
	assert.True(t, apiResponse.IsSuccessful())

	return
}

func TestSpacesDidNotFindByName(t *testing.T) {
	testSpacesDidNotFindByNameWithOrg(t,
		"some-org-guid",
		func(repo SpaceRepository, spaceName string) (cf.Space, net.ApiResponse) {
			return repo.FindByName(spaceName)
		},
	)
}

func TestSpacesDidNotFindByNameInOrg(t *testing.T) {
	org := cf.Organization{Guid: "another-org-guid"}

	testSpacesDidNotFindByNameWithOrg(t,
		"another-org-guid",
		func(repo SpaceRepository, spaceName string) (cf.Space, net.ApiResponse) {
			return repo.FindByNameInOrg(spaceName, org)
		},
	)
}

func testSpacesDidNotFindByNameWithOrg(t *testing.T, orgGuid string, findByName func(SpaceRepository, string) (cf.Space, net.ApiResponse)) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   fmt.Sprintf("/v2/organizations/%s/spaces?q=name%%3Aspace1&inline-relations-depth=1", orgGuid),
		Response: testnet.TestResponse{
			Status: http.StatusOK,
			Body:   ` { "resources": [ ] }`,
		},
	})

	ts, handler, repo := createSpacesRepo(t, request)
	defer ts.Close()

	_, apiResponse := findByName(repo, "Space1")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsError())
	assert.True(t, apiResponse.IsNotFound())
}

func TestCreateSpace(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "POST",
		Path:     "/v2/spaces",
		Matcher:  testnet.RequestBodyMatcher(`{"name":"space-name","organization_guid":"some-org-guid"}`),
		Response: testnet.TestResponse{Status: http.StatusCreated},
	})

	ts, handler, repo := createSpacesRepo(t, request)
	defer ts.Close()

	apiResponse := repo.Create("space-name")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestRenameSpace(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     "/v2/spaces/my-space-guid",
		Matcher:  testnet.RequestBodyMatcher(`{"name":"new-space-name"}`),
		Response: testnet.TestResponse{Status: http.StatusCreated},
	})

	ts, handler, repo := createSpacesRepo(t, request)
	defer ts.Close()

	space := cf.Space{Guid: "my-space-guid"}
	apiResponse := repo.Rename(space, "new-space-name")
	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestDeleteSpace(t *testing.T) {
	request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "DELETE",
		Path:     "/v2/spaces/my-space-guid?recursive=true",
		Response: testnet.TestResponse{Status: http.StatusOK},
	})

	ts, handler, repo := createSpacesRepo(t, request)
	defer ts.Close()

	space := cf.Space{Guid: "my-space-guid"}
	apiResponse := repo.Delete(space)
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func createSpacesRepo(t *testing.T, reqs ...testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo SpaceRepository) {
	ts, handler = testnet.NewTLSServer(t, reqs)

	config := &configuration.Configuration{
		AccessToken:  "BEARER my_access_token",
		Target:       ts.URL,
		Organization: cf.Organization{Guid: "some-org-guid"},
		Space:        cf.Space{Guid: "my-space-guid"},
	}
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerSpaceRepository(config, gateway)
	return
}
