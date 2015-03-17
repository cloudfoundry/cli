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

	Describe("listing service keys", func() {
		It("returns empty result when no service key is found", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/service_keys?q=service_instance_guid:fake-instance-guid",
				Response: emptyServiceKeysResponse,
			}))

			serviceKeys, err := repo.ListServiceKeys("fake-instance-guid")
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).NotTo(HaveOccurred())
			Expect(len(serviceKeys)).To(Equal(0))
		})

		It("returns correctly when service keys are found", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/service_keys?q=service_instance_guid:fake-instance-guid",
				Response: serviceKeysResponse,
			}))

			serviceKeys, err := repo.ListServiceKeys("fake-instance-guid")
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).NotTo(HaveOccurred())
			Expect(len(serviceKeys)).To(Equal(2))

			Expect(serviceKeys[0].Fields.Guid).To(Equal("fake-service-key-guid-1"))
			Expect(serviceKeys[0].Fields.Url).To(Equal("/v2/service_keys/fake-guid-1"))
			Expect(serviceKeys[0].Fields.Name).To(Equal("fake-service-key-name-1"))
			Expect(serviceKeys[0].Fields.ServiceInstanceGuid).To(Equal("fake-service-instance-guid-1"))
			Expect(serviceKeys[0].Fields.ServiceInstanceUrl).To(Equal("http://fake/service/instance/url/1"))

			Expect(serviceKeys[0].Credentials).To(HaveKeyWithValue("username", "fake-username-1"))
			Expect(serviceKeys[0].Credentials).To(HaveKeyWithValue("password", "fake-password-1"))
			Expect(serviceKeys[0].Credentials).To(HaveKeyWithValue("host", "fake-host-1"))
			Expect(serviceKeys[0].Credentials).To(HaveKeyWithValue("port", float64(3306)))
			Expect(serviceKeys[0].Credentials).To(HaveKeyWithValue("database", "fake-db-name-1"))
			Expect(serviceKeys[0].Credentials).To(HaveKeyWithValue("uri", "mysql://fake-user-1:fake-password-1@fake-host-1:3306/fake-db-name-1"))

			Expect(serviceKeys[1].Fields.Guid).To(Equal("fake-service-key-guid-2"))
			Expect(serviceKeys[1].Fields.Url).To(Equal("/v2/service_keys/fake-guid-2"))
			Expect(serviceKeys[1].Fields.Name).To(Equal("fake-service-key-name-2"))
			Expect(serviceKeys[1].Fields.ServiceInstanceGuid).To(Equal("fake-service-instance-guid-2"))
			Expect(serviceKeys[1].Fields.ServiceInstanceUrl).To(Equal("http://fake/service/instance/url/1"))

			Expect(serviceKeys[1].Credentials).To(HaveKeyWithValue("username", "fake-username-2"))
			Expect(serviceKeys[1].Credentials).To(HaveKeyWithValue("password", "fake-password-2"))
			Expect(serviceKeys[1].Credentials).To(HaveKeyWithValue("host", "fake-host-2"))
			Expect(serviceKeys[1].Credentials).To(HaveKeyWithValue("port", float64(3306)))
			Expect(serviceKeys[1].Credentials).To(HaveKeyWithValue("database", "fake-db-name-2"))
			Expect(serviceKeys[1].Credentials).To(HaveKeyWithValue("uri", "mysql://fake-user-2:fake-password-2@fake-host-2:3306/fake-db-name-2"))
		})
	})

	AfterEach(func() {
		testServer.Close()
	})
})

var emptyServiceKeysResponse = testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`}

var serviceKeysResponse = testnet.TestResponse{Status: http.StatusOK, Body: `{
	"resources": [
		{
	      "metadata": {
	        "guid": "fake-service-key-guid-1",
	        "url": "/v2/service_keys/fake-guid-1",
	        "created_at": "2015-01-13T18:52:08+00:00",
	        "updated_at": null
	      },
	      "entity": {
	        "name": "fake-service-key-name-1",
	        "service_instance_guid":"fake-service-instance-guid-1",
	        "service_instance_url":"http://fake/service/instance/url/1",
	        "credentials": {
	          "username": "fake-username-1",
	          "password": "fake-password-1",
	          "host": "fake-host-1",
	          "port": 3306,
	          "database": "fake-db-name-1",
	          "uri": "mysql://fake-user-1:fake-password-1@fake-host-1:3306/fake-db-name-1"
	        }
	      }
	    },
	    {
	      "metadata": {
	        "guid": "fake-service-key-guid-2",
	        "url": "/v2/service_keys/fake-guid-2",
	        "created_at": "2015-01-13T18:52:08+00:00",
	        "updated_at": null
	      },
	      "entity": {
	        "name": "fake-service-key-name-2",
	        "service_instance_guid":"fake-service-instance-guid-2",
	        "service_instance_url":"http://fake/service/instance/url/1",
	        "credentials": {
	          "username": "fake-username-2",
	          "password": "fake-password-2",
	          "host": "fake-host-2",
	          "port": 3306,
	          "database": "fake-db-name-2",
	          "uri": "mysql://fake-user-2:fake-password-2@fake-host-2:3306/fake-db-name-2"
	        }
	      }
	    }
	]}`,
}
