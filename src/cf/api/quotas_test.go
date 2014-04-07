package api_test

import (
	. "cf/api"
	"cf/configuration"
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

var _ = Describe("CloudControllerQuotaRepository", func() {
	var (
		testServer  *httptest.Server
		testHandler *testnet.TestHandler
		configRepo  configuration.ReadWriter
		repo        CloudControllerQuotaRepository
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		gateway := net.NewCloudControllerGateway(configRepo)
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
				Guid:        "my-quota-guid",
				Name:        "my-remote-quota",
				MemoryLimit: 1024,
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
			Expect(quotas[1].Guid).To(Equal("my-quota-guid2"))
			Expect(quotas[2].Guid).To(Equal("my-quota-guid3"))
		})
	})

	Describe("Update", func() {
		It("sets the quota for an organization", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/organizations/my-org-guid",
				Matcher:  testnet.RequestBodyMatcher(`{"quota_definition_guid":"my-quota-guid"}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			setupTestServer(req)

			err := repo.Update("my-org-guid", "my-quota-guid")
			Expect(testHandler).To(testnet.HaveAllRequestsCalled())
			Expect(err).NotTo(HaveOccurred())
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
			  "entity": { "name": "my-remote-quota", "memory_limit": 1024 }
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
