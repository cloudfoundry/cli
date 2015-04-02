package api_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/net"

	. "github.com/cloudfoundry/cli/cf/api"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Keys Repo", func() {
	var (
		testServer  *httptest.Server
		testHandler *testnet.TestHandler
		configRepo  core_config.ReadWriter
		repo        ServiceKeyRepository
	)

	setupTestServer := func(reqs ...testnet.TestRequest) {
		testServer, testHandler = testnet.NewServer(reqs)
		configRepo.SetApiEndpoint(testServer.URL)
	}

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		configRepo.SetAccessToken("BEARER my_access_token")

		gateway := net.NewCloudControllerGateway(configRepo, time.Now, &testterm.FakeUI{})
		repo = NewCloudControllerServiceKeyRepository(configRepo, gateway)
	})

	Describe("creating a service key", func() {
		It("makes the right request", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "POST",
				Path:     "/v2/service_keys",
				Matcher:  testnet.RequestBodyMatcher(`{"service_instance_guid": "fake-instance-guid", "name": "fake-key-name"}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			}))

			err := repo.CreateServiceKey("fake-instance-guid", "fake-key-name")
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns a ModelAlreadyExistsError if the service key exists", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:  "POST",
				Path:    "/v2/service_keys",
				Matcher: testnet.RequestBodyMatcher(`{"service_instance_guid":"fake-instance-guid","name":"exist-service-key"}`),
				Response: testnet.TestResponse{
					Status: http.StatusBadRequest,
					Body:   `{"code":360001,"description":"The service key name is taken: exist-service-key"}`},
			}))

			err := repo.CreateServiceKey("fake-instance-guid", "exist-service-key")
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).To(BeAssignableToTypeOf(&errors.ModelAlreadyExistsError{}))
		})

		It("returns a NotAuthorizedError when CLI user is not the space developer or admin", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:  "POST",
				Path:    "/v2/service_keys",
				Matcher: testnet.RequestBodyMatcher(`{"service_instance_guid":"fake-instance-guid","name":"fake-service-key"}`),
				Response: testnet.TestResponse{
					Status: http.StatusBadRequest,
					Body:   `{"code":10003,"description":"You are not authorized to perform the requested action"}`},
			}))

			err := repo.CreateServiceKey("fake-instance-guid", "fake-service-key")
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("You are not authorized to perform the requested action"))
		})
	})

	AfterEach(func() {
		testServer.Close()
	})
})
