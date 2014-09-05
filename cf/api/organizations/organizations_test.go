package organizations_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"

	. "github.com/cloudfoundry/cli/cf/api/organizations"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Organization Repository", func() {
	Describe("listing organizations", func() {
		It("lists the orgs from the the /v2/orgs endpoint", func() {
			firstPageOrgsRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/organizations",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{
	"next_url": "/v2/organizations?page=2",
	"resources": [
		{
		  "metadata": { "guid": "org1-guid" },
		  "entity": { "name": "Org1" }
		},
		{
		  "metadata": { "guid": "org2-guid" },
		  "entity": { "name": "Org2" }
		}
	]}`},
			})

			secondPageOrgsRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/organizations?page=2",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": [
		{
		  "metadata": { "guid": "org3-guid" },
		  "entity": { "name": "Org3" }
		}
	]}`},
			})

			testserver, handler, repo := createOrganizationRepo(firstPageOrgsRequest, secondPageOrgsRequest)
			defer testserver.Close()

			orgs := []models.Organization{}
			orgs, apiErr := repo.ListOrgs()

			Expect(len(orgs)).To(Equal(3))
			Expect(orgs[0].Guid).To(Equal("org1-guid"))
			Expect(orgs[1].Guid).To(Equal("org2-guid"))
			Expect(orgs[2].Guid).To(Equal("org3-guid"))
			Expect(apiErr).NotTo(HaveOccurred())
			Expect(handler).To(HaveAllRequestsCalled())
		})

		It("does not call the provided function when there are no orgs found", func() {
			emptyOrgsRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/organizations",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
			})

			testserver, handler, repo := createOrganizationRepo(emptyOrgsRequest)
			defer testserver.Close()

			_, apiErr := repo.ListOrgs()

			Expect(apiErr).NotTo(HaveOccurred())
			Expect(handler).To(HaveAllRequestsCalled())
		})
	})

	Describe("finding organizations by name", func() {
		It("returns the org with that name", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
		}]
	  }
	}]}`},
			})

			testserver, handler, repo := createOrganizationRepo(req)
			defer testserver.Close()
			existingOrg := models.Organization{}
			existingOrg.Guid = "org1-guid"
			existingOrg.Name = "Org1"

			org, apiErr := repo.FindByName("Org1")
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())

			Expect(org.Name).To(Equal(existingOrg.Name))
			Expect(org.Guid).To(Equal(existingOrg.Guid))
			Expect(org.QuotaDefinition.Name).To(Equal("not-your-average-quota"))
			Expect(org.QuotaDefinition.MemoryLimit).To(Equal(int64(128)))
			Expect(len(org.Spaces)).To(Equal(1))
			Expect(org.Spaces[0].Name).To(Equal("Space1"))
			Expect(org.Spaces[0].Guid).To(Equal("space1-guid"))
			Expect(len(org.Domains)).To(Equal(1))
			Expect(org.Domains[0].Name).To(Equal("cfapps.io"))
			Expect(org.Domains[0].Guid).To(Equal("domain1-guid"))
		})

		It("returns a ModelNotFoundError when the org cannot be found", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
			requestHandler := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
						Guid: "my-quota-guid",
					},
				}}

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
})

func createOrganizationRepo(reqs ...testnet.TestRequest) (testserver *httptest.Server, handler *testnet.TestHandler, repo OrganizationRepository) {
	testserver, handler = testnet.NewServer(reqs)

	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(testserver.URL)
	gateway := net.NewCloudControllerGateway(configRepo, time.Now)
	repo = NewCloudControllerOrganizationRepository(configRepo, gateway)
	return
}
