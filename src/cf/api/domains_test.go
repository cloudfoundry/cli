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

var multipleDomainsResponse = testhelpers.TestResponse{Status: http.StatusOK, Body: `
{
  "total_results": 2,
  "total_pages": 1,
  "prev_url": null,
  "next_url": null,
  "resources": [
    {
      "metadata": {
        "guid": "domain1-guid"
      },
      "entity": {
        "name": "domain1.cf-app.com"
      }
    },
    {
      "metadata": {
        "guid": "domain2-guid"
      },
      "entity": {
        "name": "domain2.cf-app.com"
      }
    }
  ]
}`}

var multipleDomainsEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/spaces/my-space-guid/domains",
	nil,
	multipleDomainsResponse,
)

func TestFindAll(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleDomainsEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
	}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerDomainRepository(config, gateway)

	domains, apiStatus := repo.FindAll()
	assert.False(t, apiStatus.IsError())
	assert.Equal(t, 2, len(domains))

	first := domains[0]
	assert.Equal(t, first.Name, "domain1.cf-app.com")
	assert.Equal(t, first.Guid, "domain1-guid")
	second := domains[1]
	assert.Equal(t, second.Name, "domain2.cf-app.com")
	assert.Equal(t, second.Guid, "domain2-guid")
}

var orgDomainsResponse = testhelpers.TestResponse{Status: http.StatusOK, Body: `
{
  "resources": [
    {
      "metadata": {
        "guid": "shared-domain-guid"
      },
      "entity": {
        "name": "shared-example.com",
        "owning_organization_guid": null,
        "wildcard": true,
        "spaces": [
          {
            "metadata": {
              "guid": "my-space-guid"
            },
            "entity": {
              "name": "my-space"
            }
          }
        ]
      }
    },
    {
      "metadata": {
        "guid": "my-domain-guid"
      },
      "entity": {
        "name": "example.com",
        "owning_organization_guid": "my-org-guid",
        "wildcard": true,
        "spaces": [
          {
            "metadata": {
              "guid": "my-space-guid"
            },
            "entity": {
              "name": "my-space"
            }
          }
        ]
      }
    }
  ]
}
`}

var orgDomainsEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/organizations/my-org-guid/domains?inline-relations-depth=1",
	nil,
	orgDomainsResponse,
)

func TestFindAllByOrg(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(orgDomainsEndpoint))
	defer ts.Close()

	org := cf.Organization{Guid: "my-org-guid"}
	config := &configuration.Configuration{
		AccessToken:  "BEARER my_access_token",
		Target:       ts.URL,
		Organization: org,
	}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerDomainRepository(config, gateway)

	domains, apiStatus := repo.FindAllByOrg(org)

	assert.False(t, apiStatus.IsError())
	assert.Equal(t, 2, len(domains))

	domain := domains[0]
	assert.True(t, domain.Shared)

	domain = domains[1]
	assert.Equal(t, domain.Name, "example.com")
	assert.Equal(t, domain.Guid, "my-domain-guid")
	assert.False(t, domain.Shared)
	assert.Equal(t, domain.Spaces[0].Name, "my-space")
}

func TestFindByNameReturnsTheDomainMatchingTheName(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleDomainsEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
	}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerDomainRepository(config, gateway)

	domain, apiStatus := repo.FindByName("domain2.cf-app.com")
	assert.False(t, apiStatus.IsError())

	assert.Equal(t, domain.Name, "domain2.cf-app.com")
	assert.Equal(t, domain.Guid, "domain2-guid")
}

func TestFindByNameReturnsTheFirstDomainIfNameEmpty(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleDomainsEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
	}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerDomainRepository(config, gateway)

	domain, apiStatus := repo.FindByName("")
	assert.False(t, apiStatus.IsError())

	assert.Equal(t, domain.Name, "domain1.cf-app.com")
	assert.Equal(t, domain.Guid, "domain1-guid")
}

func TestFindByNameWhenTheDomainIsNotFound(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleDomainsEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
	}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerDomainRepository(config, gateway)

	_, apiStatus := repo.FindByName("domain3.cf-app.com")
	assert.False(t, apiStatus.IsError())
	assert.True(t, apiStatus.IsNotFound())
}

var createDomainResponse = `
{
    "metadata": {
        "guid": "abc-123"
    },
    "entity": {
        "name": "example.com"
    }
}`

var createDomainEndpoint = testhelpers.CreateEndpoint(
	"POST",
	"/v2/domains",
	testhelpers.RequestBodyMatcher(`{"name":"example.com","wildcard":true,"owning_organization_guid":"domain1-guid"}`),
	testhelpers.TestResponse{Status: http.StatusCreated, Body: createDomainResponse},
)

func TestParkDomain(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(createDomainEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
	}

	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerDomainRepository(config, gateway)

	domainToCreate := cf.Domain{Name: "example.com"}
	owningOrg := cf.Organization{Guid: "domain1-guid"}
	createdDomain, apiStatus := repo.Create(domainToCreate, owningOrg)
	assert.False(t, apiStatus.IsError())
	assert.Equal(t, createdDomain.Guid, "abc-123")
}
