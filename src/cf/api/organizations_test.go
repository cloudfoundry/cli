package api

import (
	"cf/configuration"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var multipleOrgEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	acceptHeaderMatches := request.Header.Get("accept") == "application/json"
	methodMatches := request.Method == "GET"
	pathMatches := request.URL.Path == "/v2/organizations"
	authMatches := request.Header.Get("authorization") == "BEARER my_access_token"

	if !(acceptHeaderMatches && methodMatches && pathMatches && authMatches) {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonResponse := `

{
  "total_results": 2,
  "total_pages": 1,
  "prev_url": null,
  "next_url": null,
  "resources": [
    {
      "metadata": {
        "guid": "c4acbb8c-956c-49c8-abe4-b81881ad4138"
      },
      "entity": {
        "name": "Org1"
      }
    },
    {
      "metadata": {
        "guid": "ac411d31-5e3b-4bea-ba5e-a0540627d1e7"
      },
      "entity": {
        "name": "Org2"
      }
    }
  ]
}`
	fmt.Fprintln(writer, jsonResponse)
}

func TestFindOrganizations(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleOrgEndpoint))
	defer ts.Close()

	repo := CloudControllerOrganizationRepository{}

	config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: ts.URL}
	organizations, err := repo.FindOrganizations(config)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(organizations))
}

func TestFindOrganizationsWithIncorrectToken(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleOrgEndpoint))
	defer ts.Close()

	repo := CloudControllerOrganizationRepository{}

	config := &configuration.Configuration{AccessToken: "BEARER incorrect_access_token", Target: ts.URL}
	organizations, err := repo.FindOrganizations(config)

	assert.Error(t, err)
	assert.Equal(t, 0, len(organizations))
}

func TestOrganizationExists(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleOrgEndpoint))
	defer ts.Close()

	repo := CloudControllerOrganizationRepository{}

	config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: ts.URL}
	existingOrg := Organization{Name: "Org1"}
	nonexistingOrg := Organization{Name: "Org3"}

	assert.True(t, repo.OrganizationExists(config, existingOrg))
	assert.False(t, repo.OrganizationExists(config, nonexistingOrg))
}
