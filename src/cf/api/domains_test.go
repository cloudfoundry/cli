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

	domains, err := repo.FindAll()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(domains))

	first := domains[0]
	assert.Equal(t, first.Name, "domain1.cf-app.com")
	assert.Equal(t, first.Guid, "domain1-guid")
	second := domains[1]
	assert.Equal(t, second.Name, "domain2.cf-app.com")
	assert.Equal(t, second.Guid, "domain2-guid")
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

	domain, err := repo.FindByName("domain2.cf-app.com")
	assert.NoError(t, err)

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

	domain, err := repo.FindByName("")
	assert.NoError(t, err)

	assert.Equal(t, domain.Name, "domain1.cf-app.com")
	assert.Equal(t, domain.Guid, "domain1-guid")
}
