package api_test

import (
	"cf"
	. "cf/api"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testhelpers"
	"testing"
	"time"
)

var instancesEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/apps/my-cool-app-guid/instances",
	nil,
	testhelpers.TestResponse{Status: http.StatusOK, Body: `
{
  "1": {
    "state": "STARTING",
    "since": 1379522342.6783738
  },
  "0": {
    "state": "RUNNING",
    "since": 1379522342.6783738
  }
}`},
)

var statsEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/apps/my-cool-app-guid/stats",
	nil,
	testhelpers.TestResponse{Status: http.StatusOK, Body: `
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
}`})

var appDetailsEndpoints = func(writer http.ResponseWriter, request *http.Request) {
	if strings.HasSuffix(request.URL.Path, "/instances") {
		instancesEndpoint(writer, request)
		return
	}
	if strings.HasSuffix(request.URL.Path, "/stats") {
		statsEndpoint(writer, request)
		return
	}
}

func TestAppSummaryGetSummary(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(appDetailsEndpoints))
	defer ts.Close()

	config := configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}

	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})

	appRepo := NewCloudControllerApplicationRepository(config, gateway)
	summaryRepo := NewCloudControllerAppSummaryRepository(config, gateway, appRepo)

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}

	summary, err := summaryRepo.GetSummary(app)
	assert.False(t, err.IsError())

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
