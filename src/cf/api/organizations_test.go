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

func TestOrganizationsFindAll(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/organizations",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": [
			{
			  "metadata": { "guid": "org1-guid" },
			  "entity": { "name": "Org1" }
			},
			{
			  "metadata": { "guid": "org2-guid" },
			  "entity": { "name": "Org2" }
			}
		]}`},
	})

	ts, handler, repo := createOrganizationRepo(t, req)
	defer ts.Close()

	organizations, apiResponse := repo.FindAll()
	assert.True(t, handler.AllRequestsCalled())
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
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/organizations?q=name%3Aorg1&inline-relations-depth=1",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": [{
		  "metadata": { "guid": "org1-guid" },
		  "entity": {
			"name": "Org1",
			"spaces": [{
			  "metadata": { "guid": "space1-guid" },
			  "entity": { "name": "Space1" }
			}],
			"domains": [{
			  "metadata": { "guid": "domain1-guid" },
			  "entity": { "name": "cfapps.io" }
			}]
		  }
		}]}`},
	})

	ts, handler, repo := createOrganizationRepo(t, req)
	defer ts.Close()

	existingOrg := cf.Organization{Guid: "org1-guid", Name: "Org1"}

	org, apiResponse := repo.FindByName("Org1")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())

	assert.Equal(t, org.Name, existingOrg.Name)
	assert.Equal(t, org.Guid, existingOrg.Guid)
	assert.Equal(t, len(org.Spaces), 1)
	assert.Equal(t, org.Spaces[0].Name, "Space1")
	assert.Equal(t, org.Spaces[0].Guid, "space1-guid")
	assert.Equal(t, len(org.Domains), 1)
	assert.Equal(t, org.Domains[0].Name, "cfapps.io")
	assert.Equal(t, org.Domains[0].Guid, "domain1-guid")
}

func TestOrganizationsFindByNameWhenDoesNotExist(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/organizations?q=name%3Aorg1&inline-relations-depth=1",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
	})

	ts, handler, repo := createOrganizationRepo(t, req)
	defer ts.Close()

	_, apiResponse := repo.FindByName("org1")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsError())
	assert.True(t, apiResponse.IsNotFound())
}

func TestCreateOrganization(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "POST",
		Path:     "/v2/organizations",
		Matcher:  testnet.RequestBodyMatcher(`{"name":"my-org"}`),
		Response: testnet.TestResponse{Status: http.StatusCreated},
	})

	ts, handler, repo := createOrganizationRepo(t, req)
	defer ts.Close()

	apiResponse := repo.Create("my-org")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestRenameOrganization(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     "/v2/organizations/my-org-guid",
		Matcher:  testnet.RequestBodyMatcher(`{"name":"my-new-org"}`),
		Response: testnet.TestResponse{Status: http.StatusCreated},
	})

	ts, handler, repo := createOrganizationRepo(t, req)
	defer ts.Close()

	org := cf.Organization{Guid: "my-org-guid"}
	apiResponse := repo.Rename(org, "my-new-org")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestDeleteOrganization(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "DELETE",
		Path:     "/v2/organizations/my-org-guid?recursive=true",
		Response: testnet.TestResponse{Status: http.StatusOK},
	})

	ts, handler, repo := createOrganizationRepo(t, req)
	defer ts.Close()

	org := cf.Organization{Guid: "my-org-guid"}
	apiResponse := repo.Delete(org)
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func createOrganizationRepo(t *testing.T, reqs ...testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo OrganizationRepository) {
	ts, handler = testnet.NewTLSServer(t, reqs)

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerOrganizationRepository(config, gateway)
	return
}
