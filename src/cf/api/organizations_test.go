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

var multipleOrgResponse = testapi.TestResponse{Status: http.StatusOK, Body: `
{
  "total_results": 2,
  "total_pages": 1,
  "prev_url": null,
  "next_url": null,
  "resources": [
    {
      "metadata": {
        "guid": "org1-guid"
      },
      "entity": {
        "name": "Org1"
      }
    },
    {
      "metadata": {
        "guid": "org2-guid"
      },
      "entity": {
        "name": "Org2"
      }
    }
  ]
}`}

var multipleOrgEndpoint, multipleOrgEndpointStatus = testapi.CreateCheckableEndpoint(
	"GET",
	"/v2/organizations",
	nil,
	multipleOrgResponse,
)

func TestOrganizationsFindAll(t *testing.T) {
	multipleOrgEndpointStatus.Reset()
	ts, repo := createOrganizationRepo(multipleOrgEndpoint)
	defer ts.Close()

	organizations, apiResponse := repo.FindAll()
	assert.True(t, multipleOrgEndpointStatus.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, 2, len(organizations))

	firstOrg := organizations[0]
	assert.Equal(t, firstOrg.Name, "Org1")
	assert.Equal(t, firstOrg.Guid, "org1-guid")

	secondOrg := organizations[1]
	assert.Equal(t, secondOrg.Name, "Org2")
	assert.Equal(t, secondOrg.Guid, "org2-guid")
}

func TestOrganizationsFindByName(t *testing.T) {
	response := testapi.TestResponse{Status: http.StatusOK, Body: `
{
  "resources": [
    {
      "metadata": {
        "guid": "org1-guid"
      },
      "entity": {
        "name": "Org1",
        "spaces": [
          {
            "metadata": {
              "guid": "space1-guid"
            },
            "entity": {
              "name": "Space1"
            }
          }
        ],
        "domains": [
          {
            "metadata": {
              "guid": "domain1-guid"
            },
            "entity": {
              "name": "cfapps.io"
            }
          }
        ]
      }
    }
  ]
}`}

	endpoint, status := testapi.CreateCheckableEndpoint(
		"GET",
		"/v2/organizations?q=name%3Aorg1&inline-relations-depth=1",
		nil,
		response,
	)

	ts, repo := createOrganizationRepo(endpoint)
	defer ts.Close()

	existingOrg := cf.Organization{Guid: "org1-guid", Name: "Org1"}

	org, apiResponse := repo.FindByName("Org1")
	assert.True(t, status.Called())
	assert.False(t, apiResponse.IsNotSuccessful())

	assert.Equal(t, org.Name, existingOrg.Name)
	assert.Equal(t, org.Guid, existingOrg.Guid)
	assert.Equal(t, len(org.Spaces), 1)
	assert.Equal(t, org.Spaces[0].Name, "Space1")
	assert.Equal(t, org.Spaces[0].Guid, "space1-guid")
	assert.Equal(t, len(org.Domains), 1)
	assert.Equal(t, org.Domains[0].Name, "cfapps.io")
	assert.Equal(t, org.Domains[0].Guid, "domain1-guid")

	org, apiResponse = repo.FindByName("org1")
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestOrganizationsFindByNameWhenDoesNotExist(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"GET",
		"/v2/organizations?q=name%3Aorg1&inline-relations-depth=1",
		nil,
		testapi.TestResponse{Status: http.StatusOK, Body: `{ "resources": [ ] }`},
	)

	ts, repo := createOrganizationRepo(endpoint)
	defer ts.Close()

	_, apiResponse := repo.FindByName("org1")
	assert.True(t, status.Called())
	assert.False(t, apiResponse.IsError())
	assert.True(t, apiResponse.IsNotFound())
}

func TestCreateOrganization(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"POST",
		"/v2/organizations",
		testapi.RequestBodyMatcher(`{"name":"my-org"}`),
		testapi.TestResponse{Status: http.StatusCreated},
	)

	ts, repo := createOrganizationRepo(endpoint)
	defer ts.Close()

	apiResponse := repo.Create("my-org")
	assert.True(t, status.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestRenameOrganization(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"PUT",
		"/v2/organizations/my-org-guid",
		testapi.RequestBodyMatcher(`{"name":"my-new-org"}`),
		testapi.TestResponse{Status: http.StatusCreated},
	)

	ts, repo := createOrganizationRepo(endpoint)
	defer ts.Close()

	org := cf.Organization{Guid: "my-org-guid"}
	apiResponse := repo.Rename(org, "my-new-org")
	assert.True(t, status.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestDeleteOrganization(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"DELETE",
		"/v2/organizations/my-org-guid?recursive=true",
		nil,
		testapi.TestResponse{Status: http.StatusOK},
	)

	ts, repo := createOrganizationRepo(endpoint)
	defer ts.Close()

	org := cf.Organization{Guid: "my-org-guid"}
	apiResponse := repo.Delete(org)
	assert.True(t, status.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func createOrganizationRepo(endpoint http.HandlerFunc) (ts *httptest.Server, repo OrganizationRepository) {
	ts = httptest.NewTLSServer(endpoint)

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerOrganizationRepository(config, gateway)
	return
}
