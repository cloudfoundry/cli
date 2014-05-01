/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/github.com/cloudfoundry/cli/cf/commands/application/delete_app_test.go
   src/github.com/cloudfoundry/cli/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package api_test

import (
	. "github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("Organization Repository", func() {
	It("TestOrganizationsListOrgs", func() {
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
		apiErr := repo.ListOrgs(func(o models.Organization) bool {
			orgs = append(orgs, o)
			return true
		})

		Expect(len(orgs)).To(Equal(3))
		Expect(orgs[0].Guid).To(Equal("org1-guid"))
		Expect(orgs[1].Guid).To(Equal("org2-guid"))
		Expect(orgs[2].Guid).To(Equal("org3-guid"))
		Expect(apiErr).NotTo(HaveOccurred())
		Expect(handler).To(testnet.HaveAllRequestsCalled())
	})

	It("TestOrganizationsListOrgsWithNoOrgs", func() {
		emptyOrgsRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "GET",
			Path:     "/v2/organizations",
			Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
		})

		testserver, handler, repo := createOrganizationRepo(emptyOrgsRequest)
		defer testserver.Close()

		wasCalled := false
		apiErr := repo.ListOrgs(func(o models.Organization) bool {
			wasCalled = true
			return false
		})

		Expect(wasCalled).To(BeFalse())
		Expect(apiErr).NotTo(HaveOccurred())
		Expect(handler).To(testnet.HaveAllRequestsCalled())
	})

	It("TestOrganizationsFindByName", func() {
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
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())

		Expect(org.Name).To(Equal(existingOrg.Name))
		Expect(org.Guid).To(Equal(existingOrg.Guid))
		Expect(org.QuotaDefinition.Name).To(Equal("not-your-average-quota"))
		Expect(org.QuotaDefinition.MemoryLimit).To(Equal(uint64(128)))
		Expect(len(org.Spaces)).To(Equal(1))
		Expect(org.Spaces[0].Name).To(Equal("Space1"))
		Expect(org.Spaces[0].Guid).To(Equal("space1-guid"))
		Expect(len(org.Domains)).To(Equal(1))
		Expect(org.Domains[0].Name).To(Equal("cfapps.io"))
		Expect(org.Domains[0].Guid).To(Equal("domain1-guid"))
	})

	It("TestOrganizationsFindByNameWhenDoesNotExist", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "GET",
			Path:     "/v2/organizations?q=name%3Aorg1&inline-relations-depth=1",
			Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
		})

		testserver, handler, repo := createOrganizationRepo(req)
		defer testserver.Close()

		_, apiErr := repo.FindByName("org1")
		Expect(handler).To(testnet.HaveAllRequestsCalled())

		Expect(apiErr.(*errors.ModelNotFoundError)).NotTo(BeNil())
	})

	It("returns an api error when one occurs", func() {
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
		Expect(handler).To(testnet.HaveAllRequestsCalled())
	})

	It("TestCreateOrganization", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "POST",
			Path:     "/v2/organizations",
			Matcher:  testnet.RequestBodyMatcher(`{"name":"my-org"}`),
			Response: testnet.TestResponse{Status: http.StatusCreated},
		})

		testserver, handler, repo := createOrganizationRepo(req)
		defer testserver.Close()

		apiErr := repo.Create("my-org")
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})

	It("TestRenameOrganization", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "PUT",
			Path:     "/v2/organizations/my-org-guid",
			Matcher:  testnet.RequestBodyMatcher(`{"name":"my-new-org"}`),
			Response: testnet.TestResponse{Status: http.StatusCreated},
		})

		testserver, handler, repo := createOrganizationRepo(req)
		defer testserver.Close()

		apiErr := repo.Rename("my-org-guid", "my-new-org")
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})

	It("TestDeleteOrganization", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "DELETE",
			Path:     "/v2/organizations/my-org-guid?recursive=true",
			Response: testnet.TestResponse{Status: http.StatusOK},
		})

		testserver, handler, repo := createOrganizationRepo(req)
		defer testserver.Close()

		apiErr := repo.Delete("my-org-guid")
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})
})

func createOrganizationRepo(reqs ...testnet.TestRequest) (testserver *httptest.Server, handler *testnet.TestHandler, repo OrganizationRepository) {
	testserver, handler = testnet.NewServer(reqs)

	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(testserver.URL)
	gateway := net.NewCloudControllerGateway(configRepo)
	repo = NewCloudControllerOrganizationRepository(configRepo, gateway)
	return
}
