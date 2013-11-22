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

var getAppSummariesResponseBody = `
{
  "apps":[
    {
      "guid":"app-1-guid",
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
      "routes":[
        {
          "guid":"route-2-guid",
          "host":"app2",
          "domain":{
            "guid":"domain-1-guid",
            "name":"cfapps.io"
          }
        },
        {
          "guid":"route-2-guid",
          "host":"foo",
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
	assert.Equal(t, len(app1.RouteSummaries), 1)
	assert.Equal(t, app1.RouteSummaries[0].URL(), "app1.cfapps.io")

	assert.Equal(t, app1.State, "started")
	assert.Equal(t, app1.InstanceCount, 1)
	assert.Equal(t, app1.RunningInstances, 1)
	assert.Equal(t, app1.Memory, uint64(128))

	app2 := apps[1]
	assert.Equal(t, app2.Name, "app2")
	assert.Equal(t, app2.Guid, "app-2-guid")
	assert.Equal(t, len(app2.RouteSummaries), 2)
	assert.Equal(t, app2.RouteSummaries[0].URL(), "app2.cfapps.io")
	assert.Equal(t, app2.RouteSummaries[1].URL(), "foo.cfapps.io")

	assert.Equal(t, app2.State, "started")
	assert.Equal(t, app2.InstanceCount, 3)
	assert.Equal(t, app2.RunningInstances, 1)
	assert.Equal(t, app2.Memory, uint64(512))
}

func createAppSummaryRepo(t *testing.T, requests []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo AppSummaryRepository) {
	ts, handler = testnet.NewTLSServer(t, requests)
	space := cf.SpaceFields{}
	space.Guid = "my-space-guid"
	config := &configuration.Configuration{
		SpaceFields: space,
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}

	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerAppSummaryRepository(config, gateway)
	return
}
