package api_test

import (
	. "cf/api"
	"cf/configuration"
	"cf/errors"
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
	var (
		ts      *httptest.Server
		handler *testnet.TestHandler
		repo    DomainRepository
		config  configuration.ReadWriter
	)

	BeforeEach(func() {
		config = testconfig.NewRepositoryWithDefaults()
		gateway := net.NewCloudControllerGateway(config)
		repo = NewCloudControllerDomainRepository(config, gateway)
	})

	AfterEach(func() {
		ts.Close()
	})

	var setupTestServer = func(reqs ...testnet.TestRequest) {
		ts, handler = testnet.NewServer(reqs)
		config.SetApiEndpoint(ts.URL)
	}

	Describe("listing domains", func() {
		It("lists shared domains", func() {
			setupTestServer(firstPageSharedDomainsRequest, secondPageSharedDomainsRequest)

			receivedDomains := []models.DomainFields{}
			apiErr := repo.ListSharedDomains(func(d models.DomainFields) bool {
				receivedDomains = append(receivedDomains, d)
				return true
			})

			Expect(apiErr).NotTo(HaveOccurred())
			Expect(len(receivedDomains)).To(Equal(2))
			Expect(receivedDomains[0].Guid).To(Equal("shared-domain1-guid"))
			Expect(receivedDomains[1].Guid).To(Equal("shared-domain2-guid"))
			Expect(handler).To(testnet.HaveAllRequestsCalled())
		})

		Describe("listing the domains for an organization", func() {
			Context("when the organization-scoped domains endpoint is available", func() {
				BeforeEach(func() {
					setupTestServer(firstPageDomainsRequest, secondPageDomainsRequest)
				})

				It("uses that endpoint", func() {
					receivedDomains := []models.DomainFields{}
					apiErr := repo.ListDomainsForOrg("my-org-guid", func(d models.DomainFields) bool {
						receivedDomains = append(receivedDomains, d)
						return true
					})

					Expect(apiErr).NotTo(HaveOccurred())
					Expect(len(receivedDomains)).To(Equal(3))
					Expect(receivedDomains[0].Guid).To(Equal("domain1-guid"))
					Expect(receivedDomains[1].Guid).To(Equal("domain2-guid"))
					Expect(handler).To(testnet.HaveAllRequestsCalled())
				})
			})

			Context("when the organization-scoped endpoint returns a 404", func() {
				It("uses the global domains endpoint", func() {
					setupTestServer(notFoundDomainsRequest, oldEndpointDomainsRequest)

					receivedDomains := []models.DomainFields{}
					apiErr := repo.ListDomainsForOrg("my-org-guid", func(d models.DomainFields) bool {
						receivedDomains = append(receivedDomains, d)
						return true
					})

					Expect(apiErr).NotTo(HaveOccurred())
					Expect(len(receivedDomains)).To(Equal(1))
					Expect(receivedDomains[0].Guid).To(Equal("domain-guid"))
					Expect(handler).To(testnet.HaveAllRequestsCalled())
				})
			})
		})
	})

	It("TestDomainFindByName", func() {
		setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "GET",
			Path:   "/v2/domains?inline-relations-depth=1&q=name%3Adomain2.cf-app.com",
			Response: testnet.TestResponse{Status: http.StatusOK, Body: `
				{
					"resources": [
						{
						  "metadata": { "guid": "domain2-guid" },
						  "entity": { "name": "domain2.cf-app.com" }
						}
					]
				}`},
		}))

		domain, apiErr := repo.FindByName("domain2.cf-app.com")
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())

		Expect(domain.Name).To(Equal("domain2.cf-app.com"))
		Expect(domain.Guid).To(Equal("domain2-guid"))
	})

	Describe("finding a domain by name in an org", func() {
		It("looks in the org's domains first", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
			}))

			domain, apiErr := repo.FindByNameInOrg("domain2.cf-app.com", "my-org-guid")
			Expect(handler).To(testnet.HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())

			Expect(domain.Name).To(Equal("my-example.com"))
			Expect(domain.Guid).To(Equal("my-domain-guid"))
			Expect(domain.Shared).To(BeFalse())
		})

		It("looks for shared domains if no there are no org-specific domains", func() {
			setupTestServer(
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     "/v2/organizations/my-org-guid/domains?inline-relations-depth=1&q=name%3Adomain2.cf-app.com",
					Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
				}),

				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
				}))

			domain, apiErr := repo.FindByNameInOrg("domain2.cf-app.com", "my-org-guid")
			Expect(handler).To(testnet.HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())

			Expect(domain.Name).To(Equal("shared-example.com"))
			Expect(domain.Guid).To(Equal("shared-domain-guid"))
			Expect(domain.Shared).To(BeTrue())
		})

		It("returns not found when neither endpoint returns a domain", func() {
			setupTestServer(
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     "/v2/organizations/my-org-guid/domains?inline-relations-depth=1&q=name%3Adomain2.cf-app.com",
					Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
				}),

				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     "/v2/domains?inline-relations-depth=1&q=name%3Adomain2.cf-app.com",
					Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
				}))

			_, apiErr := repo.FindByNameInOrg("domain2.cf-app.com", "my-org-guid")
			Expect(handler).To(testnet.HaveAllRequestsCalled())
			Expect(apiErr.(*errors.ModelNotFoundError)).NotTo(BeNil())
		})

		It("returns not found when the global endpoint returns a non-shared domain", func() {
			setupTestServer(
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     "/v2/organizations/my-org-guid/domains?inline-relations-depth=1&q=name%3Adomain2.cf-app.com",
					Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
				}),

				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
					}`}}))

			_, apiErr := repo.FindByNameInOrg("domain2.cf-app.com", "my-org-guid")
			Expect(handler).To(testnet.HaveAllRequestsCalled())
			Expect(apiErr.(*errors.ModelNotFoundError)).NotTo(BeNil())
		})
	})

	Describe("creating domains", func() {
		Context("when the private domains endpoint is not available", func() {
			BeforeEach(func() {
				setupTestServer(
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
						Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
						{
							"metadata": { "guid": "abc-123" },
							"entity": { "name": "example.com" }
						}`},
					}),
				)
			})

			It("uses the general domains endpoint", func() {
				createdDomain, apiErr := repo.Create("example.com", "org-guid")

				Expect(handler).To(testnet.HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
				Expect(createdDomain.Guid).To(Equal("abc-123"))
			})
		})

		Context("when the private domains endpoint is available", func() {
			It("uses that endpoint", func() {
				setupTestServer(
					testapi.NewCloudControllerTestRequest(testnet.TestRequest{
						Method:  "POST",
						Path:    "/v2/private_domains",
						Matcher: testnet.RequestBodyMatcher(`{"name":"example.com","owning_organization_guid":"org-guid"}`),
						Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
						{
							"metadata": { "guid": "abc-123" },
							"entity": { "name": "example.com" }
						}`},
					}))

				createdDomain, apiErr := repo.Create("example.com", "org-guid")

				Expect(handler).To(testnet.HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
				Expect(createdDomain.Guid).To(Equal("abc-123"))
			})
		})
	})

	Describe("creating shared domains", func() {
		Context("when the shared domains endpoint is available", func() {
			It("uses that endpoint", func() {
				setupTestServer(
					testapi.NewCloudControllerTestRequest(testnet.TestRequest{
						Method:  "POST",
						Path:    "/v2/shared_domains",
						Matcher: testnet.RequestBodyMatcher(`{"name":"example.com"}`),
						Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
						{
							"metadata": { "guid": "abc-123" },
							"entity": { "name": "example.com" }
						}`}}),
				)

				apiErr := repo.CreateSharedDomain("example.com")

				Expect(handler).To(testnet.HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
			})
		})

		Context("when the shared domains endpoint is not available", func() {
			It("uses the general domains endpoint", func() {
				setupTestServer(
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
				)

				apiErr := repo.CreateSharedDomain("example.com")

				Expect(handler).To(testnet.HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
			})
		})
	})

	Describe("deleting domains", func() {
		Context("when the private domains endpoint is available", func() {
			It("uses the private domains endpoint", func() {
				setupTestServer(deleteDomainReq(http.StatusOK))

				apiErr := repo.Delete("my-domain-guid")

				Expect(handler).To(testnet.HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
			})
		})

		Context("when the private domains endpoint is NOT available", func() {
			It("uses the general domains endpoint", func() {
				setupTestServer(
					deleteDomainReq(http.StatusNotFound),
					testapi.NewCloudControllerTestRequest(testnet.TestRequest{
						Method:   "DELETE",
						Path:     "/v2/domains/my-domain-guid?recursive=true",
						Response: testnet.TestResponse{Status: http.StatusOK},
					}),
				)

				apiErr := repo.Delete("my-domain-guid")

				Expect(handler).To(testnet.HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
			})
		})
	})

	Describe("deleting shared domains", func() {
		Context("when the shared domains endpoint is available", func() {
			It("uses the shared domains endpoint", func() {
				setupTestServer(deleteSharedDomainReq(http.StatusOK))

				apiErr := repo.DeleteSharedDomain("my-domain-guid")

				Expect(handler).To(testnet.HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
			})
		})

		Context("when the shared domains endpoint is not available", func() {
			It("uses the old domains endpoint", func() {
				setupTestServer(
					deleteSharedDomainReq(http.StatusNotFound),
					testapi.NewCloudControllerTestRequest(testnet.TestRequest{
						Method:   "DELETE",
						Path:     "/v2/domains/my-domain-guid?recursive=true",
						Response: testnet.TestResponse{Status: http.StatusOK},
					}))

				apiErr := repo.DeleteSharedDomain("my-domain-guid")

				Expect(handler).To(testnet.HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
			})

			It("returns an error when the delete fails", func() {
				setupTestServer(deleteDomainReq(http.StatusBadRequest))

				apiErr := repo.Delete("my-domain-guid")

				Expect(handler).To(testnet.HaveAllRequestsCalled())
				Expect(apiErr).NotTo(BeNil())
			})
		})
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
