package spaces_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/api/security_groups/spaces"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SecurityGroupSpaceBinder", func() {
	var (
		repo        SecurityGroupSpaceBinder
		gateway     net.Gateway
		testServer  *httptest.Server
		testHandler *testnet.TestHandler
		configRepo  core_config.ReadWriter
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		gateway = net.NewCloudControllerGateway(configRepo, time.Now, &testterm.FakeUI{})
		repo = NewSecurityGroupSpaceBinder(configRepo, gateway)
	})

	AfterEach(func() { testServer.Close() })

	setupTestServer := func(reqs ...testnet.TestRequest) {
		testServer, testHandler = testnet.NewServer(reqs)
		configRepo.SetApiEndpoint(testServer.URL)
	}

	Describe(".BindSpace", func() {
		It("associates the security group with the space", func() {
			setupTestServer(
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "PUT",
					Path:   "/v2/security_groups/this-is-a-security-group-guid/spaces/yes-its-a-space-guid",
					Response: testnet.TestResponse{
						Status: http.StatusCreated,
						Body: `
{
  "metadata": {"guid": "fb6fdf81-ce1b-448f-ada9-09bbb8807812"},
  "entity": {"name": "dummy1", "rules": [] }
}`,
					},
				}))

			err := repo.BindSpace("this-is-a-security-group-guid", "yes-its-a-space-guid")

			Expect(err).ToNot(HaveOccurred())
			Expect(testHandler).To(HaveAllRequestsCalled())
		})
	})

	Describe(".UnbindSpace", func() {
		It("removes the associated security group from the space", func() {
			setupTestServer(
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "DELETE",
					Path:   "/v2/security_groups/this-is-a-security-group-guid/spaces/yes-its-a-space-guid",
					Response: testnet.TestResponse{
						Status: http.StatusNoContent,
					},
				}))

			err := repo.UnbindSpace("this-is-a-security-group-guid", "yes-its-a-space-guid")

			Expect(err).ToNot(HaveOccurred())
			Expect(testHandler).To(HaveAllRequestsCalled())
		})
	})
})
