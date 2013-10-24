package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testnet "testhelpers/net"
	"testing"
)

func TestFindQuotaByName(t *testing.T) {
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

	ts, handler, repo := createQuotaRepo(t, req)
	defer ts.Close()

	quota, apiResponse := repo.FindByName("my-quota")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, quota, cf.Quota{Guid: "my-quota-guid", Name: "my-remote-quota", MemoryLimit: 1024})
}

func TestUpdateQuota(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     "/v2/organizations/my-org-guid",
		Matcher:  testnet.RequestBodyMatcher(`{"quota_definition_guid":"my-quota-guid"}`),
		Response: testnet.TestResponse{Status: http.StatusCreated},
	})

	ts, handler, repo := createQuotaRepo(t, req)
	defer ts.Close()

	quota := cf.Quota{Guid: "my-quota-guid"}
	org := cf.Organization{Guid: "my-org-guid"}
	apiResponse := repo.Update(org, quota)
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func createQuotaRepo(t *testing.T, req testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo QuotaRepository) {
	ts, handler = testnet.NewTLSServer(t, []testnet.TestRequest{req})

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerQuotaRepository(config, gateway)
	return
}
