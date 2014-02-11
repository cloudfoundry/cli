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

func init() {
	Describe("CloudControllerQuotaRepository", func() {
		Describe("FindByName", func() {
			It("finds a quota given a particular name", func() {
				req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "GET",
					Path:   "/v2/quota_definitions?q=name%3Amy-quota",
					Response: testnet.TestResponse{
						Status: http.StatusOK,
						Body: `{"resources": [
							{
							  "metadata": { "guid": "my-quota-guid" },
							  "entity": { "name": "my-remote-quota", "memory_limit": 1024 }
							}
						]}`},
				})

				ts, handler, repo := createQuotaRepo(req)
				defer ts.Close()

				quota, apiResponse := repo.FindByName("my-quota")
				Expect(handler.AllRequestsCalled()).To(BeTrue())
				Expect(apiResponse.IsNotSuccessful()).To(BeFalse())
				expectedQuota := models.QuotaFields{}
				expectedQuota.Guid = "my-quota-guid"
				expectedQuota.Name = "my-remote-quota"
				expectedQuota.MemoryLimit = 1024
				Expect(quota).To(Equal(expectedQuota))
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

				apiResponse := repo.Update("my-org-guid", "my-quota-guid")
				Expect(handler.AllRequestsCalled()).To(BeTrue())
				Expect(apiResponse.IsNotSuccessful()).To(BeFalse())
			})

		})
	})
}

func createQuotaRepo(req testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo QuotaRepository) {
	ts, handler = testnet.NewTLSServer(GinkgoT(), []testnet.TestRequest{req})

	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerQuotaRepository(configRepo, gateway)
	return
}
