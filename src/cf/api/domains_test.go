package api_test

import (
	. "cf/api"
	"cf/models"
	"cf/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
)

var _ = Describe("DomainRepository", func() {
	It("TestListSharedDomains", func() {
		ts, handler, repo := createDomainRepo([]testnet.TestRequest{firstPageSharedDomainsRequest, secondPageSharedDomainsRequest})
		defer ts.Close()

		receivedDomains := []models.DomainFields{}
		apiResponse := repo.ListSharedDomains(func(d models.DomainFields) bool {
			receivedDomains = append(receivedDomains, d)
			return true
		})

		Expect(apiResponse).NotTo(HaveOccurred())
		Expect(len(receivedDomains)).To(Equal(2))
		Expect(receivedDomains[0].Guid).To(Equal("shared-domain1-guid"))
		Expect(receivedDomains[1].Guid).To(Equal("shared-domain2-guid"))
		Expect(handler).To(testnet.HaveAllRequestsCalled())
	})

	It("TestDomainListDomainsForOrgWithOldEndpoint", func() {
		ts, handler, repo := createDomainRepo([]testnet.TestRequest{notFoundDomainsRequest, oldEndpointDomainsRequest})
		defer ts.Close()

		receivedDomains := []models.DomainFields{}
		apiResponse := repo.ListDomainsForOrg("my-org-guid", func(d models.DomainFields) bool {
			receivedDomains = append(receivedDomains, d)
			return true
		})

		Expect(apiResponse).NotTo(HaveOccurred())
		Expect(len(receivedDomains)).To(Equal(1))
		Expect(receivedDomains[0].Guid).To(Equal("domain-guid"))
		Expect(handler).To(testnet.HaveAllRequestsCalled())
	})

	It("TestDomainListDomainsForOrg", func() {
		ts, handler, repo := createDomainRepo([]testnet.TestRequest{firstPageDomainsRequest, secondPageDomainsRequest})
		defer ts.Close()

		receivedDomains := []models.DomainFields{}
		apiResponse := repo.ListDomainsForOrg("my-org-guid", func(d models.DomainFields) bool {
			receivedDomains = append(receivedDomains, d)
			return true
		})

		Expect(apiResponse).NotTo(HaveOccurred())
		Expect(len(receivedDomains)).To(Equal(3))
		Expect(receivedDomains[0].Guid).To(Equal("domain1-guid"))
		Expect(receivedDomains[1].Guid).To(Equal("domain2-guid"))
		Expect(handler).To(testnet.HaveAllRequestsCalled())
	})

	It("TestListDomainsForOrgWithNoDomains", func() {
		ts, handler, repo := createDomainRepo([]testnet.TestRequest{noDomainsRequest})
		defer ts.Close()

		wasCalled := false
		apiResponse := repo.ListDomainsForOrg("my-org-guid", func(d models.DomainFields) bool {
			wasCalled = true
			return true
		})

		Expect(apiResponse).NotTo(HaveOccurred())
		Expect(wasCalled).To(BeFalse())
		Expect(handler).To(testnet.HaveAllRequestsCalled())
	})

	It("TestDomainListDomainsForOrgWithNoDomains", func() {
		emptyDomainsRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "GET",
			Path:     "/v2/organizations/my-org-guid/private_domains",
			Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": [] }`},
		})

		ts, handler, repo := createDomainRepo([]testnet.TestRequest{emptyDomainsRequest})
		defer ts.Close()

		receivedDomains := []models.DomainFields{}
		apiResponse := repo.ListDomainsForOrg("my-org-guid", func(d models.DomainFields) bool {
			receivedDomains = append(receivedDomains, d)
			return true
		})

		Expect(apiResponse).NotTo(HaveOccurred())
		Expect(handler).To(testnet.HaveAllRequestsCalled())
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

		ts, handler, repo := createDomainRepo([]testnet.TestRequest{req})
		defer ts.Close()

		domain, apiResponse := repo.FindByName("domain2.cf-app.com")
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiResponse).NotTo(HaveOccurred())

		Expect(domain.Name).To(Equal("domain2.cf-app.com"))
		Expect(domain.Guid).To(Equal("domain2-guid"))
	})

	Describe("finding a domain by name in an org", func() {
		It("looks in the org's domains first", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/organizations/my-org-guid/domains?inline-relations-depth=1&q=name%3Adomain2.cf-app.com",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `
				{
					"resources": [
						{
						  "metadata": { "guid": "my-domain-guid" },
						  "entity": {
							"name": "my-example.com",
							"owning_organization_guid": "my-org-guid"
						  }
						}
					]
				}`},
			})

			ts, handler, repo := createDomainRepo([]testnet.TestRequest{req})
			defer ts.Close()

			domain, apiResponse := repo.FindByNameInOrg("domain2.cf-app.com", "my-org-guid")
			Expect(handler).To(testnet.HaveAllRequestsCalled())
			Expect(apiResponse).NotTo(HaveOccurred())

			Expect(domain.Name).To(Equal("my-example.com"))
			Expect(domain.Guid).To(Equal("my-domain-guid"))
			Expect(domain.Shared).To(BeFalse())
		})

		It("looks for shared domains if no there are no org-specific domains", func() {
			orgDomainsReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/organizations/my-org-guid/domains?inline-relations-depth=1&q=name%3Adomain2.cf-app.com",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
			})

			sharedDomainsReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/domains?inline-relations-depth=1&q=name%3Adomain2.cf-app.com",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `
				{
					"resources": [
						{
						  "metadata": { "guid": "shared-domain-guid" },
						  "entity": {
							"name": "shared-example.com",
							"owning_organization_guid": null
						  }
						}
					]
				}`},
			})

			ts, handler, repo := createDomainRepo([]testnet.TestRequest{orgDomainsReq, sharedDomainsReq})
			defer ts.Close()

			domain, apiResponse := repo.FindByNameInOrg("domain2.cf-app.com", "my-org-guid")
			Expect(handler).To(testnet.HaveAllRequestsCalled())
			Expect(apiResponse).NotTo(HaveOccurred())

			Expect(domain.Name).To(Equal("shared-example.com"))
			Expect(domain.Guid).To(Equal("shared-domain-guid"))
			Expect(domain.Shared).To(BeTrue())
		})

		It("returns not found when neither endpoint returns a domain", func() {
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

			ts, handler, repo := createDomainRepo([]testnet.TestRequest{orgDomainsReq, sharedDomainsReq})
			defer ts.Close()

			_, apiResponse := repo.FindByNameInOrg("domain2.cf-app.com", "my-org-guid")
			Expect(handler).To(testnet.HaveAllRequestsCalled())

			Expect(apiResponse.IsNotFound()).To(BeTrue())
		})

		It("returns not found when the global endpoint returns a non-shared domain", func() {
			orgDomainsReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/organizations/my-org-guid/domains?inline-relations-depth=1&q=name%3Adomain2.cf-app.com",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
			})

			sharedDomainsReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/domains?inline-relations-depth=1&q=name%3Adomain2.cf-app.com",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `
				{
					"resources": [
						{
						  "metadata": { "guid": "shared-domain-guid" },
						  "entity": {
							"name": "shared-example.com",
							"owning_organization_guid": "some-other-org-guid"
						  }
						}
					]
				}`},
			})

			ts, handler, repo := createDomainRepo([]testnet.TestRequest{orgDomainsReq, sharedDomainsReq})
			defer ts.Close()

			_, apiResponse := repo.FindByNameInOrg("domain2.cf-app.com", "my-org-guid")
			Expect(handler).To(testnet.HaveAllRequestsCalled())

			Expect(apiResponse.IsNotFound()).To(BeTrue())
		})
	})

	It("TestCreateDomainUsingOldEndpoint", func() {
		ts, handler, repo := createDomainRepo([]testnet.TestRequest{
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

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiResponse).NotTo(HaveOccurred())
		Expect(createdDomain.Guid).To(Equal("abc-123"))
	})

	It("TestCreateDomainUsingNewEndpoint", func() {
		ts, handler, repo := createDomainRepo([]testnet.TestRequest{
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

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiResponse).NotTo(HaveOccurred())
		Expect(createdDomain.Guid).To(Equal("abc-123"))
	})

	It("TestCreateSharedDomainsWithNewEndpoint", func() {
		ts, handler, repo := createDomainRepo([]testnet.TestRequest{
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

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiResponse).NotTo(HaveOccurred())
	})

	It("TestCreateSharedDomainsWithOldEndpoint", func() {
		ts, handler, repo := createDomainRepo([]testnet.TestRequest{
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

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiResponse).NotTo(HaveOccurred())
	})

	It("TestDeleteDomainWithNewEndpoint", func() {
		ts, handler, repo := createDomainRepo([]testnet.TestRequest{
			deleteDomainReq(http.StatusOK),
		})
		defer ts.Close()

		apiResponse := repo.Delete("my-domain-guid")

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiResponse).NotTo(HaveOccurred())
	})

	It("TestDeleteDomainWithOldEndpoint", func() {
		ts, handler, repo := createDomainRepo([]testnet.TestRequest{
			deleteDomainReq(http.StatusNotFound),
			testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "DELETE",
				Path:     "/v2/domains/my-domain-guid?recursive=true",
				Response: testnet.TestResponse{Status: http.StatusOK},
			}),
		})
		defer ts.Close()

		apiResponse := repo.Delete("my-domain-guid")

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiResponse).NotTo(HaveOccurred())
	})

	It("TestDeleteSharedDomainWithNewEndpoint", func() {
		ts, handler, repo := createDomainRepo([]testnet.TestRequest{
			deleteSharedDomainReq(http.StatusOK),
		})
		defer ts.Close()

		apiResponse := repo.DeleteSharedDomain("my-domain-guid")

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiResponse).NotTo(HaveOccurred())
	})

	It("TestDeleteSharedDomainWithOldEndpoint", func() {
		ts, handler, repo := createDomainRepo([]testnet.TestRequest{
			deleteSharedDomainReq(http.StatusNotFound),
			testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "DELETE",
				Path:     "/v2/domains/my-domain-guid?recursive=true",
				Response: testnet.TestResponse{Status: http.StatusOK},
			}),
		})
		defer ts.Close()

		apiResponse := repo.DeleteSharedDomain("my-domain-guid")

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiResponse).NotTo(HaveOccurred())
	})

	It("TestDeleteDomainFailure", func() {
		req := deleteDomainReq(http.StatusBadRequest)

		ts, handler, repo := createDomainRepo([]testnet.TestRequest{req})
		defer ts.Close()

		apiResponse := repo.Delete("my-domain-guid")

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiResponse).NotTo(BeNil())
	})
})

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

func createDomainRepo(reqs []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo DomainRepository) {
	ts, handler = testnet.NewServer(reqs)
	config := testconfig.NewRepositoryWithDefaults()
	config.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerDomainRepository(config, gateway)
	return
}
