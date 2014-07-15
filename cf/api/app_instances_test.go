package api_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"

	. "github.com/cloudfoundry/cli/cf/api"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AppInstancesRepo", func() {
	It("returns instances of the app, given a guid", func() {
		ts, handler, repo := createAppInstancesRepo([]testnet.TestRequest{
			appInstancesRequest,
			appStatsRequest,
		})
		defer ts.Close()
		appGuid := "my-cool-app-guid"

		instances, err := repo.GetInstances(appGuid)
		Expect(err).NotTo(HaveOccurred())
		Expect(handler).To(HaveAllRequestsCalled())

		Expect(len(instances)).To(Equal(2))

		Expect(instances[0].State).To(Equal(models.InstanceRunning))
		Expect(instances[1].State).To(Equal(models.InstanceStarting))

		instance0 := instances[0]
		Expect(instance0.Since).To(Equal(time.Unix(1379522342, 0)))
		Expect(instance0.DiskQuota).To(Equal(uint64(1073741824)))
		Expect(instance0.DiskUsage).To(Equal(uint64(56037376)))
		Expect(instance0.MemQuota).To(Equal(uint64(67108864)))
		Expect(instance0.MemUsage).To(Equal(uint64(19218432)))
		Expect(instance0.CpuUsage).To(Equal(3.659571249238058e-05))
	})
})

var appInstancesRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/apps/my-cool-app-guid/instances",
	Response: testnet.TestResponse{Status: http.StatusOK, Body: `
{
  "1": {
    "state": "STARTING",
    "since": 1379522342.6783738
  },
  "0": {
    "state": "RUNNING",
    "since": 1379522342.6783738
  }
}`}})

func createAppInstancesRepo(requests []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo AppInstancesRepository) {
	ts, handler = testnet.NewServer(requests)
	space := models.SpaceFields{}
	space.Guid = "my-space-guid"
	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway(configRepo, time.Now)
	repo = NewCloudControllerAppInstancesRepository(configRepo, gateway)
	return
}
