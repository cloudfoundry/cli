package api_test

import (
	"cf"
	. "cf/api"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testnet "testhelpers/net"

	. "github.com/onsi/ginkgo"
	mr "github.com/tjarratt/mr_t"
	"time"
)

var appStatsRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/apps/my-cool-app-guid/stats",
	Response: testnet.TestResponse{Status: http.StatusOK, Body: `
{
  "1":{
    "stats": {
        "disk_quota": 10000,
        "mem_quota": 1024,
        "usage": {
            "cpu": 0.3,
            "disk": 10000,
            "mem": 1024
        }
    }
  },
  "0":{
    "stats": {
        "disk_quota": 1073741824,
        "mem_quota": 67108864,
        "usage": {
            "cpu": 3.659571249238058e-05,
            "disk": 56037376,
            "mem": 19218432
        }
    }
  }
}`}})

var appInstancesRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/apps/my-cool-app-guid/instances",
	Response: testnet.TestResponse{Status: http.StatusOK, Body: `
{
  "1": {
    "state": "STARTING",
    "since": 1379522342.6783738
  },
  "0": {
    "state": "RUNNING",
    "since": 1379522342.6783738
  }
}`}})

func createAppInstancesRepo(t mr.TestingT, requests []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo AppInstancesRepository) {
	ts, handler = testnet.NewTLSServer(t, requests)
	space := cf.SpaceFields{}
	space.Guid = "my-space-guid"
	config := &configuration.Configuration{
		SpaceFields: space,
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}

	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerAppInstancesRepository(config, gateway)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestAppInstancesGetInstances", func() {
			ts, handler, repo := createAppInstancesRepo(mr.T(), []testnet.TestRequest{
				appInstancesRequest,
				appStatsRequest,
			})
			defer ts.Close()
			appGuid := "my-cool-app-guid"

			instances, err := repo.GetInstances(appGuid)
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), err.IsNotSuccessful())

			assert.Equal(mr.T(), len(instances), 2)

			assert.Equal(mr.T(), instances[0].State, cf.InstanceRunning)
			assert.Equal(mr.T(), instances[1].State, cf.InstanceStarting)

			instance0 := instances[0]
			assert.Equal(mr.T(), instance0.Since, time.Unix(1379522342, 0))
			assert.Exactly(mr.T(), instance0.DiskQuota, uint64(1073741824))
			assert.Exactly(mr.T(), instance0.DiskUsage, uint64(56037376))
			assert.Exactly(mr.T(), instance0.MemQuota, uint64(67108864))
			assert.Exactly(mr.T(), instance0.MemUsage, uint64(19218432))
			assert.Equal(mr.T(), instance0.CpuUsage, 3.659571249238058e-05)
		})
	})
}
