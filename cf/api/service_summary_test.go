/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/github.com/cloudfoundry/cli/cf/commands/application/delete_app_test.go
   src/github.com/cloudfoundry/cli/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package api_test

import (
	. "github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/net"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
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

		serviceInstances, apiErr := repo.GetSummariesInCurrentSpace()
		Expect(handler).To(testnet.HaveAllRequestsCalled())

		Expect(apiErr).NotTo(HaveOccurred())
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
	ts, handler = testnet.NewServer([]testnet.TestRequest{req})
	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway(configRepo)
	repo = NewCloudControllerServiceSummaryRepository(configRepo, gateway)
	return
}
