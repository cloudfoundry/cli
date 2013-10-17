package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	"testing"
)

func TestFindQuotaByName(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"GET",
		"/v2/quota_definitions?q=name%3Amy-quota",
		nil,
		testapi.TestResponse{Status: http.StatusOK, Body: `{
  "resources": [
    {
      "metadata": {
        "guid": "my-quota-guid"
      },
      "entity": {
        "name": "my-remote-quota"
      }
    }
  ]
}`},
	)

	ts, repo := createQuotaRepo(endpoint)
	defer ts.Close()

	quota, apiResponse := repo.FindByName("my-quota")
	assert.True(t, status.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, quota, cf.Quota{Guid: "my-quota-guid", Name: "my-remote-quota"})
}

func TestUpdateQuota(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"PUT",
		"/v2/organizations/my-org-guid",
		testapi.RequestBodyMatcher(`{"quota_definition_guid":"my-quota-guid"}`),
		testapi.TestResponse{Status: http.StatusCreated},
	)

	ts, repo := createQuotaRepo(endpoint)
	defer ts.Close()

	quota := cf.Quota{Guid: "my-quota-guid"}
	org := cf.Organization{Guid: "my-org-guid"}
	apiResponse := repo.Update(org, quota)
	assert.True(t, status.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func createQuotaRepo(endpoint http.HandlerFunc) (ts *httptest.Server, repo QuotaRepository) {
	ts = httptest.NewTLSServer(endpoint)

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerQuotaRepository(config, gateway)
	return
}
