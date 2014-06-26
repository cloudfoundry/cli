package running_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"

	. "github.com/cloudfoundry/cli/cf/api/security_groups/defaults/running"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RunningSecurityGroupsRepo", func() {
	var (
		testServer  *httptest.Server
		testHandler *testnet.TestHandler
		configRepo  configuration.ReadWriter
		repo        RunningSecurityGroupsRepo
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		gateway := net.NewCloudControllerGateway((configRepo), time.Now)
		repo = NewRunningSecurityGroupsRepo(configRepo, gateway)
	})

	AfterEach(func() {
		testServer.Close()
	})

	setupTestServer := func(reqs ...testnet.TestRequest) {
		testServer, testHandler = testnet.NewServer(reqs)
		configRepo.SetApiEndpoint(testServer.URL)
	}

	Describe(".AddToDefaultRunningSet", func() {
		It("makes a correct request", func() {
			setupTestServer(
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "PUT",
					Path:   "/v2/config/running_security_groups/a-real-guid",
					Response: testnet.TestResponse{
						Status: http.StatusCreated,
						Body:   addRunningResponse,
					},
				}),
			)

			err := repo.AddToDefaultRunningSet("a-real-guid")

			Expect(err).ToNot(HaveOccurred())
			Expect(testHandler).To(HaveAllRequestsCalled())
		})
	})
})

var addRunningResponse string = `{
  "metadata": {
    "guid": "897341eb-ef31-406f-b57b-414f51583a3a",
    "url": "/v2/config/running_security_groups/897341eb-ef31-406f-b57b-414f51583a3a",
    "created_at": "2014-06-23T21:43:30+00:00",
    "updated_at": "2014-06-23T21:43:30+00:00"
  },
  "entity": {
    "name": "name-904",
    "rules": [
      {
        "protocol": "udp",
        "ports": "8080",
        "destination": "198.41.191.47/1"
      }
    ]
  }
}`
