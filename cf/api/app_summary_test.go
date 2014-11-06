package api_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/api"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AppSummaryRepository", func() {
	var (
		testServer *httptest.Server
		handler    *testnet.TestHandler
		repo       AppSummaryRepository
	)

	BeforeEach(func() {
		getAppSummariesRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "GET",
			Path:   "/v2/spaces/my-space-guid/summary",
			Response: testnet.TestResponse{
				Status: http.StatusOK,
				Body:   getAppSummariesResponseBody,
			},
		})

		testServer, handler = testnet.NewServer([]testnet.TestRequest{getAppSummariesRequest})
		configRepo := testconfig.NewRepositoryWithDefaults()
		configRepo.SetApiEndpoint(testServer.URL)
		gateway := net.NewCloudControllerGateway(configRepo, time.Now, &testterm.FakeUI{})
		repo = NewCloudControllerAppSummaryRepository(configRepo, gateway)
	})

	AfterEach(func() {
		testServer.Close()
	})

	It("returns a slice of app summaries for each instance", func() {
		apps, apiErr := repo.GetSummariesInCurrentSpace()
		Expect(handler).To(HaveAllRequestsCalled())

		Expect(apiErr).NotTo(HaveOccurred())
		Expect(2).To(Equal(len(apps)))

		app1 := apps[0]
		Expect(app1.Name).To(Equal("app1"))
		Expect(app1.Guid).To(Equal("app-1-guid"))
		Expect(len(app1.Routes)).To(Equal(1))
		Expect(app1.Routes[0].URL()).To(Equal("app1.cfapps.io"))

		Expect(app1.State).To(Equal("started"))
		Expect(app1.InstanceCount).To(Equal(1))
		Expect(app1.RunningInstances).To(Equal(1))
		Expect(app1.Memory).To(Equal(int64(128)))
		Expect(app1.PackageUpdatedAt.Format("2006-01-02T15:04:05Z07:00")).To(Equal("2014-10-24T19:54:00Z"))

		app2 := apps[1]
		Expect(app2.Name).To(Equal("app2"))
		Expect(app2.Guid).To(Equal("app-2-guid"))
		Expect(len(app2.Routes)).To(Equal(2))
		Expect(app2.Routes[0].URL()).To(Equal("app2.cfapps.io"))
		Expect(app2.Routes[1].URL()).To(Equal("foo.cfapps.io"))

		Expect(app2.State).To(Equal("started"))
		Expect(app2.InstanceCount).To(Equal(3))
		Expect(app2.RunningInstances).To(Equal(1))
		Expect(app2.Memory).To(Equal(int64(512)))
		Expect(app2.PackageUpdatedAt.Format("2006-01-02T15:04:05Z07:00")).To(Equal("2012-10-24T19:55:00Z"))
	})
})

const getAppSummariesResponseBody string = `
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
      ],
			"package_updated_at":"2014-10-24T19:54:00+00:00"
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
      ],
			"package_updated_at":"2012-10-24T19:55:00+00:00"
    }
  ]
}`
