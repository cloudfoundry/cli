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

var _ = Describe("AppStatsRepo", func() {
	It("returns stats for the app, given a guid", func() {
		ts, handler, repo := createAppStatsRepo([]testnet.TestRequest{
			appStatsRequest,
		})
		defer ts.Close()
		appGuid := "my-cool-app-guid"

		stats, err := repo.GetStats(appGuid)
		Expect(err).NotTo(HaveOccurred())
		Expect(handler).To(HaveAllRequestsCalled())

		Expect(len(stats)).To(Equal(2))

		Expect(stats[1].State).To(Equal(models.InstanceRunning))
		Expect(stats[0].State).To(Equal(models.InstanceFlapping))

		stats1 := stats[1]
		Expect(stats1.Stats.DiskQuota).To(Equal(uint64(10000)))
		Expect(stats1.Stats.Usage.Disk).To(Equal(uint64(10000)))
		Expect(stats1.Stats.MemQuota).To(Equal(uint64(1024)))
		Expect(stats1.Stats.Usage.Mem).To(Equal(uint64(1024)))
		Expect(stats1.Stats.Usage.Cpu).To(Equal(0.3))
	})
})

var appStatsRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/apps/my-cool-app-guid/stats",
	Response: testnet.TestResponse{Status: http.StatusOK, Body: `
{
  "1":{
    "state": "running",
    "stats": {
        "disk_quota": 10000,
        "mem_quota": 1024,
        "usage": {
            "cpu": 0.3,
            "disk": 10000,
            "mem": 1024,
            "time": "2014-07-14 23:33:55 +0000"
        }
    }
  },
  "0":{
    "state": "flapping",
    "stats": {
        "disk_quota": 1073741824,
        "mem_quota": 67108864,
        "usage": {
            "cpu": 3.659571249238058e-05,
            "disk": 56037376,
            "mem": 19218432,
            "time": "2014-07-14 23:33:55 +0000"
        }
    }
  }
}`}})

func createAppStatsRepo(requests []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo AppStatsRepository) {
	ts, handler = testnet.NewServer(requests)
	space := models.SpaceFields{}
	space.Guid = "my-space-guid"
	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway(configRepo, time.Now)
	repo = NewCloudControllerAppStatsRepository(configRepo, gateway)
	return
}
