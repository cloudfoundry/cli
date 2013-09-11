package api_test

import (
	"cf"
	. "cf/api"
	"cf/configuration"
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

	config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: ts.URL}
	client := NewApiClient(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerOrganizationRepository(config, client)

	organizations, err := repo.FindAll()
	assert.NoError(t, err)
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

	config := &configuration.Configuration{AccessToken: "BEARER incorrect_access_token", Target: ts.URL}
	client := NewApiClient(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerOrganizationRepository(config, client)

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
	client := NewApiClient(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerOrganizationRepository(config, client)

	existingOrg := cf.Organization{Guid: "org1-guid", Name: "Org1"}

	org, err := repo.FindByName("Org1")
	assert.NoError(t, err)
	assert.Equal(t, org, existingOrg)

	org, err = repo.FindByName("org1")
	assert.NoError(t, err)
	assert.Equal(t, org, existingOrg)

	org, err = repo.FindByName("org that does not exist")
	assert.Error(t, err)
}

var createOrgResponse = testhelpers.TestResponse{Status: http.StatusCreated, Body: `
{
  "metadata": {
    "guid": "my-org-guid",
    "url": "/v2/organizations/my-org-guid",
    "created_at": "2013-09-11 06:13:15 +0000",
    "updated_at": null
  },
  "entity": {
    "name": "my-org",
    "billing_enabled": false,
    "quota_definition_guid": "f95878e4-181b-4016-af49-b8e2fc521cb7",
    "status": "active",
    "quota_definition_url": "/v2/quota_definitions/f95878e4-181b-4016-af49-b8e2fc521cb7",
    "spaces_url": "/v2/organizations/my-org-guid/spaces",
    "domains_url": "/v2/organizations/my-org-guid/domains",
    "users_url": "/v2/organizations/my-org-guid/users",
    "managers_url": "/v2/organizations/my-org-guid/managers",
    "billing_managers_url": "/v2/organizations/my-org-guid/billing_managers",
    "auditors_url": "/v2/organizations/my-org-guid/auditors",
    "app_events_url": "/v2/organizations/my-org-guid/app_events"
  }
}`}

var createOrgEndpoint = testhelpers.CreateEndpoint(
	"POST",
	"/v2/organizations",
	testhelpers.RequestBodyMatcher(`{"name":"my-org"}`),
	createOrgResponse,
)

func TestCreateOrganization(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(createOrgEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: ts.URL}
	client := NewApiClient(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerOrganizationRepository(config, client)

	createdOrg, err := repo.Create("my-org")
	assert.NoError(t, err)

	assert.Equal(t, createdOrg.Name, "my-org")
	assert.Equal(t, createdOrg.Guid, "my-org-guid")
}
