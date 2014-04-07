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

var _ = Describe("CloudControllerQuotaRepository", func() {
	Describe("FindByName", func() {
		It("finds a Quota given a particular name", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/quota_definitions?q=name%3Amy-quota",
				Response: testnet.TestResponse{
					Status: http.StatusOK,
					Body: `{
					"resources": [
						{
						  "metadata": { "guid": "my-quota-guid" },
						  "entity": { "name": "my-remote-quota", "memory_limit": 1024 }
						}
					]}`,
				},
			})

			ts, handler, repo := createQuotaRepo(req)
			defer ts.Close()

			quota, apiErr := repo.FindByName("my-quota")
			Expect(handler).To(testnet.HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
			expectedQuota := models.QuotaFields{}
			expectedQuota.Guid = "my-quota-guid"
			expectedQuota.Name = "my-remote-quota"
			expectedQuota.MemoryLimit = 1024
			Expect(quota).To(Equal(expectedQuota))
		})

	})

	Describe("Find Quotas multipage", func() {

		It("FindByName Quota definition", func() {
			ts, handler, repo := createQuotaRepo2([]testnet.TestRequest{firstQuotaResponse, secondQuotaResponse})
			defer ts.Close()
			quota, apiErr := repo.FindByName("my-remote-quota")
			Expect(apiErr).NotTo(HaveOccurred())
			Expect(handler).To(testnet.HaveAllRequestsCalled())
			expectedQuota := models.QuotaFields{}
			expectedQuota.Guid = "my-quota-guid"
			expectedQuota.Name = "my-remote-quota"
			expectedQuota.MemoryLimit = 1024
			Expect(quota).To(Equal(expectedQuota))
		})

		It("FindAll Quota definitions", func() {
			ts, handler, repo := createQuotaRepo2([]testnet.TestRequest{firstQuotaResponse, secondQuotaResponse})
			defer ts.Close()

			quotas, apiErr := repo.FindAll()
			Expect(apiErr).NotTo(HaveOccurred())
			Expect(handler).To(testnet.HaveAllRequestsCalled())
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

			ts, handler, repo := createQuotaRepo(req)
			defer ts.Close()

			apiErr := repo.Update("my-org-guid", "my-quota-guid")
			Expect(handler).To(testnet.HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})

	})
})

var firstQuotaResponse = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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

var secondQuotaResponse = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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

func createQuotaRepo(req testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo QuotaRepository) {
	return createQuotaRepo2([]testnet.TestRequest{req})
}

func createQuotaRepo2(reqs []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo QuotaRepository) {
	ts, handler = testnet.NewServer(reqs)

	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway(configRepo)
	repo = NewCloudControllerQuotaRepository(configRepo, gateway)
	return
}
