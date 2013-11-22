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

var serviceInstanceSummariesResponse = testnet.TestResponse{Status: http.StatusOK, Body: `
{
  "apps":[
    {
      "name":"app1",
      "service_names":[
      	"my-service-instance"
      ]
    },{
      "name":"app2",
      "service_names":[
      	"my-service-instance"
      ]
    }
  ],
  "services": [
    {
      "guid": "my-service-instance-guid",
      "name": "my-service-instance",
      "bound_app_count": 2,
      "service_plan": {
        "guid": "service-plan-guid",
        "name": "spark",
        "service": {
          "guid": "service-offering-guid",
          "label": "cleardb",
          "provider": "cleardb-provider",
          "version": "n/a"
        }
      }
    }
  ]
}`}

func TestServiceSummaryGetSummariesInCurrentSpace(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/spaces/my-space-guid/summary",
		Response: serviceInstanceSummariesResponse,
	})

	ts, handler, repo := createServiceSummaryRepo(t, req)
	defer ts.Close()

	serviceInstances, apiResponse := repo.GetSummariesInCurrentSpace()
	assert.True(t, handler.AllRequestsCalled())

	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, 1, len(serviceInstances))

	instance1 := serviceInstances[0]
	assert.Equal(t, instance1.Name, "my-service-instance")
	assert.Equal(t, instance1.ServicePlan.Name, "spark")
	assert.Equal(t, instance1.ServiceOffering.Label, "cleardb")
	assert.Equal(t, instance1.ServiceOffering.Label, "cleardb")
	assert.Equal(t, instance1.ServiceOffering.Provider, "cleardb-provider")
	assert.Equal(t, instance1.ServiceOffering.Version, "n/a")
	assert.Equal(t, len(instance1.ApplicationNames), 2)
	assert.Equal(t, instance1.ApplicationNames[0], "app1")
	assert.Equal(t, instance1.ApplicationNames[1], "app2")
}

func createServiceSummaryRepo(t *testing.T, req testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo ServiceSummaryRepository) {
	ts, handler = testnet.NewTLSServer(t, []testnet.TestRequest{req})
	space := cf.SpaceFields{}
	space.Guid = "my-space-guid"
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		SpaceFields: space,
	}
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerServiceSummaryRepository(config, gateway)
	return
}
