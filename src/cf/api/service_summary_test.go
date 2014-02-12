package api_test

import (
	. "cf/api"
	"cf/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
)

var _ = Describe("Testing with ginkgo", func() {
	var serviceInstanceSummariesResponse testnet.TestResponse

	BeforeEach(func() {
		serviceInstanceSummariesResponse = testnet.TestResponse{Status: http.StatusOK, Body: `
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
			}`,
		}
	})

	It("TestServiceSummaryGetSummariesInCurrentSpace", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "GET",
			Path:     "/v2/spaces/my-space-guid/summary",
			Response: serviceInstanceSummariesResponse,
		})

		ts, handler, repo := createServiceSummaryRepo(req)
		defer ts.Close()

		serviceInstances, apiResponse := repo.GetSummariesInCurrentSpace()
		Expect(handler.AllRequestsCalled()).To(BeTrue())

		Expect(apiResponse.IsSuccessful()).To(BeTrue())
		Expect(1).To(Equal(len(serviceInstances)))

		instance1 := serviceInstances[0]
		Expect(instance1.Name).To(Equal("my-service-instance"))
		Expect(instance1.ServicePlan.Name).To(Equal("spark"))
		Expect(instance1.ServiceOffering.Label).To(Equal("cleardb"))
		Expect(instance1.ServiceOffering.Label).To(Equal("cleardb"))
		Expect(instance1.ServiceOffering.Provider).To(Equal("cleardb-provider"))
		Expect(instance1.ServiceOffering.Version).To(Equal("n/a"))
		Expect(len(instance1.ApplicationNames)).To(Equal(2))
		Expect(instance1.ApplicationNames[0]).To(Equal("app1"))
		Expect(instance1.ApplicationNames[1]).To(Equal("app2"))
	})
})

func createServiceSummaryRepo(req testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo ServiceSummaryRepository) {
	ts, handler = testnet.NewTLSServer(GinkgoT(), []testnet.TestRequest{req})
	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerServiceSummaryRepository(configRepo, gateway)
	return
}
