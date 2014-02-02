package api_test

import (
	. "cf/api"
	"cf/configuration"
	"cf/net"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testnet "testhelpers/net"
)

var allStacksResponse = testnet.TestResponse{Status: http.StatusOK, Body: `
{
  "resources": [
    {
      "metadata": {
        "guid": "50688ae5-9bfc-4bf6-a4bf-caadb21a32c6",
        "url": "/v2/stacks/50688ae5-9bfc-4bf6-a4bf-caadb21a32c6",
        "created_at": "2013-08-31 01:32:40 +0000",
        "updated_at": "2013-08-31 01:32:40 +0000"
      },
      "entity": {
        "name": "lucid64",
        "description": "Ubuntu 10.04"
      }
    },
    {
      "metadata": {
        "guid": "e8cda251-7ce8-44b9-becb-ba5f5913d8ba",
        "url": "/v2/stacks/e8cda251-7ce8-44b9-becb-ba5f5913d8ba",
        "created_at": "2013-08-31 01:32:40 +0000",
        "updated_at": "2013-08-31 01:32:40 +0000"
      },
      "entity": {
        "name": "lucid64custom",
        "description": "Fake Ubuntu 10.04"
      }
    }
  ]
}`}

func createStackRepo(t mr.TestingT, req testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo StackRepository) {
	ts, handler = testnet.NewTLSServer(t, []testnet.TestRequest{req})

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerStackRepository(config, gateway)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestStacksFindByName", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/stacks?q=name%3Alinux",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: ` { "resources": [
			{
			  "metadata": { "guid": "custom-linux-guid" },
			  "entity": { "name": "custom-linux" }
			}
  		]}`}})
			ts, handler, repo := createStackRepo(mr.T(), req)
			defer ts.Close()

			stack, apiResponse := repo.FindByName("linux")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())
			assert.Equal(mr.T(), stack.Name, "custom-linux")
			assert.Equal(mr.T(), stack.Guid, "custom-linux-guid")
		})
		It("TestStacksFindByNameNotFound", func() {

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/stacks?q=name%3Alinux",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: ` { "resources": []}`},
			})
			ts, handler, repo := createStackRepo(mr.T(), req)
			defer ts.Close()

			_, apiResponse := repo.FindByName("linux")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsNotSuccessful())
		})
		It("TestStacksFindAll", func() {

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/stacks",
				Response: allStacksResponse,
			})

			ts, handler, repo := createStackRepo(mr.T(), req)
			defer ts.Close()

			stacks, apiResponse := repo.FindAll()
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
			assert.Equal(mr.T(), len(stacks), 2)
			assert.Equal(mr.T(), stacks[0].Name, "lucid64")
			assert.Equal(mr.T(), stacks[0].Guid, "50688ae5-9bfc-4bf6-a4bf-caadb21a32c6")
		})
	})
}
