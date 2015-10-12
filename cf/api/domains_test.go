package api_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/api/strategy"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/api"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DomainRepository", func() {
	var (
		ts      *httptest.Server
		handler *testnet.TestHandler
		repo    DomainRepository
		config  core_config.ReadWriter
	)

	BeforeEach(func() {
		config = testconfig.NewRepositoryWithDefaults()
	})

	JustBeforeEach(func() {
		gateway := net.NewCloudControllerGateway((config), time.Now, &testterm.FakeUI{})
		strategy := strategy.NewEndpointStrategy(config.ApiVersion())
		repo = NewCloudControllerDomainRepository(config, gateway, strategy)
	})

	AfterEach(func() {
		ts.Close()
	})

	var setupTestServer = func(reqs ...testnet.TestRequest) {
		ts, handler = testnet.NewServer(reqs)
		config.SetApiEndpoint(ts.URL)
	}

	Describe("listing domains", func() {
		BeforeEach(func() {
			config.SetApiVersion("2.2.0")
			setupTestServer(firstPagePrivateDomainsRequest, secondPagePrivateDomainsRequest, firstPageSharedDomainsRequest, secondPageSharedDomainsRequest)
		})

		It("uses the organization-scoped domains endpoints", func() {
			receivedDomains := []models.DomainFields{}
			apiErr := repo.ListDomainsForOrg("my-org-guid", func(d models.DomainFields) bool {
				receivedDomains = append(receivedDomains, d)
				return true
			})

			Expect(apiErr).NotTo(HaveOccurred())
			Expect(len(receivedDomains)).To(Equal(6))
			Expect(receivedDomains[0].Guid).To(Equal("domain1-guid"))
			Expect(receivedDomains[1].Guid).To(Equal("domain2-guid"))
			Expect(receivedDomains[2].Guid).To(Equal("domain3-guid"))
			Expect(receivedDomains[2].Shared).To(BeFalse())
			Expect(receivedDomains[3].Guid).To(Equal("shared-domain1-guid"))
			Expect(receivedDomains[4].Guid).To(Equal("shared-domain2-guid"))
			Expect(receivedDomains[5].Guid).To(Equal("shared-domain3-guid"))
			Expect(handler).To(HaveAllRequestsCalled())
		})
	})

	Describe("getting default domain", func() {
		BeforeEach(func() {
			config.SetApiVersion("2.2.0")
			setupTestServer(firstPagePrivateDomainsRequest, secondPagePrivateDomainsRequest, firstPageSharedDomainsRequest, secondPageSharedDomainsRequest)
		})

		It("should always return back the first shared domain", func() {
			domain, apiErr := repo.FirstOrDefault("my-org-guid", nil)

			Expect(apiErr).NotTo(HaveOccurred())
			Expect(domain.Guid).To(Equal("shared-domain1-guid"))
		})
	})

	It("finds a shared domain by name", func() {
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

		domain, apiErr := repo.FindSharedByName("domain2.cf-app.com")
		Expect(handler).To(HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())

		Expect(domain.Name).To(Equal("domain2.cf-app.com"))
		Expect(domain.Guid).To(Equal("domain2-guid"))
		Expect(domain.Shared).To(BeTrue())
	})

	It("finds a private domain by name", func() {
		setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "GET",
			Path:   "/v2/domains?inline-relations-depth=1&q=name%3Adomain2.cf-app.com",
			Response: testnet.TestResponse{Status: http.StatusOK, Body: `
				{
					"resources": [
						{
						  "metadata": { "guid": "domain2-guid" },
						  "entity": { "name": "domain2.cf-app.com", "owning_organization_guid": "some-guid" }
						}
					]
				}`},
		}))

		domain, apiErr := repo.FindPrivateByName("domain2.cf-app.com")
		Expect(handler).To(HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())

		Expect(domain.Name).To(Equal("domain2.cf-app.com"))
		Expect(domain.Guid).To(Equal("domain2-guid"))
		Expect(domain.Shared).To(BeFalse())
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
			Expect(handler).To(HaveAllRequestsCalled())
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
			Expect(handler).To(HaveAllRequestsCalled())
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
			Expect(handler).To(HaveAllRequestsCalled())
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
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr.(*errors.ModelNotFoundError)).NotTo(BeNil())
		})
	})

	Describe("creating domains", func() {
		Context("when the private domains endpoint is not available", func() {
			BeforeEach(func() {
				setupTestServer(
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

				Expect(handler).To(HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
				Expect(createdDomain.Guid).To(Equal("abc-123"))
			})
		})

		Context("when the private domains endpoint is available", func() {
			BeforeEach(func() {
				config.SetApiVersion("2.2.1")
			})

			It("uses that endpoint", func() {
				setupTestServer(
					testapi.NewCloudControllerTestRequest(testnet.TestRequest{
						Method:  "POST",
						Path:    "/v2/private_domains",
						Matcher: testnet.RequestBodyMatcher(`{"name":"example.com","owning_organization_guid":"org-guid", "wildcard": true}`),
						Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
						{
							"metadata": { "guid": "abc-123" },
							"entity": { "name": "example.com" }
						}`},
					}))

				createdDomain, apiErr := repo.Create("example.com", "org-guid")

				Expect(handler).To(HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
				Expect(createdDomain.Guid).To(Equal("abc-123"))
			})
		})
	})

	Describe("creating shared domains", func() {
		Context("targeting a newer cloud controller", func() {
			BeforeEach(func() {
				config.SetApiVersion("2.2.0")
			})

			Context("when router_group_guid is the empty string", func() {
				It("uses the shared domains endpoint", func() {
					setupTestServer(
						testapi.NewCloudControllerTestRequest(testnet.TestRequest{
							Method:  "POST",
							Path:    "/v2/shared_domains",
							Matcher: testnet.RequestBodyMatcher(`{"name":"example.com", "wildcard": true}`),
							Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
						{
							"metadata": { "guid": "abc-123" },
							"entity": { "name": "example.com" }
						}`}}),
					)

					apiErr := repo.CreateSharedDomain("example.com", "")

					Expect(handler).To(HaveAllRequestsCalled())
					Expect(apiErr).NotTo(HaveOccurred())
				})
			})

			Context("when router_group_guid is not the empty string", func() {
				It("creates a shared domain with a router_group_guid", func() {
					setupTestServer(
						testapi.NewCloudControllerTestRequest(testnet.TestRequest{
							Method:  "POST",
							Path:    "/v2/shared_domains",
							Matcher: testnet.RequestBodyMatcher(`{"name":"example.com", "router_group_guid": "tcp-group", "wildcard": true}`),
							Response: testnet.TestResponse{Status: http.StatusCreated, Body: `
						{
							"metadata": { "guid": "abc-123" },
							"entity": { "name": "example.com", "router_group_guid":"tcp-group" }
						}`}}),
					)

					apiErr := repo.CreateSharedDomain("example.com", "tcp-group")

					Expect(handler).To(HaveAllRequestsCalled())
					Expect(apiErr).NotTo(HaveOccurred())
				})
			})
		})

		Context("when targeting an older cloud controller", func() {
			It("uses the general domains endpoint", func() {
				setupTestServer(
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

				apiErr := repo.CreateSharedDomain("example.com", "")

				Expect(handler).To(HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
			})
		})
	})

	Describe("deleting domains", func() {
		Context("when the private domains endpoint is available", func() {
			BeforeEach(func() {
				config.SetApiVersion("2.2.0")
				setupTestServer(deleteDomainReq(http.StatusOK))
			})

			It("uses the private domains endpoint", func() {
				apiErr := repo.Delete("my-domain-guid")

				Expect(handler).To(HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
			})
		})

		Context("when the private domains endpoint is NOT available", func() {
			BeforeEach(func() {
				setupTestServer(
					testapi.NewCloudControllerTestRequest(testnet.TestRequest{
						Method:   "DELETE",
						Path:     "/v2/domains/my-domain-guid?recursive=true",
						Response: testnet.TestResponse{Status: http.StatusOK},
					}))
			})

			It("uses the general domains endpoint", func() {
				apiErr := repo.Delete("my-domain-guid")

				Expect(handler).To(HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
			})
		})
	})

	Describe("deleting shared domains", func() {
		Context("when the shared domains endpoint is available", func() {
			BeforeEach(func() {
				config.SetApiVersion("2.2.0")
				setupTestServer(deleteSharedDomainReq(http.StatusOK))
			})

			It("uses the shared domains endpoint", func() {
				apiErr := repo.DeleteSharedDomain("my-domain-guid")

				Expect(handler).To(HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
			})

			It("returns an error when the delete fails", func() {
				setupTestServer(deleteSharedDomainReq(http.StatusBadRequest))

				apiErr := repo.DeleteSharedDomain("my-domain-guid")

				Expect(handler).To(HaveAllRequestsCalled())
				Expect(apiErr).NotTo(BeNil())
			})
		})

		Context("when the shared domains endpoint is not available", func() {
			It("uses the old domains endpoint", func() {
				setupTestServer(
					testapi.NewCloudControllerTestRequest(testnet.TestRequest{
						Method:   "DELETE",
						Path:     "/v2/domains/my-domain-guid?recursive=true",
						Response: testnet.TestResponse{Status: http.StatusOK},
					}))

				apiErr := repo.DeleteSharedDomain("my-domain-guid")

				Expect(handler).To(HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
			})
		})
	})

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
	"next_url": "/v2/organizations/my-org-guid/domains?page=2",
	"resources": [
		{
		  "metadata": {
			"guid": "domain1-guid",
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
	Path:   "/v2/organizations/my-org-guid/domains?page=2",
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
			"name": "sharedexample.com"
		  }
		},
		{
		  "metadata": {
			"guid": "shared-domain2-guid"
		  },
		  "entity": {
			"name": "some-other-shared-example.com"
		  }
		}
	]
}`},
})

var secondPageSharedDomainsRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/shared_domains?page=2",
	Response: testnet.TestResponse{Status: http.StatusOK, Body: `
{
	"resources": [
		{
		  "metadata": {
			"guid": "shared-domain3-guid"
		  },
		  "entity": {
			"name": "yet-another-shared-example.com"
		  }
		}
	]
}`},
})

var firstPagePrivateDomainsRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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

var secondPagePrivateDomainsRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
			"owning_organization_guid": null,
			"shared_organizations_url": "/v2/private_domains/domain3-guid/shared_organizations"
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
