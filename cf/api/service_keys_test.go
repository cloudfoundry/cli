package api_test

import (
	"net/http"
	"time"

	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testnet "code.cloudfoundry.org/cli/util/testhelpers/net"
	"github.com/onsi/gomega/ghttp"

	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"

	. "code.cloudfoundry.org/cli/cf/api"

	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Keys Repo", func() {
	var (
		ccServer   *ghttp.Server
		configRepo coreconfig.ReadWriter
		repo       ServiceKeyRepository
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		configRepo.SetAccessToken("BEARER my_access_token")

		ccServer = ghttp.NewServer()
		configRepo.SetAPIEndpoint(ccServer.URL())

		gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
		repo = NewCloudControllerServiceKeyRepository(configRepo, gateway)
	})

	AfterEach(func() {
		ccServer.Close()
	})

	Describe("CreateServiceKey", func() {
		It("tries to create the service key", func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v2/service_keys"),
					ghttp.RespondWith(http.StatusCreated, nil),
					ghttp.VerifyJSON(`{"service_instance_guid": "fake-instance-guid", "name": "fake-key-name"}`),
				),
			)

			err := repo.CreateServiceKey("fake-instance-guid", "fake-key-name", nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})

		Context("when the service key exists", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v2/service_keys"),
						ghttp.RespondWith(http.StatusBadRequest, `{"code":360001,"description":"The service key name is taken: exist-service-key"}`),
					),
				)
			})

			It("returns a ModelAlreadyExistsError", func() {
				err := repo.CreateServiceKey("fake-instance-guid", "exist-service-key", nil)
				Expect(err).To(BeAssignableToTypeOf(&errors.ModelAlreadyExistsError{}))
			})
		})

		Context("when the CLI user is not a space developer or admin", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v2/service_keys"),
						ghttp.RespondWith(http.StatusBadRequest, `{"code":10003,"description":"You are not authorized to perform the requested action"}`),
					),
				)
			})

			It("returns a NotAuthorizedError when CLI user is not the space developer or admin", func() {
				err := repo.CreateServiceKey("fake-instance-guid", "fake-service-key", nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("You are not authorized to perform the requested action"))
			})
		})

		Context("when there are parameters", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v2/service_keys"),
						ghttp.RespondWith(http.StatusCreated, nil),
						ghttp.VerifyJSON(`{"service_instance_guid":"fake-instance-guid","name":"fake-service-key","parameters": {"data": "hello"}}`),
					),
				)
			})

			It("includes any provided parameters", func() {
				paramsMap := make(map[string]interface{})
				paramsMap["data"] = "hello"

				err := repo.CreateServiceKey("fake-instance-guid", "fake-service-key", paramsMap)
				Expect(err).NotTo(HaveOccurred())
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns any serialization errors", func() {
				paramsMap := make(map[string]interface{})
				paramsMap["data"] = make(chan bool)

				err := repo.CreateServiceKey("instance-name", "plan-guid", paramsMap)
				Expect(err).To(MatchError("json: unsupported type: chan bool"))
			})
		})
	})

	Describe("ListServiceKeys", func() {
		Context("when no service key is found", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/service_instances/fake-instance-guid/service_keys"),
						ghttp.RespondWith(http.StatusOK, `{"resources": []}`),
					),
				)
			})

			It("returns an empty result", func() {
				serviceKeys, err := repo.ListServiceKeys("fake-instance-guid")
				Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
				Expect(err).NotTo(HaveOccurred())
				Expect(serviceKeys).To(HaveLen(0))
			})
		})

		Context("when service keys are found", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/service_instances/fake-instance-guid/service_keys"),
						ghttp.RespondWith(http.StatusOK, serviceKeysResponse),
					),
				)
			})

			It("returns the service keys", func() {
				serviceKeys, err := repo.ListServiceKeys("fake-instance-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(serviceKeys).To(HaveLen(2))

				Expect(serviceKeys[0].Fields.GUID).To(Equal("fake-service-key-guid-1"))
				Expect(serviceKeys[0].Fields.URL).To(Equal("/v2/service_keys/fake-guid-1"))
				Expect(serviceKeys[0].Fields.Name).To(Equal("fake-service-key-name-1"))
				Expect(serviceKeys[0].Fields.ServiceInstanceGUID).To(Equal("fake-service-instance-guid-1"))
				Expect(serviceKeys[0].Fields.ServiceInstanceURL).To(Equal("http://fake/service/instance/url/1"))

				Expect(serviceKeys[0].Credentials).To(HaveKeyWithValue("username", "fake-username-1"))
				Expect(serviceKeys[0].Credentials).To(HaveKeyWithValue("password", "fake-password-1"))
				Expect(serviceKeys[0].Credentials).To(HaveKeyWithValue("host", "fake-host-1"))
				Expect(serviceKeys[0].Credentials).To(HaveKeyWithValue("port", float64(3306)))
				Expect(serviceKeys[0].Credentials).To(HaveKeyWithValue("database", "fake-db-name-1"))
				Expect(serviceKeys[0].Credentials).To(HaveKeyWithValue("uri", "mysql://fake-user-1:fake-password-1@fake-host-1:3306/fake-db-name-1"))

				Expect(serviceKeys[1].Fields.GUID).To(Equal("fake-service-key-guid-2"))
				Expect(serviceKeys[1].Fields.URL).To(Equal("/v2/service_keys/fake-guid-2"))
				Expect(serviceKeys[1].Fields.Name).To(Equal("fake-service-key-name-2"))
				Expect(serviceKeys[1].Fields.ServiceInstanceGUID).To(Equal("fake-service-instance-guid-2"))
				Expect(serviceKeys[1].Fields.ServiceInstanceURL).To(Equal("http://fake/service/instance/url/1"))

				Expect(serviceKeys[1].Credentials).To(HaveKeyWithValue("username", "fake-username-2"))
				Expect(serviceKeys[1].Credentials).To(HaveKeyWithValue("password", "fake-password-2"))
				Expect(serviceKeys[1].Credentials).To(HaveKeyWithValue("host", "fake-host-2"))
				Expect(serviceKeys[1].Credentials).To(HaveKeyWithValue("port", float64(3306)))
				Expect(serviceKeys[1].Credentials).To(HaveKeyWithValue("database", "fake-db-name-2"))
				Expect(serviceKeys[1].Credentials).To(HaveKeyWithValue("uri", "mysql://fake-user-2:fake-password-2@fake-host-2:3306/fake-db-name-2"))
			})
		})

		Context("when the server responds with 403", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/service_instances/fake-instance-guid/service_keys"),
						ghttp.RespondWith(http.StatusForbidden, `{
						"code": 10003,
						"description": "You are not authorized to perform the requested action",
						"error_code": "CF-NotAuthorized"
					}`),
					),
				)
			})

			It("returns a NotAuthorizedError", func() {
				_, err := repo.ListServiceKeys("fake-instance-guid")
				Expect(err).To(BeAssignableToTypeOf(&errors.NotAuthorizedError{}))
			})
		})
	})

	Describe("GetServiceKey", func() {
		Context("when the service key is found", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/service_instances/fake-instance-guid/service_keys", "q=name:fake-service-key-name"),
						ghttp.RespondWith(http.StatusOK, serviceKeyDetailResponse),
					),
				)
			})

			It("returns service key detail", func() {
				serviceKey, err := repo.GetServiceKey("fake-instance-guid", "fake-service-key-name")
				Expect(err).NotTo(HaveOccurred())

				Expect(serviceKey.Fields.GUID).To(Equal("fake-service-key-guid"))
				Expect(serviceKey.Fields.URL).To(Equal("/v2/service_keys/fake-guid"))
				Expect(serviceKey.Fields.Name).To(Equal("fake-service-key-name"))
				Expect(serviceKey.Fields.ServiceInstanceGUID).To(Equal("fake-service-instance-guid"))
				Expect(serviceKey.Fields.ServiceInstanceURL).To(Equal("http://fake/service/instance/url"))

				Expect(serviceKey.Credentials).To(HaveKeyWithValue("username", "fake-username"))
				Expect(serviceKey.Credentials).To(HaveKeyWithValue("password", "fake-password"))
				Expect(serviceKey.Credentials).To(HaveKeyWithValue("host", "fake-host"))
				Expect(serviceKey.Credentials).To(HaveKeyWithValue("port", float64(3306)))
				Expect(serviceKey.Credentials).To(HaveKeyWithValue("database", "fake-db-name"))
				Expect(serviceKey.Credentials).To(HaveKeyWithValue("uri", "mysql://fake-user:fake-password@fake-host:3306/fake-db-name"))
			})
		})

		Context("when the service key is not found", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/service_instances/fake-instance-guid/service_keys", "q=name:non-exist-key-name"),
						ghttp.RespondWith(http.StatusOK, `{"resources": []}`),
					),
				)
			})

			It("returns an empty service key", func() {
				serviceKey, err := repo.GetServiceKey("fake-instance-guid", "non-exist-key-name")
				Expect(err).NotTo(HaveOccurred())
				Expect(serviceKey).To(Equal(models.ServiceKey{}))
			})
		})

		Context("when the server responds with 403", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/service_instances/fake-instance-guid/service_keys", "q=name:fake-service-key-name"),
						ghttp.RespondWith(http.StatusForbidden, `{
						"code": 10003,
						"description": "You are not authorized to perform the requested action",
						"error_code": "CF-NotAuthorized"
					}`),
					),
				)
			})

			It("returns a NotAuthorizedError", func() {
				_, err := repo.GetServiceKey("fake-instance-guid", "fake-service-key-name")
				Expect(err).To(BeAssignableToTypeOf(&errors.NotAuthorizedError{}))
			})
		})
	})

	Describe("DeleteServiceKey", func() {
		It("deletes service key successfully", func() {
			ccServer.AppendHandlers(
				ghttp.VerifyRequest("DELETE", "/v2/service_keys/fake-service-key-guid"),
				ghttp.RespondWith(http.StatusOK, nil),
			)

			err := repo.DeleteServiceKey("fake-service-key-guid")
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})
	})
})

var serviceKeysResponse = `{
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
]}`

var serviceKeyDetailResponse = `{
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
}`

var notAuthorizedResponse = testnet.TestResponse{Status: http.StatusForbidden, Body: `{
		"code": 10003,
		"description": "You are not authorized to perform the requested action",
		"error_code": "CF-NotAuthorized"
	}`,
}
