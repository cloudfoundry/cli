package copy_application_source_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/cloudfoundry/cli/cf/api/copy_application_source"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CopyApplicationSource", func() {
	var (
		repo       CopyApplicationSourceRepository
		testServer *httptest.Server
		configRepo core_config.ReadWriter
	)

	setupTestServer := func(reqs ...testnet.TestRequest) {
		testServer, _ = testnet.NewServer(reqs)
		configRepo.SetApiEndpoint(testServer.URL)
	}

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		gateway := net.NewCloudControllerGateway(configRepo, time.Now, &testterm.FakeUI{})
		repo = NewCloudControllerCopyApplicationSourceRepository(configRepo, gateway)
	})

	AfterEach(func() {
		testServer.Close()
	})

	Describe(".CopyApplication", func() {
		BeforeEach(func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "POST",
				Path:   "/v2/apps/target-app-guid/copy_bits",
				Matcher: testnet.RequestBodyMatcher(`{
					"source_app_guid": "source-app-guid"
				}`),
				Response: testnet.TestResponse{
					Status: http.StatusCreated,
				},
			}))
		})

		It("should return a CopyApplicationModel", func() {
			err := repo.CopyApplication("source-app-guid", "target-app-guid")
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
