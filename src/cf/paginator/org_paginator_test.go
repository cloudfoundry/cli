package paginator_test

import (
	"cf/configuration"
	"cf/net"
	"cf/paginator"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testnet "testhelpers/net"
	"testing"
)

func TestOrganizationPaginator(t *testing.T) {
	firstReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/organizations",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{
		"total_pages": 2,
		"next_url": "/v2/organizations?page=2",
		"resources": [
			{
			  "metadata": { "guid": "org1-guid" },
			  "entity": { "name": "Org1" }
			}
		]}`},
	})

	secondReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/organizations?page=2",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{
		"total_pages": 2,
		"next_url": "",
		"resources": [
			{
			  "metadata": { "guid": "org2-guid" },
			  "entity": { "name": "Org2" }
			}
		]}`},
	})

	ts, handler, p := createOrgPaginator(t, firstReq, secondReq)
	defer ts.Close()

	assert.True(t, p.HasNext())

	firstChunk, apiResponse := p.Next()

	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, len(firstChunk), 1)

	firstOutput := firstChunk[0]
	assert.Contains(t, firstOutput, "Org1")
	assert.True(t, p.HasNext())

	secondChunk, apiResponse := p.Next()

	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, len(secondChunk), 1)

	secondOutput := secondChunk[0]
	assert.Contains(t, secondOutput, "Org2")

	assert.False(t, p.HasNext())

	assert.True(t, handler.AllRequestsCalled())
}

func createOrgPaginator(t *testing.T, reqs ...testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, p *paginator.OrganizationPaginator) {
	ts, handler = testnet.NewTLSServer(t, reqs)

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway()
	p = paginator.NewOrganizationPaginator(config, gateway)
	return
}
