package api_test

import (
	"cf"
	. "cf/api"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testnet "testhelpers/net"
	"testing"
)

var noDomainsRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/domains?inline-relations-depth=1",
	Response: testnet.TestResponse{Status: http.StatusOK, Body: `
{
	"next_url": "",
	"resources": []
}`},
})

var firstPageSharedDomainsRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/shared_domains?inline-relations-depth=1",
	Response: testnet.TestResponse{Status: http.StatusOK, Body: `
{
	"next_url": "/v2/shared_domains?inline-relations-depth=1&page=2",
	"resources": [
		{
		  "metadata": {
			"guid": "shared-domain1-guid"
		  },
		  "entity": {
			"name": "shared-example1.com"
 		  }
		}
	]
}`},
})

var secondPageSharedDomainsRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/shared_domains?inline-relations-depth=1",
	Response: testnet.TestResponse{Status: http.StatusOK, Body: `
{
	"resources": [
		{
		  "metadata": {
			"guid": "shared-domain2-guid"
		  },
		  "entity": {
			"name": "shared-example2.com"
 		  }
		}
	]
}`},
})

var firstPageDomainsRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/domains?inline-relations-depth=1",
	Response: testnet.TestResponse{Status: http.StatusOK, Body: `{
		"next_url": "/v2/domains?inline-relations-depth=1&page=2",
		"resources": [
	{
      "metadata": {
        "guid": "domain1-guid"
      },
      "entity": {
        "name": "example.com",
        "owning_organization_guid": "my-org-guid",
        "spaces": [
          {
            "metadata": { "guid": "my-space-guid" },
            "entity": { "name": "my-space" }
          }
        ]
      }
    },
    {
      "metadata": {
        "guid": "domain2-guid"
      },
      "entity": {
        "name": "some-shared.example.com",
        "owning_organization_guid": null,
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

var secondPageDomainsRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/domains?inline-relations-depth=1&page=2",
	Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": [
    {
      "metadata": {
        "guid": "not-in-my-org-domain-guid"
      },
      "entity": {
        "name": "example.com",
        "owning_organization_guid": "not-my-org-guid",
        "spaces": []
      }
    },
	{
      "metadata": {
        "guid": "domain3-guid"
      },
      "entity": {
        "name": "example.com",
        "owning_organization_guid": "my-org-guid",
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

func TestListSharedDomains(t *testing.T) {
	ts, handler, repo := createDomainRepo(t, []testnet.TestRequest{firstPageSharedDomainsRequest, secondPageSharedDomainsRequest})
	defer ts.Close()

	receivedDomains := []cf.Domain{}
	apiResponse := repo.ListSharedDomains(ListDomainsCallback(func(domains []cf.Domain) bool {
		receivedDomains = append(receivedDomains, domains...)
		return true
	}))

	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, len(receivedDomains), 2)
	assert.Equal(t, receivedDomains[0].Guid, "shared-domain1-guid")
	assert.Equal(t, receivedDomains[1].Guid, "shared-domain2-guid")
	assert.True(t, handler.AllRequestsCalled())
}

func TestDomainListDomainsForOrg(t *testing.T) {
	ts, handler, repo := createDomainRepo(t, []testnet.TestRequest{firstPageDomainsRequest, secondPageDomainsRequest})
	defer ts.Close()

	receivedDomains := []cf.Domain{}
	apiResponse := repo.ListDomainsForOrg("my-org-guid", ListDomainsCallback(func(domains []cf.Domain) bool {
		receivedDomains = append(receivedDomains, domains...)
		return true
	}))

	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, len(receivedDomains), 3)
	assert.Equal(t, receivedDomains[0].Guid, "domain1-guid")
	assert.Equal(t, receivedDomains[1].Guid, "domain2-guid")
	assert.True(t, handler.AllRequestsCalled())
}

func TestListDomainsForOrgWithNoDomains(t *testing.T) {
	ts, handler, repo := createDomainRepo(t, []testnet.TestRequest{noDomainsRequest})
	defer ts.Close()

	wasCalled := false
	apiResponse := repo.ListDomainsForOrg("my-org-guid", ListDomainsCallback(func(domains []cf.Domain) bool {
		wasCalled = true
		return true
	}))

	assert.True(t, apiResponse.IsSuccessful())
	assert.False(t, wasCalled)
	assert.True(t, handler.AllRequestsCalled())
}

func TestDomainListDomainsForOrgWithNoDomains(t *testing.T) {
	emptyDomainsRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/domains?inline-relations-depth=1",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": [] }`},
	})

	ts, handler, repo := createDomainRepo(t, []testnet.TestRequest{emptyDomainsRequest})
	defer ts.Close()

	receivedDomains := []cf.Domain{}
	apiResponse := repo.ListDomainsForOrg("my-org-guid", ListDomainsCallback(func(domains []cf.Domain) bool {
		receivedDomains = append(receivedDomains, domains...)
		return true
	}))

	assert.True(t, apiResponse.IsSuccessful())
	assert.True(t, handler.AllRequestsCalled())
}

func TestDomainFindByName(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/domains?inline-relations-depth=1&q=name%3Adomain2.cf-app.com",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": [
			{
			  "metadata": { "guid": "domain2-guid" },
			  "entity": { "name": "domain2.cf-app.com" }
			}
		]}`},
	})

	ts, handler, repo := createDomainRepo(t, []testnet.TestRequest{req})
	defer ts.Close()

	domain, apiResponse := repo.FindByName("domain2.cf-app.com")
	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())

	assert.Equal(t, domain.Name, "domain2.cf-app.com")
	assert.Equal(t, domain.Guid, "domain2-guid")
}

func TestDomainFindByNameInCurrentSpace(t *testing.T) {
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

	ts, handler, repo := createDomainRepo(t, []testnet.TestRequest{req})
	defer ts.Close()

	domain, apiResponse := repo.FindByNameInCurrentSpace("domain2.cf-app.com")
	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())

	assert.Equal(t, domain.Name, "domain2.cf-app.com")
	assert.Equal(t, domain.Guid, "domain2-guid")
}

func TestDomainFindByNameInCurrentSpaceWhenNotFound(t *testing.T) {
	spaceDomainsReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/spaces/my-space-guid/domains?q=name%3Adomain2.cf-app.com",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
	})

	sharedDomainsReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/domains?q=name%3Adomain2.cf-app.com",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
	})

	ts, handler, repo := createDomainRepo(t, []testnet.TestRequest{spaceDomainsReq, sharedDomainsReq})
	defer ts.Close()

	_, apiResponse := repo.FindByNameInCurrentSpace("domain2.cf-app.com")
	assert.True(t, handler.AllRequestsCalled())

	assert.False(t, apiResponse.IsError())
	assert.True(t, apiResponse.IsNotFound())
}

func TestDomainFindByNameInCurrentSpaceWhenFoundAsSharedDomain(t *testing.T) {
	spaceDomainsReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/spaces/my-space-guid/domains?q=name%3Adomain2.cf-app.com",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
	})

	sharedDomainsReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/domains?q=name%3Adomain2.cf-app.com",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": [
			{
			  "metadata": { "guid": "shared-domain-guid" },
			  "entity": {
			    "name": "shared-domain.cf-app.com",
				"owning_organization_guid": null
			  }
			}
		]}`},
	})

	ts, handler, repo := createDomainRepo(t, []testnet.TestRequest{spaceDomainsReq, sharedDomainsReq})
	defer ts.Close()

	domain, apiResponse := repo.FindByNameInCurrentSpace("domain2.cf-app.com")
	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())

	assert.Equal(t, domain.Name, "shared-domain.cf-app.com")
	assert.Equal(t, domain.Guid, "shared-domain-guid")
}

func TestDomainFindByNameInCurrentSpaceWhenFoundInDomainsButNotShared(t *testing.T) {
	spaceDomainsReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/spaces/my-space-guid/domains?q=name%3Adomain2.cf-app.com",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
	})

	sharedDomainsReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/domains?q=name%3Adomain2.cf-app.com",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": [
			{
			  "metadata": { "guid": "some-domain-guid" },
			  "entity": {
			    "name": "some.cf-app.com",
				"owning_organization_guid": "some-org-guid"
			  }
			}
		]}`},
	})

	ts, handler, repo := createDomainRepo(t, []testnet.TestRequest{spaceDomainsReq, sharedDomainsReq})
	defer ts.Close()

	_, apiResponse := repo.FindByNameInCurrentSpace("domain2.cf-app.com")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsError())
	assert.True(t, apiResponse.IsNotFound())
}

func TestDomainFindByNameInOrg(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/organizations/my-org-guid/domains?inline-relations-depth=1&q=name%3Adomain2.cf-app.com",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": [
			{
			  "metadata": { "guid": "my-domain-guid" },
			  "entity": {
				"name": "my-example.com",
				"owning_organization_guid": "my-org-guid",
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

	ts, handler, repo := createDomainRepo(t, []testnet.TestRequest{req})
	defer ts.Close()

	domain, apiResponse := repo.FindByNameInOrg("domain2.cf-app.com", "my-org-guid")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())

	assert.Equal(t, domain.Name, "my-example.com")
	assert.Equal(t, domain.Guid, "my-domain-guid")
	assert.False(t, domain.Shared)
	assert.Equal(t, domain.Spaces[0].Name, "my-space")
}

func TestDomainFindByNameInOrgWhenNotFoundOnBothEndpoints(t *testing.T) {
	orgDomainsReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/organizations/my-org-guid/domains?inline-relations-depth=1&q=name%3Adomain2.cf-app.com",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
	})

	sharedDomainsReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/domains?inline-relations-depth=1&q=name%3Adomain2.cf-app.com",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
	})

	ts, handler, repo := createDomainRepo(t, []testnet.TestRequest{orgDomainsReq, sharedDomainsReq})
	defer ts.Close()

	_, apiResponse := repo.FindByNameInOrg("domain2.cf-app.com", "my-org-guid")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsError())
	assert.True(t, apiResponse.IsNotFound())
}

func TestDomainFindByNameInOrgWhenFoundAsSharedDomain(t *testing.T) {
	orgDomainsReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/organizations/my-org-guid/domains?inline-relations-depth=1&q=name%3Adomain2.cf-app.com",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
	})

	sharedDomainsReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/domains?inline-relations-depth=1&q=name%3Adomain2.cf-app.com",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": [
			{
			  "metadata": { "guid": "shared-domain-guid" },
			  "entity": {
				"name": "shared-example.com",
				"owning_organization_guid": null,
				"wildcard": true,
				"spaces": []
			  }
			}
		]}`},
	})

	ts, handler, repo := createDomainRepo(t, []testnet.TestRequest{orgDomainsReq, sharedDomainsReq})
	defer ts.Close()

	domain, apiResponse := repo.FindByNameInOrg("domain2.cf-app.com", "my-org-guid")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())

	assert.Equal(t, domain.Name, "shared-example.com")
	assert.Equal(t, domain.Guid, "shared-domain-guid")
	assert.True(t, domain.Shared)
}

func TestDomainFindByNameInOrgWhenFoundInDomainsButNotShared(t *testing.T) {
	orgDomainsReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/organizations/my-org-guid/domains?inline-relations-depth=1&q=name%3Adomain2.cf-app.com",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
	})

	sharedDomainsReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/domains?inline-relations-depth=1&q=name%3Adomain2.cf-app.com",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": [
			{
			  "metadata": { "guid": "shared-domain-guid" },
			  "entity": {
				"name": "shared-example.com",
				"owning_organization_guid": "some-other-org-guid",
				"wildcard": true,
				"spaces": []
			  }
			}
		]}`},
	})

	ts, handler, repo := createDomainRepo(t, []testnet.TestRequest{orgDomainsReq, sharedDomainsReq})
	defer ts.Close()

	_, apiResponse := repo.FindByNameInOrg("domain2.cf-app.com", "my-org-guid")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsError())
	assert.True(t, apiResponse.IsNotFound())
}

func TestCreateDomain(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:  "POST",
		Path:    "/v2/domains",
		Matcher: testnet.RequestBodyMatcher(`{"name":"example.com","wildcard":true,"owning_organization_guid":"org-guid"}`),
		Response: testnet.TestResponse{Status: http.StatusCreated, Body: `{
			"metadata": { "guid": "abc-123" },
			"entity": { "name": "example.com" }
		}`},
	})

	ts, handler, repo := createDomainRepo(t, []testnet.TestRequest{req})
	defer ts.Close()

	createdDomain, apiResponse := repo.Create("example.com", "org-guid")

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

	ts, handler, repo := createDomainRepo(t, []testnet.TestRequest{req})
	defer ts.Close()

	apiResponse := repo.CreateSharedDomain("example.com")

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

	ts, handler, repo := createDomainRepo(t, []testnet.TestRequest{req})
	defer ts.Close()

	apiResponse := repo.Delete("my-domain-guid")

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestDeleteDomainFailure(t *testing.T) {
	req := deleteDomainReq(http.StatusBadRequest)

	ts, handler, repo := createDomainRepo(t, []testnet.TestRequest{req})
	defer ts.Close()

	apiResponse := repo.Delete("my-domain-guid")

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

	ts, handler, repo := createDomainRepo(t, []testnet.TestRequest{req})
	defer ts.Close()

	apiResponse := repo.Map("my-domain-guid", "my-space-guid")

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestMapDomainWhenServerError(t *testing.T) {
	req := mapDomainReq(http.StatusBadRequest)

	ts, handler, repo := createDomainRepo(t, []testnet.TestRequest{req})
	defer ts.Close()

	apiResponse := repo.Map("my-domain-guid", "my-space-guid")

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

	ts, handler, repo := createDomainRepo(t, []testnet.TestRequest{req})
	defer ts.Close()

	apiResponse := repo.Unmap("my-domain-guid", "my-space-guid")

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func createDomainRepo(t *testing.T, reqs []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo DomainRepository) {
	ts, handler = testnet.NewTLSServer(t, reqs)
	org := cf.OrganizationFields{}
	org.Guid = "my-org-guid"
	space := cf.SpaceFields{}
	space.Guid = "my-space-guid"

	config := &configuration.Configuration{
		AccessToken:        "BEARER my_access_token",
		Target:             ts.URL,
		SpaceFields:        space,
		OrganizationFields: org,
	}
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerDomainRepository(config, gateway)
	return
}
