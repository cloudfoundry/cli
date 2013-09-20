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
    },
    {
      "metadata": {
        "guid": "org2-guid"
      },
      "entity": {
        "name": "Org2",
        "spaces": [
          {
            "metadata": {
              "guid": "space2-guid"
            },
            "entity": {
              "name": "Space2"
            }
          }
        ]
      }
    }
  ]
}`}

var multipleOrgEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/organizations?inline-relations-depth=1",
	nil,
	multipleOrgResponse,
)

func TestOrganizationsFindAll(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleOrgEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: ts.URL}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerOrganizationRepository(config, gateway)

	organizations, err := repo.FindAll()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(organizations))

	firstOrg := organizations[0]
	assert.Equal(t, firstOrg.Name, "Org1")
	assert.Equal(t, firstOrg.Guid, "org1-guid")
	assert.Equal(t, len(firstOrg.Spaces), 1)
	assert.Equal(t, firstOrg.Spaces[0].Name, "Space1")
	assert.Equal(t, firstOrg.Spaces[0].Guid, "space1-guid")
	assert.Equal(t, len(firstOrg.Domains), 1)
	assert.Equal(t, firstOrg.Domains[0].Name, "cfapps.io")
	assert.Equal(t, firstOrg.Domains[0].Guid, "domain1-guid")
	secondOrg := organizations[1]
	assert.Equal(t, secondOrg.Name, "Org2")
	assert.Equal(t, secondOrg.Guid, "org2-guid")
	assert.Equal(t, len(secondOrg.Spaces), 1)
	assert.Equal(t, secondOrg.Spaces[0].Name, "Space2")
	assert.Equal(t, secondOrg.Spaces[0].Guid, "space2-guid")
	assert.Equal(t, len(secondOrg.Domains), 0)
}

func TestOrganizationsFindAllWithIncorrectToken(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleOrgEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{AccessToken: "BEARER incorrect_access_token", Target: ts.URL}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerOrganizationRepository(config, gateway)

	var (
		organizations []cf.Organization
		err           error
	)

	// Capture output so debugging info does not show up in test
	// output
	testhelpers.CaptureOutput(func() {
		organizations, err = repo.FindAll()
	})

	assert.Error(t, err)
	assert.Equal(t, 0, len(organizations))
}

func TestOrganizationsFindByName(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleOrgEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: ts.URL}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerOrganizationRepository(config, gateway)

	existingOrg := cf.Organization{Guid: "org1-guid", Name: "Org1"}

	org, err := repo.FindByName("Org1")
	assert.NoError(t, err)
	assert.Equal(t, org.Name, existingOrg.Name)
	assert.Equal(t, org.Guid, existingOrg.Guid)

	org, err = repo.FindByName("org1")
	assert.NoError(t, err)
	assert.Equal(t, org.Name, existingOrg.Name)
	assert.Equal(t, org.Guid, existingOrg.Guid)

	org, err = repo.FindByName("org that does not exist")
	assert.Error(t, err)
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

	config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: ts.URL}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerOrganizationRepository(config, gateway)

	err := repo.Create("my-org")
	assert.NoError(t, err)
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

	config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: ts.URL}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerOrganizationRepository(config, gateway)

	org := cf.Organization{Guid: "my-org-guid"}
	err := repo.Rename(org, "my-new-org")
	assert.NoError(t, err)
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

	config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: ts.URL}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerOrganizationRepository(config, gateway)

	org := cf.Organization{Guid: "my-org-guid"}
	err := repo.Delete(org)
	assert.NoError(t, err)
}
