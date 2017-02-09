package api_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	. "code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	testnet "code.cloudfoundry.org/cli/util/testhelpers/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AppSummaryRepository", func() {
	var (
		testServer *httptest.Server
		handler    *testnet.TestHandler
		repo       AppSummaryRepository
	)

	Describe("GetSummariesInCurrentSpace()", func() {
		BeforeEach(func() {
			getAppSummariesRequest := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/spaces/my-space-guid/summary",
				Response: testnet.TestResponse{
					Status: http.StatusOK,
					Body:   getAppSummariesResponseBody,
				},
			})

			testServer, handler = testnet.NewServer([]testnet.TestRequest{getAppSummariesRequest})
			configRepo := testconfig.NewRepositoryWithDefaults()
			configRepo.SetAPIEndpoint(testServer.URL)
			gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
			repo = NewCloudControllerAppSummaryRepository(configRepo, gateway)
		})

		AfterEach(func() {
			testServer.Close()
		})

		It("returns a slice of app summaries for each instance", func() {
			apps, apiErr := repo.GetSummariesInCurrentSpace()
			Expect(handler).To(HaveAllRequestsCalled())

			Expect(apiErr).NotTo(HaveOccurred())
			Expect(3).To(Equal(len(apps)))

			app1 := apps[0]
			Expect(app1.Name).To(Equal("app1"))
			Expect(app1.GUID).To(Equal("app-1-guid"))
			Expect(app1.BuildpackURL).To(Equal("go_buildpack"))
			Expect(len(app1.Routes)).To(Equal(1))
			Expect(app1.Routes[0].URL()).To(Equal("app1.cfapps.io"))

			Expect(app1.State).To(Equal("started"))
			Expect(app1.Command).To(Equal("start_command"))
			Expect(app1.InstanceCount).To(Equal(1))
			Expect(app1.RunningInstances).To(Equal(1))
			Expect(app1.Memory).To(Equal(int64(128)))
			Expect(app1.PackageUpdatedAt.Format("2006-01-02T15:04:05Z07:00")).To(Equal("2014-10-24T19:54:00Z"))
			Expect(app1.AppPorts).To(Equal([]int{8080, 9090}))

			app2 := apps[1]
			Expect(app2.Name).To(Equal("app2"))
			Expect(app2.Command).To(Equal(""))
			Expect(app2.GUID).To(Equal("app-2-guid"))
			Expect(len(app2.Routes)).To(Equal(2))
			Expect(app2.Routes[0].URL()).To(Equal("app2.cfapps.io"))
			Expect(app2.Routes[1].URL()).To(Equal("foo.cfapps.io"))
			Expect(app2.AppPorts).To(HaveLen(0))

			Expect(app2.State).To(Equal("started"))
			Expect(app2.InstanceCount).To(Equal(3))
			Expect(app2.RunningInstances).To(Equal(1))
			Expect(app2.Memory).To(Equal(int64(512)))
			Expect(app2.PackageUpdatedAt.Format("2006-01-02T15:04:05Z07:00")).To(Equal("2012-10-24T19:55:00Z"))

			nullUpdateAtApp := apps[2]
			Expect(nullUpdateAtApp.PackageUpdatedAt).To(BeNil())
		})
	})

	Describe("GetSummary()", func() {
		BeforeEach(func() {
			getAppSummaryRequest := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/apps/app1-guid/summary",
				Response: testnet.TestResponse{
					Status: http.StatusOK,
					Body:   getAppSummaryResponseBody,
				},
			})

			testServer, handler = testnet.NewServer([]testnet.TestRequest{getAppSummaryRequest})
			configRepo := testconfig.NewRepositoryWithDefaults()
			configRepo.SetAPIEndpoint(testServer.URL)
			gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
			repo = NewCloudControllerAppSummaryRepository(configRepo, gateway)
		})

		AfterEach(func() {
			testServer.Close()
		})

		It("returns the app summary", func() {
			app, apiErr := repo.GetSummary("app1-guid")
			Expect(handler).To(HaveAllRequestsCalled())

			Expect(apiErr).NotTo(HaveOccurred())

			Expect(app.Name).To(Equal("app1"))
			Expect(app.GUID).To(Equal("app-1-guid"))
			Expect(app.BuildpackURL).To(Equal("go_buildpack"))
			Expect(len(app.Routes)).To(Equal(1))
			Expect(app.Routes[0].URL()).To(Equal("app1.cfapps.io"))

			Expect(app.State).To(Equal("started"))
			Expect(app.Command).To(Equal("start_command"))
			Expect(app.InstanceCount).To(Equal(1))
			Expect(app.RunningInstances).To(Equal(1))
			Expect(app.Memory).To(Equal(int64(128)))
			Expect(app.PackageUpdatedAt.Format("2006-01-02T15:04:05Z07:00")).To(Equal("2014-10-24T19:54:00Z"))
			Expect(app.StackGUID).To(Equal("the-stack-guid"))
			Expect(app.HealthCheckType).To(Equal("some-health-check-type"))
			Expect(app.HealthCheckHTTPEndpoint).To(Equal("/some-endpoint"))
		})
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
			"command": "start_command",
      "instances":1,
			"buildpack":"go_buildpack",
      "state":"STARTED",
      "service_names":[
      	"my-service-instance"
      ],
			"package_updated_at":"2014-10-24T19:54:00+00:00",
			"ports":[
				8080,
				9090
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
      ],
			"package_updated_at":"2012-10-24T19:55:00+00:00",
			"ports":null
    },{
      "guid":"app-with-null-updated-at-guid",
      "routes":[
        {
          "guid":"route-3-guid",
          "host":"app3",
          "domain":{
            "guid":"domain-3-guid",
            "name":"cfapps.io"
          }
        }
      ],
      "running_instances":1,
      "name":"app-with-null-updated-at",
			"memory":512,
      "instances":3,
      "state":"STARTED",
      "service_names":[
      	"my-service-instance"
      ],
			"package_updated_at":null,
			"ports":null
    }
  ]
}`

const getAppSummaryResponseBody string = `
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
		"stack_guid":"the-stack-guid",
		"memory":128,
		"command": "start_command",
		"instances":1,
		"buildpack":"go_buildpack",
		"state":"STARTED",
		"service_names":[
			"my-service-instance"
		],
		"package_updated_at":"2014-10-24T19:54:00+00:00",
		"health_check_type":"some-health-check-type",
		"health_check_http_endpoint":"/some-endpoint"
}`
