package api_test

import (
	. "cf/api"
	"cf/configuration"
	"cf/models"
	"cf/net"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testnet "testhelpers/net"
)

func createQuotaRepo(t mr.TestingT, req testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo QuotaRepository) {
	ts, handler = testnet.NewTLSServer(t, []testnet.TestRequest{req})

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerQuotaRepository(config, gateway)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestFindQuotaByName", func() {
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

			ts, handler, repo := createQuotaRepo(mr.T(), req)
			defer ts.Close()

			quota, apiResponse := repo.FindByName("my-quota")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
			expectedQuota := models.QuotaFields{}
			expectedQuota.Guid = "my-quota-guid"
			expectedQuota.Name = "my-remote-quota"
			expectedQuota.MemoryLimit = 1024
			assert.Equal(mr.T(), quota, expectedQuota)
		})
		It("TestUpdateQuota", func() {

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/organizations/my-org-guid",
				Matcher:  testnet.RequestBodyMatcher(`{"quota_definition_guid":"my-quota-guid"}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			ts, handler, repo := createQuotaRepo(mr.T(), req)
			defer ts.Close()

			apiResponse := repo.Update("my-org-guid", "my-quota-guid")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
		})
	})
}
