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

func TestFindAllInCurrentSpace(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/spaces/my-space-guid/domains",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{ "resources": [
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
		]}`},
	})

	ts, handler, repo := createDomainRepo(t, req)
	defer ts.Close()

	domains, apiResponse := repo.FindAllInCurrentSpace()
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, 2, len(domains))

	first := domains[0]
	assert.Equal(t, first.Name, "domain1.cf-app.com")
	assert.Equal(t, first.Guid, "domain1-guid")
	second := domains[1]
	assert.Equal(t, second.Name, "domain2.cf-app.com")
	assert.Equal(t, second.Guid, "domain2-guid")
}

var orgDomainsResponse = testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": [
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
]}`}

func TestFindAllByOrg(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/organizations/my-org-guid/domains?inline-relations-depth=1",
		Response: orgDomainsResponse,
	})

	ts, handler, repo := createDomainRepo(t, req)
	defer ts.Close()

	org := cf.Organization{Guid: "my-org-guid"}
	domains, apiResponse := repo.FindAllByOrg(org)

	assert.True(t, handler.AllRequestsCalled())
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

func TestFindByNameInCurrentSpace(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/spaces/my-space-guid/domains?q=name%3Adomain2.cf-app.com",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": [
			{
			  "metadata": { "guid": "domain2-guid" },
			  "entity": { "name": "domain2.cf-app.com" }
			}
		]}`},
	})

	ts, handler, repo := createDomainRepo(t, req)
	defer ts.Close()

	domain, apiResponse := repo.FindByNameInCurrentSpace("domain2.cf-app.com")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())

	assert.Equal(t, domain.Name, "domain2.cf-app.com")
	assert.Equal(t, domain.Guid, "domain2-guid")
}

func TestFindByNameInCurrentSpaceWhenNotFound(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/spaces/my-space-guid/domains?q=name%3Adomain2.cf-app.com",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
	})

	ts, handler, repo := createDomainRepo(t, req)
	defer ts.Close()

	_, apiResponse := repo.FindByNameInCurrentSpace("domain2.cf-app.com")
	assert.True(t, handler.AllRequestsCalled())

	assert.False(t, apiResponse.IsError())
	assert.True(t, apiResponse.IsNotFound())
}

func TestFindByNameInOrg(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/organizations/my-org-guid/domains?inline-relations-depth=1&q=name%3Adomain2.cf-app.com",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": [
			{
			  "metadata": { "guid": "shared-domain-guid" },
			  "entity": {
				"name": "shared-example.com",
				"owning_organization_guid": null,
				"wildcard": true,
				"spaces": [
				  {
					"metadata": { "guid": "my-space-guid" },
					"entity": { "name": "my-space" }
				  }
				]
			  }
			}
		]}`},
	})

	ts, handler, repo := createDomainRepo(t, req)
	defer ts.Close()

	domain, apiResponse := repo.FindByNameInOrg("domain2.cf-app.com", cf.Organization{Guid: "my-org-guid"})
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())

	assert.Equal(t, domain.Name, "shared-example.com")
	assert.Equal(t, domain.Guid, "shared-domain-guid")
	assert.True(t, domain.Shared)
	assert.Equal(t, domain.Spaces[0].Name, "my-space")
}

func TestCreateDomain(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:  "POST",
		Path:    "/v2/domains",
		Matcher: testnet.RequestBodyMatcher(`{"name":"example.com","wildcard":true,"owning_organization_guid":"domain1-guid"}`),
		Response: testnet.TestResponse{Status: http.StatusCreated, Body: `{
			"metadata": { "guid": "abc-123" },
			"entity": { "name": "example.com" }
		}`},
	})

	ts, handler, repo := createDomainRepo(t, req)
	defer ts.Close()

	domainToCreate := cf.Domain{Name: "example.com"}
	owningOrg := cf.Organization{Guid: "domain1-guid"}
	createdDomain, apiResponse := repo.Create(domainToCreate, owningOrg)

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, createdDomain.Guid, "abc-123")
}

func TestShareDomain(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:  "POST",
		Path:    "/v2/domains",
		Matcher: testnet.RequestBodyMatcher(`{"name":"example.com","wildcard":true}`),
		Response: testnet.TestResponse{Status: http.StatusCreated, Body: ` {
			"metadata": { "guid": "abc-123" },
			"entity": { "name": "example.com" }
		}`},
	})

	ts, handler, repo := createDomainRepo(t, req)
	defer ts.Close()

	apiResponse := repo.CreateSharedDomain(cf.Domain{Name: "example.com"})

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

func deleteDomainReq(statusCode int) testnet.TestRequest {
	return testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "DELETE",
		Path:     "/v2/domains/my-domain-guid?recursive=true",
		Response: testnet.TestResponse{Status: statusCode},
	})
}

func TestDeleteDomainSuccess(t *testing.T) {
	req := deleteDomainReq(http.StatusOK)

	ts, handler, repo := createDomainRepo(t, req)
	defer ts.Close()

	domain := cf.Domain{Name: "example.com", Guid: "my-domain-guid"}

	apiResponse := repo.Delete(domain)

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestDeleteDomainFailure(t *testing.T) {
	req := deleteDomainReq(http.StatusBadRequest)

	ts, handler, repo := createDomainRepo(t, req)
	defer ts.Close()

	domain := cf.Domain{Name: "example.com", Guid: "my-domain-guid"}

	apiResponse := repo.Delete(domain)

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsNotSuccessful())
}

func mapDomainReq(statusCode int) testnet.TestRequest {
	return testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     "/v2/spaces/my-space-guid/domains/my-domain-guid",
		Response: testnet.TestResponse{Status: statusCode},
	})
}

func TestMapDomainSuccess(t *testing.T) {
	req := mapDomainReq(http.StatusOK)

	ts, handler, repo := createDomainRepo(t, req)
	defer ts.Close()

	space := cf.Space{Name: "my-space", Guid: "my-space-guid"}
	domain := cf.Domain{Name: "example.com", Guid: "my-domain-guid"}

	apiResponse := repo.Map(domain, space)

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestMapDomainWhenServerError(t *testing.T) {
	req := mapDomainReq(http.StatusBadRequest)

	ts, handler, repo := createDomainRepo(t, req)
	defer ts.Close()

	space := cf.Space{Name: "my-space", Guid: "my-space-guid"}
	domain := cf.Domain{Name: "example.com", Guid: "my-domain-guid"}

	apiResponse := repo.Map(domain, space)

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsNotSuccessful())
}

func unmapDomainReq(statusCode int) testnet.TestRequest {
	return testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "DELETE",
		Path:     "/v2/spaces/my-space-guid/domains/my-domain-guid",
		Response: testnet.TestResponse{Status: statusCode},
	})
}

func TestUnmapDomainSuccess(t *testing.T) {
	req := unmapDomainReq(http.StatusOK)

	ts, handler, repo := createDomainRepo(t, req)
	defer ts.Close()

	space := cf.Space{Name: "my-space", Guid: "my-space-guid"}
	domain := cf.Domain{Name: "example.com", Guid: "my-domain-guid"}

	apiResponse := repo.Unmap(domain, space)

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func createDomainRepo(t *testing.T, req testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo DomainRepository) {
	ts, handler = testnet.NewTLSServer(t, []testnet.TestRequest{req})

	config := &configuration.Configuration{
		AccessToken:  "BEARER my_access_token",
		Target:       ts.URL,
		Space:        cf.Space{Guid: "my-space-guid"},
		Organization: cf.Organization{Guid: "my-org-guid"},
	}
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerDomainRepository(config, gateway)
	return
}
