package api_test

import (
	"net/http"
	"time"

	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"

	. "code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("ServiceBindingsRepository", func() {
	var (
		server     *ghttp.Server
		configRepo coreconfig.ReadWriter
		repo       CloudControllerServiceBindingRepository
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		configRepo = testconfig.NewRepositoryWithDefaults()
		configRepo.SetAPIEndpoint(server.URL())
		gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
		repo = NewCloudControllerServiceBindingRepository(configRepo, gateway)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Create", func() {
		var requestBody string

		Context("when the service binding can be created", func() {
			JustBeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v2/service_bindings"),
						ghttp.VerifyJSON(requestBody),
						ghttp.RespondWith(http.StatusCreated, nil),
					),
				)
			})

			Context("no parameters passed", func() {
				BeforeEach(func() {
					requestBody = `{
						"app_guid":"my-app-guid",
						"service_instance_guid":"my-service-instance-guid"
					}`
				})

				It("creates the service binding", func() {
					err := repo.Create("my-service-instance-guid", "my-app-guid", nil)
					Expect(err).NotTo(HaveOccurred())

					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("when there are arbitrary parameters", func() {
				BeforeEach(func() {
					requestBody = `{
						"app_guid":"my-app-guid",
						"service_instance_guid":"my-service-instance-guid",
						"parameters": { "foo": "bar" }
					}`
				})

				It("send the parameters as part of the request body", func() {
					err := repo.Create(
						"my-service-instance-guid",
						"my-app-guid",
						map[string]interface{}{"foo": "bar"},
					)
					Expect(err).NotTo(HaveOccurred())

					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				Context("and there is a failure during serialization", func() {
					It("returns the serialization error", func() {
						paramsMap := make(map[string]interface{})
						paramsMap["data"] = make(chan bool)

						err := repo.Create("my-service-instance-guid", "my-app-guid", paramsMap)
						Expect(err).To(MatchError("json: unsupported type: chan bool"))
					})
				})
			})
		})

		Context("when an API error occurs", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v2/service_bindings"),
						ghttp.VerifyJSON(`{
							"app_guid":"my-app-guid",
							"service_instance_guid":"my-service-instance-guid"
						}`),
						ghttp.RespondWith(http.StatusBadRequest, `{
							"code":90003,
							"description":"The app space binding to service is taken: 7b959018-110a-4913-ac0a-d663e613cdea 346bf237-7eef-41a7-b892-68fb08068f09"
						}`),
					),
				)
			})

			It("returns an error", func() {
				err := repo.Create("my-service-instance-guid", "my-app-guid", nil)
				Expect(err).To(HaveOccurred())
				Expect(err.(errors.HTTPError).ErrorCode()).To(Equal("90003"))
			})
		})
	})

	Describe("Delete", func() {
		var serviceInstance models.ServiceInstance

		BeforeEach(func() {
			serviceInstance.GUID = "my-service-instance-guid"
		})

		Context("when binding does exist", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/v2/service_bindings/service-binding-2-guid"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)

				serviceInstance.ServiceBindings = []models.ServiceBindingFields{
					{
						URL:     "/v2/service_bindings/service-binding-1-guid",
						AppGUID: "app-1-guid",
					},
					{
						URL:     "/v2/service_bindings/service-binding-2-guid",
						AppGUID: "app-2-guid",
					},
				}
			})

			It("deletes the service binding with the given guid", func() {
				found, err := repo.Delete(serviceInstance, "app-2-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())

				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})
		})

		Context("when binding does not exist", func() {
			It("does not return an error", func() {
				found, err := repo.Delete(serviceInstance, "app-3-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeFalse())

				Expect(server.ReceivedRequests()).To(HaveLen(0))
			})
		})
	})

	Describe("ListAllForService", func() {
		Context("when binding does exist", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/service_instances/service-instance-guid/service_bindings"),
						ghttp.RespondWith(http.StatusOK, `{
						"total_results": 2,
						"total_pages": 2,
						"next_url": "/v2/service_instances/service-instance-guid/service_bindings?page=2",
						"resources": [
							{
								"metadata": {
									"guid": "service-binding-1-guid",
									"url": "/v2/service_bindings/service-binding-1-guid",
									"created_at": "2016-04-22T19:33:31Z",
									"updated_at": null
								},
								"entity": {
									"app_guid": "app-guid-1"
								}
							}
						]
					}`)),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/service_instances/service-instance-guid/service_bindings", "page=2"),
						ghttp.RespondWith(http.StatusOK, `{
						"total_results": 2,
						"total_pages": 2,
						"resources": [
							{
								"metadata": {
									"guid": "service-binding-2-guid",
									"url": "/v2/service_bindings/service-binding-2-guid",
									"created_at": "2016-04-22T19:33:31Z",
									"updated_at": null
								},
								"entity": {
									"app_guid": "app-guid-2"
								}
							}
						]
					}`)),
				)
			})

			It("returns the list of service instances", func() {
				bindings, err := repo.ListAllForService("service-instance-guid")
				Expect(err).NotTo(HaveOccurred())

				Expect(bindings).To(HaveLen(2))
				Expect(bindings[0].AppGUID).To(Equal("app-guid-1"))
				Expect(bindings[1].AppGUID).To(Equal("app-guid-2"))
			})
		})

		Context("when the service does not exist", func() {
			BeforeEach(func() {
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/service_instances/service-instance-guid/service_bindings"),
					ghttp.RespondWith(http.StatusGatewayTimeout, nil),
				))
			})

			It("returns an error", func() {
				_, err := repo.ListAllForService("service-instance-guid")
				Expect(err).To(HaveOccurred())

				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})
		})
	})
})
