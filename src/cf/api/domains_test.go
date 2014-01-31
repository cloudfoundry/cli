package api_test

import (
	"cf"
	. "cf/api"
	"cf/configuration"
	"cf/net"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testnet "testhelpers/net"
)

var noDomainsRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/organizations/my-org-guid/private_domains",
	Response: testnet.TestResponse{Status: http.StatusOK, Body: `
{
	"next_url": "",
	"resources": []
}`},
})

var firstPageSharedDomainsRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/shared_domains",
	Response: testnet.TestResponse{Status: http.StatusOK, Body: `
{
	"next_url": "/v2/shared_domains?page=2",
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
	Path:   "/v2/shared_domains",
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

var notFoundDomainsRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method:   "GET",
	Path:     "/v2/organizations/my-org-guid/private_domains",
	Response: testnet.TestResponse{Status: http.StatusNotFound},
})
var oldEndpointDomainsRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/domains",
	Response: testnet.TestResponse{Status: http.StatusOK, Body: `{
	"resources": [
		{
		  "metadata": {
			"guid": "domain-guid"
		  },
		  "entity": {
			"name": "example.com",
			"owning_organization_guid": "my-org-guid"
		  }
		}
	]
}`}})

var firstPageDomainsRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/organizations/my-org-guid/private_domains",
	Response: testnet.TestResponse{Status: http.StatusOK, Body: `
{
	"next_url": "/v2/organizations/my-org-guid/private_domains?page=2",
	"resources": [
		{
		  "metadata": {
			"guid": "domain1-guid"
		  },
		  "entity": {
			"name": "example.com",
			"owning_organization_guid": "my-org-guid"
		  }
		},
		{
		  "metadata": {
			"guid": "domain2-guid"
		  },
		  "entity": {
			"name": "some-example.com",
			"owning_organization_guid": "my-org-guid"
		  }
		}
	]
}`},
})

var secondPageDomainsRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/organizations/my-org-guid/private_domains?page=2",
	Response: testnet.TestResponse{Status: http.StatusOK, Body: `
{
	"resources": [
		{
		  "metadata": {
			"guid": "domain3-guid"
		  },
		  "entity": {
			"name": "example.com",
			"owning_organization_guid": "my-org-guid"
		  }
		}
	]
}`},
})

func deleteDomainReq(statusCode int) testnet.TestRequest {
	return testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "DELETE",
		Path:     "/v2/private_domains/my-domain-guid?recursive=true",
		Response: testnet.TestResponse{Status: statusCode},
	})
}

func deleteSharedDomainReq(statusCode int) testnet.TestRequest {
	return testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "DELETE",
		Path:     "/v2/shared_domains/my-domain-guid?recursive=true",
		Response: testnet.TestResponse{Status: statusCode},
	})
}

func createDomainRepo(t mr.TestingT, reqs []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo DomainRepository) {
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
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestListSharedDomains", func() {
			ts, handler, repo := createDomainRepo(mr.T(), []testnet.TestRequest{firstPageSharedDomainsRequest, secondPageSharedDomainsRequest})
			defer ts.Close()

			receivedDomains := []cf.Domain{}
			apiResponse := repo.ListSharedDomains(ListDomainsCallback(func(domains []cf.Domain) bool {
				receivedDomains = append(receivedDomains, domains...)
				return true
			}))

			assert.True(mr.T(), apiResponse.IsSuccessful())
			assert.Equal(mr.T(), len(receivedDomains), 2)
			assert.Equal(mr.T(), receivedDomains[0].Guid, "shared-domain1-guid")
			assert.Equal(mr.T(), receivedDomains[1].Guid, "shared-domain2-guid")
			assert.True(mr.T(), handler.AllRequestsCalled())
		})
		It("TestDomainListDomainsForOrgWithOldEndpoint", func() {

			ts, handler, repo := createDomainRepo(mr.T(), []testnet.TestRequest{notFoundDomainsRequest, oldEndpointDomainsRequest})
			defer ts.Close()

			receivedDomains := []cf.Domain{}
			apiResponse := repo.ListDomainsForOrg("my-org-guid", ListDomainsCallback(func(domains []cf.Domain) bool {
				receivedDomains = append(receivedDomains, domains...)
				return true
			}))

			assert.True(mr.T(), apiResponse.IsSuccessful())
			assert.Equal(mr.T(), len(receivedDomains), 1)
			assert.Equal(mr.T(), receivedDomains[0].Guid, "domain-guid")
			assert.True(mr.T(), handler.AllRequestsCalled())
		})
		It("TestDomainListDomainsForOrg", func() {

			ts, handler, repo := createDomainRepo(mr.T(), []testnet.TestRequest{firstPageDomainsRequest, secondPageDomainsRequest})
			defer ts.Close()

			receivedDomains := []cf.Domain{}
			apiResponse := repo.ListDomainsForOrg("my-org-guid", ListDomainsCallback(func(domains []cf.Domain) bool {
				receivedDomains = append(receivedDomains, domains...)
				return true
			}))

			assert.True(mr.T(), apiResponse.IsSuccessful())
			assert.Equal(mr.T(), len(receivedDomains), 3)
			assert.Equal(mr.T(), receivedDomains[0].Guid, "domain1-guid")
			assert.Equal(mr.T(), receivedDomains[1].Guid, "domain2-guid")
			assert.True(mr.T(), handler.AllRequestsCalled())
		})
		It("TestListDomainsForOrgWithNoDomains", func() {

			ts, handler, repo := createDomainRepo(mr.T(), []testnet.TestRequest{noDomainsRequest})
			defer ts.Close()

			wasCalled := false
			apiResponse := repo.ListDomainsForOrg("my-org-guid", ListDomainsCallback(func(domains []cf.Domain) bool {
				wasCalled = true
				return true
			}))

			assert.True(mr.T(), apiResponse.IsSuccessful())
			assert.False(mr.T(), wasCalled)
			assert.True(mr.T(), handler.AllRequestsCalled())
		})
		It("TestDomainListDomainsForOrgWithNoDomains", func() {

			emptyDomainsRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/organizations/my-org-guid/private_domains",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": [] }`},
			})

			ts, handler, repo := createDomainRepo(mr.T(), []testnet.TestRequest{emptyDomainsRequest})
			defer ts.Close()

			receivedDomains := []cf.Domain{}
			apiResponse := repo.ListDomainsForOrg("my-org-guid", ListDomainsCallback(func(domains []cf.Domain) bool {
				receivedDomains = append(receivedDomains, domains...)
				return true
			}))

			assert.True(mr.T(), apiResponse.IsSuccessful())
			assert.True(mr.T(), handler.AllRequestsCalled())
		})
		It("TestDomainFindByName", func() {

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

			ts, handler, repo := createDomainRepo(mr.T(), []testnet.TestRequest{req})
			defer ts.Close()

			domain, apiResponse := repo.FindByName("domain2.cf-app.com")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())

			assert.Equal(mr.T(), domain.Name, "domain2.cf-app.com")
			assert.Equal(mr.T(), domain.Guid, "domain2-guid")
		})
		It("TestDomainFindByNameInCurrentSpace", func() {

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

			ts, handler, repo := createDomainRepo(mr.T(), []testnet.TestRequest{req})
			defer ts.Close()

			domain, apiResponse := repo.FindByNameInCurrentSpace("domain2.cf-app.com")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())

			assert.Equal(mr.T(), domain.Name, "domain2.cf-app.com")
			assert.Equal(mr.T(), domain.Guid, "domain2-guid")
		})
		It("TestDomainFindByNameInCurrentSpaceWhenNotFound", func() {

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

			ts, handler, repo := createDomainRepo(mr.T(), []testnet.TestRequest{spaceDomainsReq, sharedDomainsReq})
			defer ts.Close()

			_, apiResponse := repo.FindByNameInCurrentSpace("domain2.cf-app.com")
			assert.True(mr.T(), handler.AllRequestsCalled())

			assert.False(mr.T(), apiResponse.IsError())
			assert.True(mr.T(), apiResponse.IsNotFound())
		})
		It("TestDomainFindByNameInCurrentSpaceWhenFoundAsSharedDomain", func() {

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

			ts, handler, repo := createDomainRepo(mr.T(), []testnet.TestRequest{spaceDomainsReq, sharedDomainsReq})
			defer ts.Close()

			domain, apiResponse := repo.FindByNameInCurrentSpace("domain2.cf-app.com")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())

			assert.Equal(mr.T(), domain.Name, "shared-domain.cf-app.com")
			assert.Equal(mr.T(), domain.Guid, "shared-domain-guid")
		})
		It("TestDomainFindByNameInCurrentSpaceWhenFoundInDomainsButNotShared", func() {

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

			ts, handler, repo := createDomainRepo(mr.T(), []testnet.TestRequest{spaceDomainsReq, sharedDomainsReq})
			defer ts.Close()

			_, apiResponse := repo.FindByNameInCurrentSpace("domain2.cf-app.com")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsError())
			assert.True(mr.T(), apiResponse.IsNotFound())
		})
		It("TestDomainFindByNameInOrg", func() {

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/organizations/my-org-guid/domains?inline-relations-depth=1&q=name%3Adomain2.cf-app.com",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": [
			{
			  "metadata": { "guid": "my-domain-guid" },
			  "entity": {
				"name": "my-example.com",
				"owning_organization_guid": "my-org-guid"
			  }
			}
		]}`},
			})

			ts, handler, repo := createDomainRepo(mr.T(), []testnet.TestRequest{req})
			defer ts.Close()

			domain, apiResponse := repo.FindByNameInOrg("domain2.cf-app.com", "my-org-guid")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())

			assert.Equal(mr.T(), domain.Name, "my-example.com")
			assert.Equal(mr.T(), domain.Guid, "my-domain-guid")
			assert.False(mr.T(), domain.Shared)
		})
		It("TestDomainFindByNameInOrgWhenNotFoundOnBothEndpoints", func() {

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

			ts, handler, repo := createDomainRepo(mr.T(), []testnet.TestRequest{orgDomainsReq, sharedDomainsReq})
			defer ts.Close()

			_, apiResponse := repo.FindByNameInOrg("domain2.cf-app.com", "my-org-guid")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsError())
			assert.True(mr.T(), apiResponse.IsNotFound())
		})
		It("TestDomainFindByNameInOrgWhenFoundAsSharedDomain", func() {

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
				"owning_organization_guid": null
			  }
			}
		]}`},
			})

			ts, handler, repo := createDomainRepo(mr.T(), []testnet.TestRequest{orgDomainsReq, sharedDomainsReq})
			defer ts.Close()

			domain, apiResponse := repo.FindByNameInOrg("domain2.cf-app.com", "my-org-guid")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())

			assert.Equal(mr.T(), domain.Name, "shared-example.com")
			assert.Equal(mr.T(), domain.Guid, "shared-domain-guid")
			assert.True(mr.T(), domain.Shared)
		})
		It("TestDomainFindByNameInOrgWhenFoundInDomainsButNotShared", func() {

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
				"owning_organization_guid": "some-other-org-guid"
			  }
			}
		]}`},
			})

			ts, handler, repo := createDomainRepo(mr.T(), []testnet.TestRequest{orgDomainsReq, sharedDomainsReq})
			defer ts.Close()

			_, apiResponse := repo.FindByNameInOrg("domain2.cf-app.com", "my-org-guid")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsError())
			assert.True(mr.T(), apiResponse.IsNotFound())
		})
		It("TestCreateDomainUsingOldEndpoint", func() {

			ts, handler, repo := createDomainRepo(mr.T(), []testnet.TestRequest{
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "POST",
					Path:     "/v2/private_domains",
					Matcher:  testnet.RequestBodyMatcher(`{"name":"example.com","owning_organization_guid":"org-guid"}`),
					Response: testnet.TestResponse{Status: http.StatusNotFound},
				}),
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:  "POST",
					Path:    "/v2/domains",
					Matcher: testnet.RequestBodyMatcher(`{"name":"example.com","owning_organization_guid":"org-guid", "wildcard": true}`),
					Response: testnet.TestResponse{Status: http.StatusCreated, Body: `{
				"metadata": { "guid": "abc-123" },
				"entity": { "name": "example.com" }
			}`},
				}),
			})
			defer ts.Close()

			createdDomain, apiResponse := repo.Create("example.com", "org-guid")

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
			assert.Equal(mr.T(), createdDomain.Guid, "abc-123")
		})
		It("TestCreateDomainUsingNewEndpoint", func() {

			ts, handler, repo := createDomainRepo(mr.T(), []testnet.TestRequest{
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:  "POST",
					Path:    "/v2/private_domains",
					Matcher: testnet.RequestBodyMatcher(`{"name":"example.com","owning_organization_guid":"org-guid"}`),
					Response: testnet.TestResponse{Status: http.StatusCreated, Body: `{
				"metadata": { "guid": "abc-123" },
				"entity": { "name": "example.com" }
			}`},
				}),
			})
			defer ts.Close()

			createdDomain, apiResponse := repo.Create("example.com", "org-guid")

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
			assert.Equal(mr.T(), createdDomain.Guid, "abc-123")
		})
		It("TestCreateSharedDomainsWithNewEndpoint", func() {

			ts, handler, repo := createDomainRepo(mr.T(), []testnet.TestRequest{
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:  "POST",
					Path:    "/v2/shared_domains",
					Matcher: testnet.RequestBodyMatcher(`{"name":"example.com"}`),
					Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
			{
				"metadata": { "guid": "abc-123" },
				"entity": { "name": "example.com" }
			}`},
				}),
			})
			defer ts.Close()

			apiResponse := repo.CreateSharedDomain("example.com")

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())
		})
		It("TestCreateSharedDomainsWithOldEndpoint", func() {

			ts, handler, repo := createDomainRepo(mr.T(), []testnet.TestRequest{
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "POST",
					Path:     "/v2/shared_domains",
					Matcher:  testnet.RequestBodyMatcher(`{"name":"example.com"}`),
					Response: testnet.TestResponse{Status: http.StatusNotFound},
				}),
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:  "POST",
					Path:    "/v2/domains",
					Matcher: testnet.RequestBodyMatcher(`{"name":"example.com", "wildcard": true}`),
					Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
			{
				"metadata": { "guid": "abc-123" },
				"entity": { "name": "example.com" }
			}`},
				}),
			})
			defer ts.Close()

			apiResponse := repo.CreateSharedDomain("example.com")

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())
		})
		It("TestDeleteDomainWithNewEndpoint", func() {

			ts, handler, repo := createDomainRepo(mr.T(), []testnet.TestRequest{
				deleteDomainReq(http.StatusOK),
			})
			defer ts.Close()

			apiResponse := repo.Delete("my-domain-guid")

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
		})
		It("TestDeleteDomainWithOldEndpoint", func() {

			ts, handler, repo := createDomainRepo(mr.T(), []testnet.TestRequest{
				deleteDomainReq(http.StatusNotFound),
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "DELETE",
					Path:     "/v2/domains/my-domain-guid?recursive=true",
					Response: testnet.TestResponse{Status: http.StatusOK},
				}),
			})
			defer ts.Close()

			apiResponse := repo.Delete("my-domain-guid")

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
		})
		It("TestDeleteSharedDomainWithNewEndpoint", func() {

			ts, handler, repo := createDomainRepo(mr.T(), []testnet.TestRequest{
				deleteSharedDomainReq(http.StatusOK),
			})
			defer ts.Close()

			apiResponse := repo.DeleteSharedDomain("my-domain-guid")

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
		})
		It("TestDeleteSharedDomainWithOldEndpoint", func() {

			ts, handler, repo := createDomainRepo(mr.T(), []testnet.TestRequest{
				deleteSharedDomainReq(http.StatusNotFound),
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "DELETE",
					Path:     "/v2/domains/my-domain-guid?recursive=true",
					Response: testnet.TestResponse{Status: http.StatusOK},
				}),
			})
			defer ts.Close()

			apiResponse := repo.DeleteSharedDomain("my-domain-guid")

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
		})
		It("TestDeleteDomainFailure", func() {

			req := deleteDomainReq(http.StatusBadRequest)

			ts, handler, repo := createDomainRepo(mr.T(), []testnet.TestRequest{req})
			defer ts.Close()

			apiResponse := repo.Delete("my-domain-guid")

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsNotSuccessful())
		})
	})
}
