package api_test

import (
	"cf"
	. "cf/api"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testhelpers"
	"testing"
)

var multipleOrgResponse = testhelpers.TestResponse{Status: http.StatusOK, Body: `
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

var multipleOrgEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/organizations",
	nil,
	multipleOrgResponse,
)

func TestOrganizationsFindAll(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleOrgEndpoint))
	defer ts.Close()

	config := configuration.Configuration{AccessToken: "BEARER my_access_token", Target: ts.URL}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerOrganizationRepository(config, gateway)

	organizations, apiStatus := repo.FindAll()
	assert.False(t, apiStatus.IsError())
	assert.Equal(t, 2, len(organizations))

	firstOrg := organizations[0]
	assert.Equal(t, firstOrg.Name, "Org1")
	assert.Equal(t, firstOrg.Guid, "org1-guid")

	secondOrg := organizations[1]
	assert.Equal(t, secondOrg.Name, "Org2")
	assert.Equal(t, secondOrg.Guid, "org2-guid")
}

func TestOrganizationsFindAllWithIncorrectToken(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleOrgEndpoint))
	defer ts.Close()

	config := configuration.Configuration{AccessToken: "BEARER incorrect_access_token", Target: ts.URL}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerOrganizationRepository(config, gateway)

	var (
		organizations []cf.Organization
		apiStatus     net.ApiStatus
	)

	// Capture output so debugging info does not show up in test
	// output
	testhelpers.CaptureOutput(func() {
		organizations, apiStatus = repo.FindAll()
	})

	assert.True(t, apiStatus.IsError())
	assert.Equal(t, 0, len(organizations))
}

var findOrgByNameResponse = testhelpers.TestResponse{Status: http.StatusOK, Body: `
{
  "total_results": 1,
  "total_pages": 1,
  "prev_url": null,
  "next_url": null,
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

var findOrgByNameEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/organizations?q=name%3Aorg1&inline-relations-depth=1",
	nil,
	findOrgByNameResponse,
)

func TestOrganizationsFindByName(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(findOrgByNameEndpoint))
	defer ts.Close()

	config := configuration.Configuration{AccessToken: "BEARER my_access_token", Target: ts.URL}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerOrganizationRepository(config, gateway)

	existingOrg := cf.Organization{Guid: "org1-guid", Name: "Org1"}

	org, apiStatus := repo.FindByName("Org1")
	assert.False(t, apiStatus.IsError())
	assert.True(t, org.IsFound())
	assert.Equal(t, org.Name, existingOrg.Name)
	assert.Equal(t, org.Guid, existingOrg.Guid)
	assert.Equal(t, len(org.Spaces), 1)
	assert.Equal(t, org.Spaces[0].Name, "Space1")
	assert.Equal(t, org.Spaces[0].Guid, "space1-guid")
	assert.Equal(t, len(org.Domains), 1)
	assert.Equal(t, org.Domains[0].Name, "cfapps.io")
	assert.Equal(t, org.Domains[0].Guid, "domain1-guid")

	org, apiStatus = repo.FindByName("org1")
	assert.False(t, apiStatus.IsError())
	assert.True(t, org.IsFound())
}

var findOrgByNameDoesNotExistResponse = testhelpers.TestResponse{Status: http.StatusOK, Body: `
{
  "total_results": 0,
  "total_pages": 0,
  "prev_url": null,
  "next_url": null,
  "resources": [

  ]
}`}

var findOrgByNameDoesNotExistEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/organizations?q=name%3Aorg1&inline-relations-depth=1",
	nil,
	findOrgByNameDoesNotExistResponse,
)

func TestOrganizationsFindByNameWhenDoesNotExist(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(findOrgByNameDoesNotExistEndpoint))
	defer ts.Close()

	config := configuration.Configuration{AccessToken: "BEARER my_access_token", Target: ts.URL}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerOrganizationRepository(config, gateway)

	org, apiStatus := repo.FindByName("org1")
	assert.False(t, apiStatus.IsError())
	assert.False(t, org.IsFound())
}

var createOrgEndpoint = testhelpers.CreateEndpoint(
	"POST",
	"/v2/organizations",
	testhelpers.RequestBodyMatcher(`{"name":"my-org"}`),
	testhelpers.TestResponse{Status: http.StatusCreated},
)

func TestCreateOrganization(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(createOrgEndpoint))
	defer ts.Close()

	config := configuration.Configuration{AccessToken: "BEARER my_access_token", Target: ts.URL}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerOrganizationRepository(config, gateway)

	apiStatus := repo.Create("my-org")
	assert.False(t, apiStatus.IsError())
}

var renameOrgEndpoint = testhelpers.CreateEndpoint(
	"PUT",
	"/v2/organizations/my-org-guid",
	testhelpers.RequestBodyMatcher(`{"name":"my-new-org"}`),
	testhelpers.TestResponse{Status: http.StatusCreated},
)

func TestRenameOrganization(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(renameOrgEndpoint))
	defer ts.Close()

	config := configuration.Configuration{AccessToken: "BEARER my_access_token", Target: ts.URL}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerOrganizationRepository(config, gateway)

	org := cf.Organization{Guid: "my-org-guid"}
	apiStatus := repo.Rename(org, "my-new-org")
	assert.False(t, apiStatus.IsError())
}

var deleteOrgEndpoint = testhelpers.CreateEndpoint(
	"DELETE",
	"/v2/organizations/my-org-guid?recursive=true",
	nil,
	testhelpers.TestResponse{Status: http.StatusOK},
)

func TestDeleteOrganization(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(deleteOrgEndpoint))
	defer ts.Close()

	config := configuration.Configuration{AccessToken: "BEARER my_access_token", Target: ts.URL}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerOrganizationRepository(config, gateway)

	org := cf.Organization{Guid: "my-org-guid"}
	apiStatus := repo.Delete(org)
	assert.False(t, apiStatus.IsError())
}
