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

var serviceInstanceSummariesResponse = testapi.TestResponse{Status: http.StatusOK, Body: `
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
	endpoint, status := testapi.CreateCheckableEndpoint(
		"GET",
		"/v2/spaces/my-space-guid/summary",
		nil,
		serviceInstanceSummariesResponse,
	)

	ts, repo := createServiceSummaryRepo(endpoint)
	defer ts.Close()

	serviceInstances, apiResponse := repo.GetSummariesInCurrentSpace()
	assert.True(t, status.Called())

	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, 1, len(serviceInstances))

	instance1 := serviceInstances[0]
	assert.Equal(t, instance1.Name, "my-service-instance")
	assert.Equal(t, instance1.ServicePlan.Name, "spark")
	assert.Equal(t, instance1.ServiceOffering().Label, "cleardb")
	assert.Equal(t, instance1.ServicePlan.ServiceOffering.Label, "cleardb")
	assert.Equal(t, instance1.ServicePlan.ServiceOffering.Provider, "cleardb-provider")
	assert.Equal(t, instance1.ServicePlan.ServiceOffering.Version, "n/a")
	assert.Equal(t, len(instance1.ApplicationNames), 2)
	assert.Equal(t, instance1.ApplicationNames[0], "app1")
	assert.Equal(t, instance1.ApplicationNames[1], "app2")
}

func createServiceSummaryRepo(endpoint http.HandlerFunc) (ts *httptest.Server, repo ServiceSummaryRepository) {
	ts = httptest.NewTLSServer(endpoint)

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
	}
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerServiceSummaryRepository(config, gateway)
	return
}
