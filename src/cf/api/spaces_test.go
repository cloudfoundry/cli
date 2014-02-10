package api_test

import (
	. "cf/api"
	"cf/configuration"
	"cf/models"
	"cf/net"
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testnet "testhelpers/net"
)

func testSpacesFindByNameWithOrg(t mr.TestingT, orgGuid string, findByName func(SpaceRepository, string) (models.Space, net.ApiResponse)) {
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

	space, apiResponse := findByName(repo, "Space1")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, space.Name, "Space1")
	assert.Equal(t, space.Guid, "space1-guid")

	assert.Equal(t, space.Organization.Guid, "org1-guid")

	assert.Equal(t, len(space.Applications), 2)
	assert.Equal(t, space.Applications[0].Guid, "app1-guid")
	assert.Equal(t, space.Applications[1].Guid, "app2-guid")

	assert.Equal(t, len(space.Domains), 1)
	assert.Equal(t, space.Domains[0].Guid, "domain1-guid")

	assert.Equal(t, len(space.ServiceInstances), 1)
	assert.Equal(t, space.ServiceInstances[0].Guid, "service1-guid")

	assert.True(t, apiResponse.IsSuccessful())
	return
}

func testSpacesDidNotFindByNameWithOrg(t mr.TestingT, orgGuid string, findByName func(SpaceRepository, string) (models.Space, net.ApiResponse)) {
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

func createSpacesRepo(t mr.TestingT, reqs ...testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo SpaceRepository) {
	ts, handler = testnet.NewTLSServer(t, reqs)
	org4 := models.OrganizationFields{}
	org4.Guid = "some-org-guid"

	space5 := models.SpaceFields{}
	space5.Guid = "my-space-guid"
	config := &configuration.Configuration{
		AccessToken:        "BEARER my_access_token",
		Target:             ts.URL,
		OrganizationFields: org4,
		SpaceFields:        space5,
	}
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerSpaceRepository(config, gateway)
	return
}

var _ = Describe("Space Repository", func() {
	It("lists all the spaces", func() {
		firstPageSpacesRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "GET",
			Path:   "/v2/organizations/some-org-guid/spaces",
			Response: testnet.TestResponse{
				Status: http.StatusOK,
				Body: `
				{
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
				Body: `
				{
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

		ts, handler, repo := createSpacesRepo(mr.T(), firstPageSpacesRequest, secondPageSpacesRequest)
		defer ts.Close()

		spaces := []models.Space{}
		apiResponse := repo.ListSpaces(func(space models.Space) bool {
			spaces = append(spaces, space)
			return true
		})

		assert.Equal(mr.T(), len(spaces), 2)
		assert.Equal(mr.T(), spaces[0].Guid, "acceptance-space-guid")
		assert.Equal(mr.T(), spaces[1].Guid, "staging-space-guid")
		assert.True(mr.T(), apiResponse.IsSuccessful())
		assert.True(mr.T(), handler.AllRequestsCalled())
	})

	Describe("finding spaces by name", func() {
		It("returns the space", func() {
			testSpacesFindByNameWithOrg(mr.T(), "some-org-guid",
				func(repo SpaceRepository, spaceName string) (models.Space, net.ApiResponse) {
					return repo.FindByName(spaceName)
				},
			)
		})

		It("can find spaces in a particular org", func() {
			testSpacesFindByNameWithOrg(mr.T(), "another-org-guid",
				func(repo SpaceRepository, spaceName string) (models.Space, net.ApiResponse) {
					return repo.FindByNameInOrg(spaceName, "another-org-guid")
				},
			)
		})

		It("returns a 'not found' response when the space doesn't exist", func() {
			testSpacesDidNotFindByNameWithOrg(mr.T(), "some-org-guid",
				func(repo SpaceRepository, spaceName string) (models.Space, net.ApiResponse) {
					return repo.FindByName(spaceName)
				},
			)
		})

		It("returns a 'not found' response when the space doesn't exist in the given org", func() {
			testSpacesDidNotFindByNameWithOrg(mr.T(), "another-org-guid",
				func(repo SpaceRepository, spaceName string) (models.Space, net.ApiResponse) {
					return repo.FindByNameInOrg(spaceName, "another-org-guid")
				},
			)
		})
	})

	It("creates spaces", func() {
		request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:  "POST",
			Path:    "/v2/spaces",
			Matcher: testnet.RequestBodyMatcher(`{"name":"space-name","organization_guid":"some-org-guid"}`),
			Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
			{
				"metadata": {
					"guid": "space-guid"
				},
				"entity": {
					"name": "space-name"
				}
			}`},
		})

		ts, handler, repo := createSpacesRepo(mr.T(), request)
		defer ts.Close()

		space, apiResponse := repo.Create("space-name", "some-org-guid")
		assert.True(mr.T(), handler.AllRequestsCalled())
		assert.False(mr.T(), apiResponse.IsNotSuccessful())
		assert.Equal(mr.T(), space.Guid, "space-guid")
	})

	It("renames spaces", func() {
		request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "PUT",
			Path:     "/v2/spaces/my-space-guid",
			Matcher:  testnet.RequestBodyMatcher(`{"name":"new-space-name"}`),
			Response: testnet.TestResponse{Status: http.StatusCreated},
		})

		ts, handler, repo := createSpacesRepo(mr.T(), request)
		defer ts.Close()

		apiResponse := repo.Rename("my-space-guid", "new-space-name")
		assert.True(mr.T(), handler.AllRequestsCalled())
		assert.True(mr.T(), apiResponse.IsSuccessful())
	})

	It("deletes spaces", func() {
		request := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "DELETE",
			Path:     "/v2/spaces/my-space-guid?recursive=true",
			Response: testnet.TestResponse{Status: http.StatusOK},
		})

		ts, handler, repo := createSpacesRepo(mr.T(), request)
		defer ts.Close()

		apiResponse := repo.Delete("my-space-guid")
		assert.True(mr.T(), handler.AllRequestsCalled())
		assert.False(mr.T(), apiResponse.IsNotSuccessful())
	})
})
