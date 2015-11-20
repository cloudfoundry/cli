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
	"github.com/cloudfoundry/cli/cf/models"
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

	AfterEach(func() {
		testServer.Close()
	})

	Describe("CreateServiceKey", func() {
		It("makes the right request", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "POST",
				Path:     "/v2/service_keys",
				Matcher:  testnet.RequestBodyMatcher(`{"service_instance_guid": "fake-instance-guid", "name": "fake-key-name"}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			}))

			err := repo.CreateServiceKey("fake-instance-guid", "fake-key-name", nil)
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

			err := repo.CreateServiceKey("fake-instance-guid", "exist-service-key", nil)
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

			err := repo.CreateServiceKey("fake-instance-guid", "fake-service-key", nil)
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("You are not authorized to perform the requested action"))
		})

		Context("when there are parameters", func() {
			It("sends the parameters as part of the request body", func() {
				setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "POST",
					Path:     "/v2/service_keys",
					Matcher:  testnet.RequestBodyMatcher(`{"service_instance_guid":"fake-instance-guid","name":"fake-service-key","parameters": {"data": "hello"}}`),
					Response: testnet.TestResponse{Status: http.StatusCreated},
				}))

				paramsMap := make(map[string]interface{})
				paramsMap["data"] = "hello"

				err := repo.CreateServiceKey("fake-instance-guid", "fake-service-key", paramsMap)
				Expect(testHandler).To(HaveAllRequestsCalled())
				Expect(err).NotTo(HaveOccurred())
			})

			Context("and there is a failure during serialization", func() {
				It("returns the serialization error", func() {
					setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
						Method:   "POST",
						Path:     "/v2/service_keys",
						Matcher:  testnet.RequestBodyMatcher(`{"service_instance_guid":"fake-instance-guid","name":"fake-service-key","parameters": {"data": "hello"}}`),
						Response: testnet.TestResponse{Status: http.StatusCreated},
					}))

					paramsMap := make(map[string]interface{})
					paramsMap["data"] = make(chan bool)

					err := repo.CreateServiceKey("instance-name", "plan-guid", paramsMap)
					Expect(err).To(MatchError("json: unsupported type: chan bool"))
				})
			})
		})
	})

	Describe("ListServiceKeys", func() {
		It("returns empty result when no service key is found", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/service_instances/fake-instance-guid/service_keys",
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
				Path:     "/v2/service_instances/fake-instance-guid/service_keys",
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

		It("returns a NotAuthorizedError when server response is 403", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/service_instances/fake-instance-guid/service_keys",
				Response: notAuthorizedResponse,
			}))

			_, err := repo.ListServiceKeys("fake-instance-guid")
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).To(BeAssignableToTypeOf(&errors.NotAuthorizedError{}))
		})
	})

	Describe("GetServiceKey", func() {
		It("returns service key detail", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/service_instances/fake-instance-guid/service_keys?q=name:fake-service-key-name",
				Response: serviceKeyDetailResponse,
			}))

			serviceKey, err := repo.GetServiceKey("fake-instance-guid", "fake-service-key-name")
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).NotTo(HaveOccurred())

			Expect(serviceKey.Fields.Guid).To(Equal("fake-service-key-guid"))
			Expect(serviceKey.Fields.Url).To(Equal("/v2/service_keys/fake-guid"))
			Expect(serviceKey.Fields.Name).To(Equal("fake-service-key-name"))
			Expect(serviceKey.Fields.ServiceInstanceGuid).To(Equal("fake-service-instance-guid"))
			Expect(serviceKey.Fields.ServiceInstanceUrl).To(Equal("http://fake/service/instance/url"))

			Expect(serviceKey.Credentials).To(HaveKeyWithValue("username", "fake-username"))
			Expect(serviceKey.Credentials).To(HaveKeyWithValue("password", "fake-password"))
			Expect(serviceKey.Credentials).To(HaveKeyWithValue("host", "fake-host"))
			Expect(serviceKey.Credentials).To(HaveKeyWithValue("port", float64(3306)))
			Expect(serviceKey.Credentials).To(HaveKeyWithValue("database", "fake-db-name"))
			Expect(serviceKey.Credentials).To(HaveKeyWithValue("uri", "mysql://fake-user:fake-password@fake-host:3306/fake-db-name"))
		})

		It("returns empty result when the service key is not found", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/service_instances/fake-instance-guid/service_keys?q=name:non-exist-key-name",
				Response: emptyServiceKeysResponse,
			}))

			serviceKey, err := repo.GetServiceKey("fake-instance-guid", "non-exist-key-name")
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).NotTo(HaveOccurred())
			Expect(serviceKey).To(Equal(models.ServiceKey{}))
		})

		It("returns a NotAuthorizedError when server response is 403", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/service_instances/fake-instance-guid/service_keys?q=name:fake-service-key-name",
				Response: notAuthorizedResponse,
			}))

			_, err := repo.GetServiceKey("fake-instance-guid", "fake-service-key-name")
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).To(BeAssignableToTypeOf(&errors.NotAuthorizedError{}))
		})
	})

	Describe("DeleteServiceKey", func() {
		It("deletes service key successfully", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "DELETE",
				Path:   "/v2/service_keys/fake-service-key-guid",
			}))

			err := repo.DeleteServiceKey("fake-service-key-guid")
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).NotTo(HaveOccurred())
		})
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

var serviceKeyDetailResponse = testnet.TestResponse{Status: http.StatusOK, Body: `{
	"resources": [
		{
	      "metadata": {
	        "guid": "fake-service-key-guid",
	        "url": "/v2/service_keys/fake-guid",
	        "created_at": "2015-01-13T18:52:08+00:00",
	        "updated_at": null
	      },
	      "entity": {
	        "name": "fake-service-key-name",
	        "service_instance_guid":"fake-service-instance-guid",
	        "service_instance_url":"http://fake/service/instance/url",
	        "credentials": {
	          "username": "fake-username",
	          "password": "fake-password",
	          "host": "fake-host",
	          "port": 3306,
	          "database": "fake-db-name",
	          "uri": "mysql://fake-user:fake-password@fake-host:3306/fake-db-name"
	        }
	      }
		}]
	}`,
}

var notAuthorizedResponse = testnet.TestResponse{Status: http.StatusForbidden, Body: `{
		"code": 10003,
		"description": "You are not authorized to perform the requested action",
		"error_code": "CF-NotAuthorized"
	}`,
}
