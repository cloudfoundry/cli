package api_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"

	. "github.com/cloudfoundry/cli/cf/api"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StagingSecurityGroupsRepo", func() {
	var (
		testServer  *httptest.Server
		testHandler *testnet.TestHandler
		configRepo  configuration.ReadWriter
		repo        StagingSecurityGroupsRepo
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		gateway := net.NewCloudControllerGateway((configRepo), time.Now)
		repo = NewStagingSecurityGroupsRepo(configRepo, gateway)
	})

	AfterEach(func() {
		testServer.Close()
	})

	Describe("AddToDefaultStagingSet", func() {
		It("makes a correct request", func() {
			testServer, testHandler = testnet.NewServer([]testnet.TestRequest{
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "POST",
					Path:   "/v2/config/staging_security_groups/a-real-guid",
					Response: testnet.TestResponse{
						Status: http.StatusCreated,
						Body: `
{
  "metadata": {
    "guid": "897341eb-ef31-406f-b57b-414f51583a3a",
    "url": "/v2/config/staging_security_groups/897341eb-ef31-406f-b57b-414f51583a3a",
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
}
`,
					},
				}),
			})

			configRepo.SetApiEndpoint(testServer.URL)
			err := repo.AddToDefaultStagingSet(models.ApplicationSecurityGroupFields{Guid: "a-real-guid"})

			Expect(err).ToNot(HaveOccurred())
			Expect(testHandler).To(testnet.HaveAllRequestsCalled())
		})
	})
})
