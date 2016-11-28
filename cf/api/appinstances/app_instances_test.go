package appinstances_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testnet "code.cloudfoundry.org/cli/util/testhelpers/net"

	. "code.cloudfoundry.org/cli/cf/api/appinstances"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AppInstancesRepo", func() {
	Describe("Getting the instances for an application", func() {
		It("returns instances of the app, given a guid", func() {
			ts, handler, repo := createAppInstancesRepo([]testnet.TestRequest{
				appInstancesRequest,
				appStatsRequest,
			})
			defer ts.Close()
			appGUID := "my-cool-app-guid"

			instances, err := repo.GetInstances(appGUID)
			Expect(err).NotTo(HaveOccurred())
			Expect(handler).To(HaveAllRequestsCalled())

			Expect(len(instances)).To(Equal(2))

			Expect(instances[0].State).To(Equal(models.InstanceRunning))
			Expect(instances[1].State).To(Equal(models.InstanceStarting))
			Expect(instances[1].Details).To(Equal("insufficient resources"))

			instance0 := instances[0]
			Expect(instance0.Since).To(Equal(time.Unix(1379522342, 0)))
			Expect(instance0.DiskQuota).To(Equal(int64(1073741824)))
			Expect(instance0.DiskUsage).To(Equal(int64(56037376)))
			Expect(instance0.MemQuota).To(Equal(int64(67108864)))
			Expect(instance0.MemUsage).To(Equal(int64(19218432)))
			Expect(instance0.CPUUsage).To(Equal(3.659571249238058e-05))
		})
	})

	Describe("Deleting an instance for an application", func() {
		It("returns no error if the response is successful", func() {
			ts, handler, repo := createAppInstancesRepo([]testnet.TestRequest{
				deleteInstanceRequest,
			})
			defer ts.Close()
			appGUID := "my-cool-app-guid"

			err := repo.DeleteInstance(appGUID, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(handler).To(HaveAllRequestsCalled())
		})

		It("returns the error if the response is unsuccessful", func() {
			ts, handler, repo := createAppInstancesRepo([]testnet.TestRequest{
				deleteInstanceFromUnkownApp,
			})
			defer ts.Close()
			appGUID := "some-wrong-app-guid"

			err := repo.DeleteInstance(appGUID, 0)
			Expect(err).To(HaveOccurred())
			Expect(handler).To(HaveAllRequestsCalled())

		})
	})
})

var appStatsRequest = apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/apps/my-cool-app-guid/stats",
	Response: testnet.TestResponse{Status: http.StatusOK, Body: `
{
  "1":{
    "stats": {
        "disk_quota": 10000,
        "mem_quota": 1024,
        "usage": {
            "cpu": 0.3,
            "disk": 10000,
            "mem": 1024
        }
    }
  },
  "0":{
    "stats": {
        "disk_quota": 1073741824,
        "mem_quota": 67108864,
        "usage": {
            "cpu": 3.659571249238058e-05,
            "disk": 56037376,
            "mem": 19218432
        }
    }
  }
}`}})

var appInstancesRequest = apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/apps/my-cool-app-guid/instances",
	Response: testnet.TestResponse{Status: http.StatusOK, Body: `
{
  "1": {
    "state": "STARTING",
    "details": "insufficient resources",
    "since": 1379522342.6783738
  },
  "0": {
    "state": "RUNNING",
    "since": 1379522342.6783738
  }
}`}})

var deleteInstanceRequest = apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
	Method:   "DELETE",
	Path:     "/v2/apps/my-cool-app-guid/instances/0",
	Response: testnet.TestResponse{Status: http.StatusNoContent, Body: `{}`},
})

var deleteInstanceFromUnkownApp = apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
	Method:   "DELETE",
	Path:     "/v2/apps/some-wrong-app-guid/instances/0",
	Response: testnet.TestResponse{Status: http.StatusNotFound, Body: `{}`},
})

func createAppInstancesRepo(requests []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo Repository) {
	ts, handler = testnet.NewServer(requests)
	space := models.SpaceFields{}
	space.GUID = "my-space-guid"
	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetAPIEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
	repo = NewCloudControllerAppInstancesRepository(configRepo, gateway)
	return
}
