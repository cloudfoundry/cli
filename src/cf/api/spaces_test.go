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

var multipleSpacesResponse = testhelpers.TestResponse{Status: http.StatusOK, Body: `
{
  "resources": [
    {
      "metadata": {
        "guid": "acceptance-space-guid"
      },
      "entity": {
        "name": "acceptance"
      }
    },
    {
      "metadata": {
        "guid": "staging-space-guid"
      },
      "entity": {
        "name": "staging"
      }
    }
  ]
}`}

var multipleSpacesEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/organizations/some-org-guid/spaces",
	nil,
	multipleSpacesResponse,
)

func TestSpacesFindAll(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleSpacesEndpoint))
	defer ts.Close()

	repo := CloudControllerSpaceRepository{}
	config := &configuration.Configuration{
		AccessToken:  "BEARER my_access_token",
		Target:       ts.URL,
		Organization: cf.Organization{Guid: "some-org-guid"},
	}
	spaces, err := repo.FindAll(config)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(spaces))

	firstSpace := spaces[0]
	assert.Equal(t, firstSpace.Name, "acceptance")
	assert.Equal(t, firstSpace.Guid, "acceptance-space-guid")

	secondSpace := spaces[1]
	assert.Equal(t, secondSpace.Name, "staging")
	assert.Equal(t, secondSpace.Guid, "staging-space-guid")
}

func TestSpacesFindAllWithIncorrectToken(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleSpacesEndpoint))
	defer ts.Close()

	repo := CloudControllerSpaceRepository{}

	config := &configuration.Configuration{
		AccessToken:  "BEARER incorrect_access_token",
		Target:       ts.URL,
		Organization: cf.Organization{Guid: "some-org-guid"},
	}
	spaces, err := repo.FindAll(config)

	assert.Error(t, err)
	assert.Equal(t, 0, len(spaces))
}

func TestSpacesFindByName(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleSpacesEndpoint))
	defer ts.Close()

	repo := CloudControllerSpaceRepository{}
	config := &configuration.Configuration{
		AccessToken:  "BEARER my_access_token",
		Target:       ts.URL,
		Organization: cf.Organization{Guid: "some-org-guid"},
	}
	existingOrg := cf.Space{Guid: "staging-space-guid", Name: "staging"}

	org, err := repo.FindByName(config, "staging")
	assert.NoError(t, err)
	assert.Equal(t, org, existingOrg)

	org, err = repo.FindByName(config, "Staging")
	assert.NoError(t, err)
	assert.Equal(t, org, existingOrg)

	org, err = repo.FindByName(config, "space that does not exist")
	assert.Error(t, err)
}
