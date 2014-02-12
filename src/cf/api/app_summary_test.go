package api_test

import (
	. "cf/api"
	"cf/net"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
)

var _ = Describe("AppSummaryRepository", func() {
	It("TestGetAppSummariesInCurrentSpace", func() {
		getAppSummariesRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "GET",
			Path:     "/v2/spaces/my-space-guid/summary",
			Response: testnet.TestResponse{Status: http.StatusOK, Body: getAppSummariesResponseBody},
		})

		ts, handler, repo := createAppSummaryRepo(mr.T(), []testnet.TestRequest{getAppSummariesRequest})
		defer ts.Close()

		apps, apiResponse := repo.GetSummariesInCurrentSpace()
		assert.True(mr.T(), handler.AllRequestsCalled())

		assert.True(mr.T(), apiResponse.IsSuccessful())
		assert.Equal(mr.T(), 2, len(apps))

		app1 := apps[0]
		assert.Equal(mr.T(), app1.Name, "app1")
		assert.Equal(mr.T(), app1.Guid, "app-1-guid")
		assert.Equal(mr.T(), len(app1.RouteSummaries), 1)
		assert.Equal(mr.T(), app1.RouteSummaries[0].URL(), "app1.cfapps.io")

		assert.Equal(mr.T(), app1.State, "started")
		assert.Equal(mr.T(), app1.InstanceCount, 1)
		assert.Equal(mr.T(), app1.RunningInstances, 1)
		assert.Equal(mr.T(), app1.Memory, uint64(128))

		app2 := apps[1]
		assert.Equal(mr.T(), app2.Name, "app2")
		assert.Equal(mr.T(), app2.Guid, "app-2-guid")
		assert.Equal(mr.T(), len(app2.RouteSummaries), 2)
		assert.Equal(mr.T(), app2.RouteSummaries[0].URL(), "app2.cfapps.io")
		assert.Equal(mr.T(), app2.RouteSummaries[1].URL(), "foo.cfapps.io")

		assert.Equal(mr.T(), app2.State, "started")
		assert.Equal(mr.T(), app2.InstanceCount, 3)
		assert.Equal(mr.T(), app2.RunningInstances, 1)
		assert.Equal(mr.T(), app2.Memory, uint64(512))
	})
})

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

func createAppSummaryRepo(t mr.TestingT, requests []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo AppSummaryRepository) {
	ts, handler = testnet.NewTLSServer(t, requests)
	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerAppSummaryRepository(configRepo, gateway)
	return
}
