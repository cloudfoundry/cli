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
	"time"
)

var getAppSummariesResponseBody = `
{
  "apps":[
    {
      "guid":"app-1-guid",
      "urls":["app1.cfapps.io"],
      "routes":[
        {
          "guid":"route-1-guid",
          "host":"app1",
          "domain":{
            "guid":"domain-1-guid",
            "name":"cfapps.io"
          }
        }
      ],
      "running_instances":1,
      "name":"app1",
      "memory":128,
      "instances":1,
      "state":"STARTED",
      "service_names":[
      	"my-service-instance"
      ]
    },{
      "guid":"app-2-guid",
      "urls":["app2.cfapps.io", "foo.cfapps.io"],
      "routes":[
        {
          "guid":"route-2-guid",
          "host":"app2",
          "domain":{
            "guid":"domain-1-guid",
            "name":"cfapps.io"
          }
        }
      ],
      "running_instances":1,
      "name":"app2",
      "memory":512,
      "instances":3,
      "state":"STARTED",
      "service_names":[
      	"my-service-instance"
      ]
    }
  ]
}`

func TestGetAppSummariesInCurrentSpace(t *testing.T) {
	getAppSummariesRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/spaces/my-space-guid/summary",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: getAppSummariesResponseBody},
	})

	ts, handler, repo := createAppSummaryRepo(t, []testnet.TestRequest{getAppSummariesRequest})
	defer ts.Close()

	apps, apiResponse := repo.GetSummariesInCurrentSpace()
	assert.True(t, handler.AllRequestsCalled())

	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, 2, len(apps))

	app1 := apps[0]
	assert.Equal(t, app1.Name, "app1")
	assert.Equal(t, app1.Guid, "app-1-guid")
	assert.Equal(t, len(app1.Urls), 1)
	assert.Equal(t, app1.Urls[0], "app1.cfapps.io")

	assert.Equal(t, app1.State, "started")
	assert.Equal(t, app1.Instances, 1)
	assert.Equal(t, app1.RunningInstances, 1)
	assert.Equal(t, app1.Memory, uint64(128))

	app2 := apps[1]
	assert.Equal(t, app2.Name, "app2")
	assert.Equal(t, app2.Guid, "app-2-guid")
	assert.Equal(t, len(app2.Urls), 2)
	assert.Equal(t, app2.Urls[0], "app2.cfapps.io")
	assert.Equal(t, app2.Urls[1], "foo.cfapps.io")

	assert.Equal(t, app2.State, "started")
	assert.Equal(t, app2.Instances, 3)
	assert.Equal(t, app2.RunningInstances, 1)
	assert.Equal(t, app2.Memory, uint64(512))
}

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

var appSummaryRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/apps/my-cool-app-guid/summary",
	Response: testnet.TestResponse{Status: http.StatusOK, Body: `
{
  "guid": "app1-guid",
  "name": "App1",
  "routes": [
    {
      "guid": "route-1-guid",
      "host": "app1",
      "domain": {
        "guid": "domain-1-guid",
        "name": "cfapps.io"
      }
    }
  ],
  "running_instances": 1,
  "memory": 128,
  "instances": 1
}`}})

func TestAppSummaryGetSummary(t *testing.T) {
	ts, handler, repo := createAppSummaryRepo(t, []testnet.TestRequest{
		appSummaryRequest,
		appInstancesRequest,
		appStatsRequest,
	})
	defer ts.Close()

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}

	summary, err := repo.GetSummary(app)
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, err.IsNotSuccessful())

	assert.Equal(t, summary.App.Name, app.Name)

	assert.Equal(t, len(summary.Instances), 2)

	instance0 := summary.Instances[0]
	instance1 := summary.Instances[1]
	assert.Equal(t, instance0.State, cf.InstanceRunning)
	assert.Equal(t, instance1.State, cf.InstanceStarting)

	time0 := time.Unix(1379522342, 0)
	assert.Equal(t, instance0.Since, time0)
	assert.Exactly(t, instance0.DiskQuota, uint64(1073741824))
	assert.Exactly(t, instance0.DiskUsage, uint64(56037376))
	assert.Exactly(t, instance0.MemQuota, uint64(67108864))
	assert.Exactly(t, instance0.MemUsage, uint64(19218432))
	assert.Equal(t, instance0.CpuUsage, 3.659571249238058e-05)
}

func createAppSummaryRepo(t *testing.T, requests []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo AppSummaryRepository) {
	ts, handler = testnet.NewTLSServer(t, requests)

	config := &configuration.Configuration{
		Space:       cf.Space{Guid: "my-space-guid"},
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}

	gateway := net.NewCloudControllerGateway()
	appRepo := NewCloudControllerApplicationRepository(config, gateway)
	repo = NewCloudControllerAppSummaryRepository(config, gateway, appRepo)
	return
}
