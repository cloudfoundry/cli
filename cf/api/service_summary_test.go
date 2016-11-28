package api_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testnet "code.cloudfoundry.org/cli/util/testhelpers/net"

	. "code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceSummaryRepository", func() {
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
					  "last_operation": {
						  "type": "create",
						  "state": "in progress",
							"description": "50% done"
					  },
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

	It("gets a summary of services in the given space", func() {
		req := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "GET",
			Path:     "/v2/spaces/my-space-guid/summary",
			Response: serviceInstanceSummariesResponse,
		})

		ts, handler, repo := createServiceSummaryRepo(req)
		defer ts.Close()

		serviceInstances, apiErr := repo.GetSummariesInCurrentSpace()
		Expect(handler).To(HaveAllRequestsCalled())

		Expect(apiErr).NotTo(HaveOccurred())
		Expect(1).To(Equal(len(serviceInstances)))

		instance1 := serviceInstances[0]
		Expect(instance1.Name).To(Equal("my-service-instance"))
		Expect(instance1.LastOperation.Type).To(Equal("create"))
		Expect(instance1.LastOperation.State).To(Equal("in progress"))
		Expect(instance1.LastOperation.Description).To(Equal("50% done"))
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
	ts, handler = testnet.NewServer([]testnet.TestRequest{req})
	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetAPIEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
	repo = NewCloudControllerServiceSummaryRepository(configRepo, gateway)
	return
}
