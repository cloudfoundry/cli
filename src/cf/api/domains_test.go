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

var multipleDomainsEndpoint, multipleDomainsEndpointStatus = testhelpers.CreateCheckableEndpoint(
	"GET",
	"/v2/spaces/my-space-guid/domains",
	nil,
	multipleDomainsResponse,
)

func TestFindAllInCurrentSpace(t *testing.T) {
	multipleDomainsEndpointStatus.Reset()

	ts, repo := createDomainRepo(multipleDomainsEndpoint)
	defer ts.Close()

	domains, apiResponse := repo.FindAllInCurrentSpace()
	assert.True(t, multipleDomainsEndpointStatus.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
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

var orgDomainsEndpoint, orgDomainsEndpointStatus = testhelpers.CreateCheckableEndpoint(
	"GET",
	"/v2/organizations/my-org-guid/domains?inline-relations-depth=1",
	nil,
	orgDomainsResponse,
)

func TestFindAllByOrg(t *testing.T) {
	orgDomainsEndpointStatus.Reset()

	ts, repo := createDomainRepo(orgDomainsEndpoint)
	defer ts.Close()

	org := cf.Organization{Guid: "my-org-guid"}
	domains, apiResponse := repo.FindAllByOrg(org)

	assert.True(t, orgDomainsEndpointStatus.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, 2, len(domains))

	domain := domains[0]
	assert.True(t, domain.Shared)

	domain = domains[1]
	assert.Equal(t, domain.Name, "example.com")
	assert.Equal(t, domain.Guid, "my-domain-guid")
	assert.False(t, domain.Shared)
	assert.Equal(t, domain.Spaces[0].Name, "my-space")
}

func TestFindByNameInCurrentSpaceReturnsTheDomainMatchingTheName(t *testing.T) {
	multipleDomainsEndpointStatus.Reset()

	ts, repo := createDomainRepo(multipleDomainsEndpoint)
	defer ts.Close()

	domain, apiResponse := repo.FindByNameInCurrentSpace("domain2.cf-app.com")
	assert.True(t, multipleDomainsEndpointStatus.Called())
	assert.False(t, apiResponse.IsNotSuccessful())

	assert.Equal(t, domain.Name, "domain2.cf-app.com")
	assert.Equal(t, domain.Guid, "domain2-guid")
}

func TestFindByNameInCurrentSpaceReturnsTheFirstDomainIfNameEmpty(t *testing.T) {
	multipleDomainsEndpointStatus.Reset()

	ts, repo := createDomainRepo(multipleDomainsEndpoint)
	defer ts.Close()

	_, apiResponse := repo.FindByNameInCurrentSpace("")
	assert.True(t, multipleDomainsEndpointStatus.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestFindByNameInCurrentSpaceReturnsNotFoundIfNameEmptyAndNoDomains(t *testing.T) {
	endpoint, status := testhelpers.CreateCheckableEndpoint(
		"GET",
		"/v2/spaces/my-space-guid/domains",
		nil,
		testhelpers.TestResponse{Status: http.StatusOK, Body: `
{
  "resources": []
}
`},
	)

	ts, repo := createDomainRepo(endpoint)
	defer ts.Close()

	_, apiResponse := repo.FindByNameInCurrentSpace("")
	assert.True(t, status.Called())
	assert.False(t, apiResponse.IsError())
	assert.True(t, apiResponse.IsNotFound())
}

func TestFindByNameInCurrentSpaceWhenTheDomainIsNotFound(t *testing.T) {
	multipleDomainsEndpointStatus.Reset()

	ts, repo := createDomainRepo(multipleDomainsEndpoint)
	defer ts.Close()

	_, apiResponse := repo.FindByNameInCurrentSpace("domain3.cf-app.com")
	assert.True(t, multipleDomainsEndpointStatus.Called())
	assert.False(t, apiResponse.IsError())
	assert.True(t, apiResponse.IsNotFound())
}

func TestCreateDomain(t *testing.T) {
	createDomainResponse := `
{
    "metadata": {
        "guid": "abc-123"
    },
    "entity": {
        "name": "example.com"
    }
}`

	endpoint, status := testhelpers.CreateCheckableEndpoint(
		"POST",
		"/v2/domains",
		testhelpers.RequestBodyMatcher(`{"name":"example.com","wildcard":true,"owning_organization_guid":"domain1-guid"}`),
		testhelpers.TestResponse{Status: http.StatusCreated, Body: createDomainResponse},
	)

	ts, repo := createDomainRepo(endpoint)
	defer ts.Close()

	domainToCreate := cf.Domain{Name: "example.com"}
	owningOrg := cf.Organization{Guid: "domain1-guid"}
	createdDomain, apiResponse := repo.Create(domainToCreate, owningOrg)
	assert.True(t, status.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, createdDomain.Guid, "abc-123")
}

func TestShareDomain(t *testing.T) {
	shareDomainResponse := `
{
    "metadata": {
        "guid": "abc-123"
    },
    "entity": {
        "name": "example.com"
    }
}`

	endpoint, status := testhelpers.CreateCheckableEndpoint(
		"POST",
		"/v2/domains",
		testhelpers.RequestBodyMatcher(`{"name":"example.com","wildcard":true,"shared":true}`),
		testhelpers.TestResponse{Status: http.StatusCreated, Body: shareDomainResponse},
	)

	ts, repo := createDomainRepo(endpoint)
	defer ts.Close()

	domainToShare := cf.Domain{Name: "example.com"}
	apiResponse := repo.Share(domainToShare)
	assert.True(t, status.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestFindByNameInOrgWhenDomainExists(t *testing.T) {
	orgDomainsEndpointStatus.Reset()

	ts, repo := createDomainRepo(orgDomainsEndpoint)
	defer ts.Close()

	domainName := "example.com"
	org := cf.Organization{Name: "my-org", Guid: "my-org-guid"}
	domain, apiResponse := repo.FindByNameInOrg(domainName, org)

	assert.True(t, orgDomainsEndpointStatus.Called())
	assert.Equal(t, domain.Name, domainName)
	assert.Equal(t, domain.Guid, "my-domain-guid")
	assert.False(t, apiResponse.IsNotSuccessful())
}

func mapDomainEndpoint(statusCode int) (hf http.HandlerFunc, status *testhelpers.RequestStatus) {
	hf, status = testhelpers.CreateCheckableEndpoint(
		"PUT",
		"/v2/spaces/my-space-guid/domains/my-domain-guid",
		nil,
		testhelpers.TestResponse{Status: statusCode},
	)
	return
}

func TestMapDomainSuccess(t *testing.T) {
	endpoint, reqStatus := mapDomainEndpoint(http.StatusOK)

	ts, repo := createDomainRepo(endpoint)
	defer ts.Close()

	space := cf.Space{Name: "my-space", Guid: "my-space-guid"}
	domain := cf.Domain{Name: "example.com", Guid: "my-domain-guid"}

	apiResponse := repo.MapDomain(domain, space)

	assert.True(t, reqStatus.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestMapDomainWhenServerError(t *testing.T) {
	endpoint, reqStatus := mapDomainEndpoint(http.StatusBadRequest)

	ts, repo := createDomainRepo(endpoint)
	defer ts.Close()

	space := cf.Space{Name: "my-space", Guid: "my-space-guid"}
	domain := cf.Domain{Name: "example.com", Guid: "my-domain-guid"}

	apiResponse := repo.MapDomain(domain, space)

	assert.True(t, reqStatus.Called())
	assert.True(t, apiResponse.IsNotSuccessful())
}

func unmapDomainEndpoint(statusCode int) (hf http.HandlerFunc, status *testhelpers.RequestStatus) {
	hf, status = testhelpers.CreateCheckableEndpoint(
		"DELETE",
		"/v2/spaces/my-space-guid/domains/my-domain-guid",
		nil,
		testhelpers.TestResponse{Status: statusCode},
	)
	return
}

func TestUnmapDomainSuccess(t *testing.T) {
	endpoint, reqStatus := unmapDomainEndpoint(http.StatusOK)

	ts, repo := createDomainRepo(endpoint)
	defer ts.Close()

	space := cf.Space{Name: "my-space", Guid: "my-space-guid"}
	domain := cf.Domain{Name: "example.com", Guid: "my-domain-guid"}

	apiResponse := repo.UnmapDomain(domain, space)

	assert.True(t, reqStatus.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func deleteDomainEndpoint(statusCode int) (hf http.HandlerFunc, status *testhelpers.RequestStatus) {
	hf, status = testhelpers.CreateCheckableEndpoint(
		"DELETE",
		"/v2/domains/my-domain-guid?recursive=true",
		nil,
		testhelpers.TestResponse{Status: statusCode},
	)
	return
}

func TestDeleteDomainSuccess(t *testing.T) {
	endpoint, reqStatus := deleteDomainEndpoint(http.StatusOK)

	ts, repo := createDomainRepo(endpoint)
	defer ts.Close()

	domain := cf.Domain{Name: "example.com", Guid: "my-domain-guid"}

	apiResponse := repo.DeleteDomain(domain)

	assert.True(t, reqStatus.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestDeleteDomainFailure(t *testing.T) {
	endpoint, reqStatus := deleteDomainEndpoint(http.StatusBadRequest)

	ts, repo := createDomainRepo(endpoint)
	defer ts.Close()

	domain := cf.Domain{Name: "example.com", Guid: "my-domain-guid"}

	apiResponse := repo.DeleteDomain(domain)

	assert.True(t, reqStatus.Called())
	assert.True(t, apiResponse.IsNotSuccessful())
}

func createDomainRepo(endpoint http.HandlerFunc) (ts *httptest.Server, repo DomainRepository) {
	ts = httptest.NewTLSServer(endpoint)

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
		Organization:  cf.Organization{Guid: "my-org-guid"},
	}
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerDomainRepository(config, gateway)
	return
}
