package organizations_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testnet "code.cloudfoundry.org/cli/util/testhelpers/net"

	. "code.cloudfoundry.org/cli/cf/api/organizations"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Organization Repository", func() {
	Describe("ListOrgs", func() {
		var (
			ccServer *ghttp.Server
			repo     CloudControllerOrganizationRepository
		)

		Context("when there are orgs", func() {
			BeforeEach(func() {
				ccServer = ghttp.NewServer()
				configRepo := testconfig.NewRepositoryWithDefaults()
				configRepo.SetAPIEndpoint(ccServer.URL())
				gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
				repo = NewCloudControllerOrganizationRepository(configRepo, gateway)
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/organizations", "order-by=name"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
						"total_results": 3,
						"total_pages": 2,
						"prev_url": null,
						"next_url": "/v2/organizations?order-by=name&page=2",
						"resources": [
							{
								"metadata": { "guid": "org3-guid" },
								"entity": { "name": "Alpha" }
							},
							{
								"metadata": { "guid": "org2-guid" },
								"entity": { "name": "Beta" }
							}
						]
					}`),
					),
				)

				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/organizations", "order-by=name&page=2"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
						"total_results": 3,
						"total_pages": 2,
						"prev_url": null,
						"next_url": null,
						"resources": [
							{
								"metadata": { "guid": "org1-guid" },
								"entity": { "name": "Gamma" }
							}
						]
					}`),
					),
				)
			})

			AfterEach(func() {
				ccServer.Close()
			})

			Context("when given a non-zero positive limit", func() {
				It("should return no more than the limit number of organizations", func() {
					orgs, err := repo.ListOrgs(2)
					Expect(err).NotTo(HaveOccurred())
					Expect(len(orgs)).To(Equal(2))
				})

				It("should not make more requests than necessary to retrieve the requested number of orgs", func() {
					_, err := repo.ListOrgs(2)
					Expect(err).NotTo(HaveOccurred())
					Expect(ccServer.ReceivedRequests()).Should(HaveLen(1))
				})
			})

			Context("when given a zero limit", func() {
				It("should return all organizations", func() {
					orgs, err := repo.ListOrgs(0)
					Expect(err).NotTo(HaveOccurred())
					Expect(len(orgs)).To(Equal(3))
				})

				It("lists the orgs from the the /v2/orgs endpoint in alphabetical order", func() {
					orgs, apiErr := repo.ListOrgs(0)

					Expect(len(orgs)).To(Equal(3))
					Expect(orgs[0].GUID).To(Equal("org3-guid"))
					Expect(orgs[1].GUID).To(Equal("org2-guid"))
					Expect(orgs[2].GUID).To(Equal("org1-guid"))

					Expect(orgs[0].Name).To(Equal("Alpha"))
					Expect(orgs[1].Name).To(Equal("Beta"))
					Expect(orgs[2].Name).To(Equal("Gamma"))
					Expect(apiErr).NotTo(HaveOccurred())
				})
			})
		})

		Context("when there are no orgs", func() {
			BeforeEach(func() {
				ccServer = ghttp.NewServer()
				configRepo := testconfig.NewRepositoryWithDefaults()
				configRepo.SetAPIEndpoint(ccServer.URL())
				gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
				repo = NewCloudControllerOrganizationRepository(configRepo, gateway)
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/organizations"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusOK, `{"resources": []}`),
					),
				)
			})

			AfterEach(func() {
				ccServer.Close()
			})

			It("does not call the provided function", func() {
				_, apiErr := repo.ListOrgs(0)

				Expect(apiErr).NotTo(HaveOccurred())
			})
		})
	})

	Describe(".GetManyOrgsByGUID", func() {
		It("requests each org", func() {
			firstOrgRequest := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/organizations/org1-guid",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{
		  "metadata": { "guid": "org1-guid" },
		  "entity": { "name": "Org1" }
		}`},
			})
			secondOrgRequest := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/organizations/org2-guid",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{
			"metadata": { "guid": "org2-guid" },
		  "entity": { "name": "Org2" }
	  }`},
			})
			testserver, handler, repo := createOrganizationRepo(firstOrgRequest, secondOrgRequest)
			defer testserver.Close()

			orgGUIDs := []string{"org1-guid", "org2-guid"}
			orgs, err := repo.GetManyOrgsByGUID(orgGUIDs)
			Expect(err).NotTo(HaveOccurred())

			Expect(handler).To(HaveAllRequestsCalled())
			Expect(len(orgs)).To(Equal(2))
			Expect(orgs[0].GUID).To(Equal("org1-guid"))
			Expect(orgs[1].GUID).To(Equal("org2-guid"))
		})
	})

	Describe("finding organizations by name", func() {
		It("returns the org with that name", func() {
			req := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/organizations?q=name%3Aorg1&inline-relations-depth=1",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": [{
	  "metadata": { "guid": "org1-guid" },
	  "entity": {
		"name": "Org1",
		"quota_definition": {
		  "entity": {
			"name": "not-your-average-quota",
			"memory_limit": 128
		  }
		},
		"spaces": [{
		  "metadata": { "guid": "space1-guid" },
		  "entity": { "name": "Space1" }
		}],
		"domains": [{
		  "metadata": { "guid": "domain1-guid" },
		  "entity": { "name": "cfapps.io" }
		}],
		"space_quota_definitions":[{
			"metadata": {"guid": "space-quota1-guid"},
			"entity": {"name": "space-quota1"}
		}]
	  }
	}]}`},
			})

			testserver, handler, repo := createOrganizationRepo(req)
			defer testserver.Close()
			existingOrg := models.Organization{}
			existingOrg.GUID = "org1-guid"
			existingOrg.Name = "Org1"

			org, apiErr := repo.FindByName("Org1")
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())

			Expect(org.Name).To(Equal(existingOrg.Name))
			Expect(org.GUID).To(Equal(existingOrg.GUID))
			Expect(org.QuotaDefinition.Name).To(Equal("not-your-average-quota"))
			Expect(org.QuotaDefinition.MemoryLimit).To(Equal(int64(128)))
			Expect(len(org.Spaces)).To(Equal(1))
			Expect(org.Spaces[0].Name).To(Equal("Space1"))
			Expect(org.Spaces[0].GUID).To(Equal("space1-guid"))
			Expect(len(org.Domains)).To(Equal(1))
			Expect(org.Domains[0].Name).To(Equal("cfapps.io"))
			Expect(org.Domains[0].GUID).To(Equal("domain1-guid"))
			Expect(len(org.SpaceQuotas)).To(Equal(1))
			Expect(org.SpaceQuotas[0].Name).To(Equal("space-quota1"))
			Expect(org.SpaceQuotas[0].GUID).To(Equal("space-quota1-guid"))
		})

		It("returns a ModelNotFoundError when the org cannot be found", func() {
			req := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/organizations?q=name%3Aorg1&inline-relations-depth=1",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
			})

			testserver, handler, repo := createOrganizationRepo(req)
			defer testserver.Close()

			_, apiErr := repo.FindByName("org1")
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr.(*errors.ModelNotFoundError)).NotTo(BeNil())
		})

		It("returns an api error when the response is not successful", func() {
			requestHandler := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/organizations?q=name%3Aorg1&inline-relations-depth=1",
				Response: testnet.TestResponse{Status: http.StatusBadGateway, Body: `{"resources": []}`},
			})

			testserver, handler, repo := createOrganizationRepo(requestHandler)
			defer testserver.Close()

			_, apiErr := repo.FindByName("org1")
			_, ok := apiErr.(*errors.ModelNotFoundError)
			Expect(ok).To(BeFalse())
			Expect(handler).To(HaveAllRequestsCalled())
		})
	})

	Describe(".Create", func() {
		It("creates the org and sends only the org name if the quota flag is not provided", func() {
			org := models.Organization{
				OrganizationFields: models.OrganizationFields{
					Name: "my-org",
				}}

			req := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "POST",
				Path:     "/v2/organizations",
				Matcher:  testnet.RequestBodyMatcher(`{"name":"my-org"}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			testserver, handler, repo := createOrganizationRepo(req)
			defer testserver.Close()

			apiErr := repo.Create(org)
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})

		It("creates the org with the provided quota", func() {
			org := models.Organization{
				OrganizationFields: models.OrganizationFields{
					Name: "my-org",
					QuotaDefinition: models.QuotaFields{
						GUID: "my-quota-guid",
					},
				}}

			req := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "POST",
				Path:     "/v2/organizations",
				Matcher:  testnet.RequestBodyMatcher(`{"name":"my-org", "quota_definition_guid":"my-quota-guid"}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			testserver, handler, repo := createOrganizationRepo(req)
			defer testserver.Close()

			apiErr := repo.Create(org)
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})
	})

	Describe("renaming orgs", func() {
		It("renames the org with the given guid", func() {
			req := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/organizations/my-org-guid",
				Matcher:  testnet.RequestBodyMatcher(`{"name":"my-new-org"}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			testserver, handler, repo := createOrganizationRepo(req)
			defer testserver.Close()

			apiErr := repo.Rename("my-org-guid", "my-new-org")
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})
	})

	Describe("deleting orgs", func() {
		It("deletes the org with the given guid", func() {
			req := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "DELETE",
				Path:     "/v2/organizations/my-org-guid?recursive=true",
				Response: testnet.TestResponse{Status: http.StatusOK},
			})

			testserver, handler, repo := createOrganizationRepo(req)
			defer testserver.Close()

			apiErr := repo.Delete("my-org-guid")
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})
	})

	Describe("SharePrivateDomain", func() {
		It("shares the private domain with the given org", func() {
			req := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/organizations/my-org-guid/private_domains/domain-guid",
				Response: testnet.TestResponse{Status: http.StatusOK},
			})

			testserver, handler, repo := createOrganizationRepo(req)
			defer testserver.Close()

			apiErr := repo.SharePrivateDomain("my-org-guid", "domain-guid")
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})
	})

	Describe("UnsharePrivateDomain", func() {
		It("unshares the private domain with the given org", func() {
			req := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "DELETE",
				Path:     "/v2/organizations/my-org-guid/private_domains/domain-guid",
				Response: testnet.TestResponse{Status: http.StatusOK},
			})

			testserver, handler, repo := createOrganizationRepo(req)
			defer testserver.Close()

			apiErr := repo.UnsharePrivateDomain("my-org-guid", "domain-guid")
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})
	})
})

func createOrganizationRepo(reqs ...testnet.TestRequest) (testserver *httptest.Server, handler *testnet.TestHandler, repo OrganizationRepository) {
	testserver, handler = testnet.NewServer(reqs)

	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetAPIEndpoint(testserver.URL)
	gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
	repo = NewCloudControllerOrganizationRepository(configRepo, gateway)
	return
}
