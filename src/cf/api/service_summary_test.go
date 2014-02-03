package api_test

import (
	"cf"
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

func createServiceSummaryRepo(t mr.TestingT, req testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo ServiceSummaryRepository) {
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
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestServiceSummaryGetSummariesInCurrentSpace", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/spaces/my-space-guid/summary",
				Response: serviceInstanceSummariesResponse,
			})

			ts, handler, repo := createServiceSummaryRepo(mr.T(), req)
			defer ts.Close()

			serviceInstances, apiResponse := repo.GetSummariesInCurrentSpace()
			assert.True(mr.T(), handler.AllRequestsCalled())

			assert.True(mr.T(), apiResponse.IsSuccessful())
			assert.Equal(mr.T(), 1, len(serviceInstances))

			instance1 := serviceInstances[0]
			assert.Equal(mr.T(), instance1.Name, "my-service-instance")
			assert.Equal(mr.T(), instance1.ServicePlan.Name, "spark")
			assert.Equal(mr.T(), instance1.ServiceOffering.Label, "cleardb")
			assert.Equal(mr.T(), instance1.ServiceOffering.Label, "cleardb")
			assert.Equal(mr.T(), instance1.ServiceOffering.Provider, "cleardb-provider")
			assert.Equal(mr.T(), instance1.ServiceOffering.Version, "n/a")
			assert.Equal(mr.T(), len(instance1.ApplicationNames), 2)
			assert.Equal(mr.T(), instance1.ApplicationNames[0], "app1")
			assert.Equal(mr.T(), instance1.ApplicationNames[1], "app2")
		})
	})
}
