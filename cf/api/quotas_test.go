package api_test

import (
	"net/http"
	"net/http/httptest"
	"time"
	. "github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CloudControllerQuotaRepository", func() {
	var (
		testServer  *httptest.Server
		testHandler *testnet.TestHandler
		configRepo  configuration.ReadWriter
		repo        CloudControllerQuotaRepository
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		gateway := net.NewCloudControllerGateway((configRepo), time.Now)
		repo = NewCloudControllerQuotaRepository(configRepo, gateway)
	})

	AfterEach(func() {
		testServer.Close()
	})

	setupTestServer := func(reqs ...testnet.TestRequest) {
		testServer, testHandler = testnet.NewServer(reqs)
		configRepo.SetApiEndpoint(testServer.URL)
	}

	Describe("FindByName", func() {
		BeforeEach(func() {
			setupTestServer(firstQuotaRequest, secondQuotaRequest)
		})

		It("Finds Quota definitions by name", func() {
			quota, err := repo.FindByName("my-remote-quota")

			Expect(err).NotTo(HaveOccurred())
			Expect(testHandler).To(testnet.HaveAllRequestsCalled())
			Expect(quota).To(Equal(models.QuotaFields{
				Guid:                    "my-quota-guid",
				Name:                    "my-remote-quota",
				MemoryLimit:             1024,
				RoutesLimit:             123,
				ServicesLimit:           321,
				NonBasicServicesAllowed: true,
			}))
		})
	})

	Describe("FindAll", func() {
		BeforeEach(func() {
			setupTestServer(firstQuotaRequest, secondQuotaRequest)
		})

		It("finds all Quota definitions", func() {
			quotas, err := repo.FindAll()

			Expect(err).NotTo(HaveOccurred())
			Expect(testHandler).To(testnet.HaveAllRequestsCalled())
			Expect(len(quotas)).To(Equal(3))
			Expect(quotas[0].Guid).To(Equal("my-quota-guid"))
			Expect(quotas[0].Name).To(Equal("my-remote-quota"))
			Expect(quotas[0].MemoryLimit).To(Equal(uint64(1024)))
			Expect(quotas[0].RoutesLimit).To(Equal(123))
			Expect(quotas[0].ServicesLimit).To(Equal(321))

			Expect(quotas[1].Guid).To(Equal("my-quota-guid2"))
			Expect(quotas[2].Guid).To(Equal("my-quota-guid3"))
		})
	})

	Describe("AssignQuotaToOrg", func() {
		It("sets the quota for an organization", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/organizations/my-org-guid",
				Matcher:  testnet.RequestBodyMatcher(`{"quota_definition_guid":"my-quota-guid"}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			setupTestServer(req)

			err := repo.AssignQuotaToOrg("my-org-guid", "my-quota-guid")
			Expect(testHandler).To(testnet.HaveAllRequestsCalled())
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Create", func() {
		It("creates a new quota with the given name", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "POST",
				Path:   "/v2/quota_definitions",
				Matcher: testnet.RequestBodyMatcher(`{
					"name": "not-so-strict",
					"non_basic_services_allowed": false,
					"total_services": 1,
					"total_routes": 12,
					"memory_limit": 123
				}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})
			setupTestServer(req)

			quota := models.QuotaFields{
				Name:          "not-so-strict",
				ServicesLimit: 1,
				RoutesLimit:   12,
				MemoryLimit:   123,
			}
			err := repo.Create(quota)
			Expect(err).NotTo(HaveOccurred())
			Expect(testHandler).To(testnet.HaveAllRequestsCalled())
		})
	})

	Describe("Update", func() {
		It("updates an existing quota", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "PUT",
				Path:   "/v2/quota_definitions/my-quota-guid",
				Matcher: testnet.RequestBodyMatcher(`{
					"guid": "my-quota-guid",
					"non_basic_services_allowed": false,
					"name": "amazing-quota",
					"total_services": 1,
					"total_routes": 12,
					"memory_limit": 123
				}`),
			}))

			quota := models.QuotaFields{
				Guid:          "my-quota-guid",
				Name:          "amazing-quota",
				ServicesLimit: 1,
				RoutesLimit:   12,
				MemoryLimit:   123,
			}

			err := repo.Update(quota)
			Expect(err).NotTo(HaveOccurred())
			Expect(testHandler).To(testnet.HaveAllRequestsCalled())
		})
	})

	Describe("Delete", func() {
		It("deletes the quota with the given name", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "DELETE",
				Path:     "/v2/quota_definitions/my-quota-guid",
				Response: testnet.TestResponse{Status: http.StatusNoContent},
			})
			setupTestServer(req)

			err := repo.Delete("my-quota-guid")
			Expect(err).NotTo(HaveOccurred())
			Expect(testHandler).To(testnet.HaveAllRequestsCalled())
		})
	})
})

var firstQuotaRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/quota_definitions",
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `{
		"next_url": "/v2/quota_definitions?page=2",
		"resources": [
			{
			  "metadata": { "guid": "my-quota-guid" },
			  "entity": {
			  	"name": "my-remote-quota",
			  	"memory_limit": 1024,
			  	"total_routes": 123,
			  	"total_services": 321,
			  	"non_basic_services_allowed": true
			  }
			}
		]}`,
	},
})

var secondQuotaRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/quota_definitions?page=2",
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `{
		"resources": [
			{
			  "metadata": { "guid": "my-quota-guid2" },
			  "entity": { "name": "my-remote-quota2", "memory_limit": 1024 }
			},
			{
			  "metadata": { "guid": "my-quota-guid3" },
			  "entity": { "name": "my-remote-quota3", "memory_limit": 1024 }
			}
		]}`,
	},
})
