package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testnet "code.cloudfoundry.org/cli/util/testhelpers/net"

	. "code.cloudfoundry.org/cli/cf/api"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Services Repo", func() {
	var (
		testServer  *httptest.Server
		testHandler *testnet.TestHandler
		configRepo  coreconfig.ReadWriter
		repo        ServiceRepository
	)

	setupTestServer := func(reqs ...testnet.TestRequest) {
		testServer, testHandler = testnet.NewServer(reqs)
		configRepo.SetAPIEndpoint(testServer.URL)
	}

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		configRepo.SetAccessToken("BEARER my_access_token")

		gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
		repo = NewCloudControllerServiceRepository(configRepo, gateway)
	})

	AfterEach(func() {
		if testServer != nil {
			testServer.Close()
		}
	})

	Describe("GetAllServiceOfferings", func() {
		BeforeEach(func() {
			setupTestServer(
				apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     "/v2/services",
					Response: firstOfferingsResponse,
				}),
				apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     "/v2/services",
					Response: multipleOfferingsResponse,
				}),
			)
		})

		It("gets all public service offerings", func() {
			offerings, err := repo.GetAllServiceOfferings()

			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).NotTo(HaveOccurred())
			Expect(len(offerings)).To(Equal(3))

			firstOffering := offerings[0]
			Expect(firstOffering.Label).To(Equal("first-Offering 1"))
			Expect(firstOffering.Version).To(Equal("1.0"))
			Expect(firstOffering.Description).To(Equal("first Offering 1 description"))
			Expect(firstOffering.Provider).To(Equal("Offering 1 provider"))
			Expect(firstOffering.GUID).To(Equal("first-offering-1-guid"))
		})
	})

	Describe("GetServiceOfferingsForSpace", func() {
		It("gets all service offerings in a given space", func() {
			setupTestServer(
				apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     "/v2/spaces/my-space-guid/services",
					Response: firstOfferingsForSpaceResponse,
				}),
				apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     "/v2/spaces/my-space-guid/services",
					Response: multipleOfferingsResponse,
				}))

			offerings, err := repo.GetServiceOfferingsForSpace("my-space-guid")

			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).NotTo(HaveOccurred())

			Expect(len(offerings)).To(Equal(3))

			firstOffering := offerings[0]
			Expect(firstOffering.Label).To(Equal("first-Offering 1"))
			Expect(firstOffering.Version).To(Equal("1.0"))
			Expect(firstOffering.Description).To(Equal("first Offering 1 description"))
			Expect(firstOffering.Provider).To(Equal("Offering 1 provider"))
			Expect(firstOffering.GUID).To(Equal("first-offering-1-guid"))
			Expect(len(firstOffering.Plans)).To(Equal(0))

			secondOffering := offerings[1]
			Expect(secondOffering.Label).To(Equal("Offering 1"))
			Expect(secondOffering.Version).To(Equal("1.0"))
			Expect(secondOffering.Description).To(Equal("Offering 1 description"))
			Expect(secondOffering.Provider).To(Equal("Offering 1 provider"))
			Expect(secondOffering.GUID).To(Equal("offering-1-guid"))
			Expect(len(secondOffering.Plans)).To(Equal(0))
		})
	})

	Describe("find by service broker", func() {
		BeforeEach(func() {
			body1 := `
{
   "total_results": 2,
   "total_pages": 2,
   "prev_url": null,
   "next_url": "/v2/services?q=service_broker_guid%3Amy-service-broker-guid&page=2",
   "resources": [
      {
         "metadata": {
            "guid": "my-service-guid"
         },
         "entity": {
            "label": "my-service",
            "provider": "androsterone-ensphere",
            "description": "Dummy addon that is cool",
            "version": "damageableness-preheat"
         }
      }
   ]
}`
			body2 := `
{
   "total_results": 1,
   "total_pages": 1,
   "next_url": null,
   "resources": [
      {
         "metadata": {
            "guid": "my-service-guid2"
         },
         "entity": {
            "label": "my-service2",
            "provider": "androsterone-ensphere",
            "description": "Dummy addon that is cooler",
            "version": "seraphine-lowdah"
         }
      }
   ]
}`

			setupTestServer(
				apifakes.NewCloudControllerTestRequest(
					testnet.TestRequest{
						Method:   "GET",
						Path:     "/v2/services?q=service_broker_guid%3Amy-service-broker-guid",
						Response: testnet.TestResponse{Status: http.StatusOK, Body: body1},
					}),
				apifakes.NewCloudControllerTestRequest(
					testnet.TestRequest{
						Method:   "GET",
						Path:     "/v2/services?q=service_broker_guid%3Amy-service-broker-guid",
						Response: testnet.TestResponse{Status: http.StatusOK, Body: body2},
					}),
			)
		})

		It("returns the service brokers services", func() {
			services, err := repo.ListServicesFromBroker("my-service-broker-guid")

			Expect(err).NotTo(HaveOccurred())
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(len(services)).To(Equal(2))

			Expect(services[0].GUID).To(Equal("my-service-guid"))
			Expect(services[1].GUID).To(Equal("my-service-guid2"))
		})
	})

	Describe("returning services for many brokers", func() {
		path1 := "/v2/services?q=service_broker_guid%20IN%20my-service-broker-guid,my-service-broker-guid2"
		body1 := `
{
   "total_results": 2,
   "total_pages": 2,
   "prev_url": null,
	 "next_url": "/v2/services?q=service_broker_guid%20IN%20my-service-broker-guid,my-service-broker-guid2&page=2",
   "resources": [
     {
         "metadata": {
            "guid": "my-service-guid"
         },
         "entity": {
            "label": "my-service",
            "provider": "androsterone-ensphere",
            "description": "Dummy addon that is cool",
            "version": "damageableness-preheat"
         }
			 }
   ]
}`
		path2 := "/v2/services?q=service_broker_guid%20IN%20my-service-broker-guid,my-service-broker-guid2&page=2"
		body2 := `
{
   "total_results": 2,
   "total_pages": 2,
   "prev_url": "/v2/services?q=service_broker_guid%20IN%20my-service-broker-guid,my-service-broker-guid2",
	 "next_url": null,
   "resources": [
      {
         "metadata": {
            "guid": "my-service-guid2"
         },
         "entity": {
            "label": "my-service2",
            "provider": "androsterone-ensphere",
            "description": "Dummy addon that is cool",
            "version": "damageableness-preheat"
         }
      }
   ]
}`
		BeforeEach(func() {
			setupTestServer(
				apifakes.NewCloudControllerTestRequest(
					testnet.TestRequest{
						Method:   "GET",
						Path:     path1,
						Response: testnet.TestResponse{Status: http.StatusOK, Body: body1},
					}),
				apifakes.NewCloudControllerTestRequest(
					testnet.TestRequest{
						Method:   "GET",
						Path:     path2,
						Response: testnet.TestResponse{Status: http.StatusOK, Body: body2},
					}),
			)
		})

		It("returns the service brokers services", func() {
			brokerGUIDs := []string{"my-service-broker-guid", "my-service-broker-guid2"}
			services, err := repo.ListServicesFromManyBrokers(brokerGUIDs)

			Expect(err).NotTo(HaveOccurred())
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(len(services)).To(Equal(2))

			Expect(services[0].GUID).To(Equal("my-service-guid"))
			Expect(services[1].GUID).To(Equal("my-service-guid2"))
		})
	})

	Describe("creating a service instance", func() {
		It("makes the right request", func() {
			setupTestServer(apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "POST",
				Path:     "/v2/service_instances?accepts_incomplete=true",
				Matcher:  testnet.RequestBodyMatcher(`{"name":"instance-name","service_plan_guid":"plan-guid","space_guid":"my-space-guid"}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			}))

			err := repo.CreateServiceInstance("instance-name", "plan-guid", nil, nil)
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there are parameters", func() {
			It("sends the parameters as part of the request body", func() {
				setupTestServer(apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "POST",
					Path:     "/v2/service_instances?accepts_incomplete=true",
					Matcher:  testnet.RequestBodyMatcher(`{"name":"instance-name","service_plan_guid":"plan-guid","space_guid":"my-space-guid","parameters": {"data": "hello"}}`),
					Response: testnet.TestResponse{Status: http.StatusCreated},
				}))

				paramsMap := make(map[string]interface{})
				paramsMap["data"] = "hello"

				err := repo.CreateServiceInstance("instance-name", "plan-guid", paramsMap, nil)
				Expect(testHandler).To(HaveAllRequestsCalled())
				Expect(err).NotTo(HaveOccurred())
			})

			Context("and there is a failure during serialization", func() {
				It("returns the serialization error", func() {
					paramsMap := make(map[string]interface{})
					paramsMap["data"] = make(chan bool)

					err := repo.CreateServiceInstance("instance-name", "plan-guid", paramsMap, nil)
					Expect(err).To(MatchError("json: unsupported type: chan bool"))
				})
			})
		})

		Context("when there are tags", func() {
			It("sends the tags as part of the request body", func() {
				setupTestServer(apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "POST",
					Path:     "/v2/service_instances?accepts_incomplete=true",
					Matcher:  testnet.RequestBodyMatcher(`{"name":"instance-name","service_plan_guid":"plan-guid","space_guid":"my-space-guid","tags": ["foo", "bar"]}`),
					Response: testnet.TestResponse{Status: http.StatusCreated},
				}))

				tags := []string{"foo", "bar"}

				err := repo.CreateServiceInstance("instance-name", "plan-guid", nil, tags)
				Expect(testHandler).To(HaveAllRequestsCalled())
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the name is taken but an identical service exists", func() {
			BeforeEach(func() {
				setupTestServer(
					apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
						Method:  "POST",
						Path:    "/v2/service_instances?accepts_incomplete=true",
						Matcher: testnet.RequestBodyMatcher(`{"name":"my-service","service_plan_guid":"plan-guid","space_guid":"my-space-guid"}`),
						Response: testnet.TestResponse{
							Status: http.StatusBadRequest,
							Body:   `{"code":60002,"description":"The service instance name is taken: my-service"}`,
						}}),
					findServiceInstanceReq,
					serviceOfferingReq)
			})

			It("returns a ModelAlreadyExistsError if the plan is the same", func() {
				err := repo.CreateServiceInstance("my-service", "plan-guid", nil, nil)
				Expect(testHandler).To(HaveAllRequestsCalled())
				Expect(err).To(BeAssignableToTypeOf(&errors.ModelAlreadyExistsError{}))
			})
		})

		Context("when the name is taken and no identical service instance exists", func() {
			BeforeEach(func() {
				setupTestServer(
					apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
						Method:  "POST",
						Path:    "/v2/service_instances?accepts_incomplete=true",
						Matcher: testnet.RequestBodyMatcher(`{"name":"my-service","service_plan_guid":"different-plan-guid","space_guid":"my-space-guid"}`),
						Response: testnet.TestResponse{
							Status: http.StatusBadRequest,
							Body:   `{"code":60002,"description":"The service instance name is taken: my-service"}`,
						}}),
					findServiceInstanceReq,
					serviceOfferingReq)
			})

			It("fails if the plan is different", func() {
				err := repo.CreateServiceInstance("my-service", "different-plan-guid", nil, nil)

				Expect(testHandler).To(HaveAllRequestsCalled())
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(errors.NewHTTPError(400, "", "")))
			})
		})
	})

	Describe("UpdateServiceInstance", func() {
		It("makes the right request", func() {
			setupTestServer(apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/service_instances/instance-guid?accepts_incomplete=true",
				Matcher:  testnet.RequestBodyMatcher(`{"service_plan_guid":"plan-guid", "tags": null}`),
				Response: testnet.TestResponse{Status: http.StatusOK},
			}))

			err := repo.UpdateServiceInstance("instance-guid", "plan-guid", nil, nil)
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).NotTo(HaveOccurred())
		})

		Context("When the instance or plan is not found", func() {
			It("fails", func() {
				setupTestServer(apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "PUT",
					Path:     "/v2/service_instances/instance-guid?accepts_incomplete=true",
					Matcher:  testnet.RequestBodyMatcher(`{"service_plan_guid":"plan-guid", "tags": null}`),
					Response: testnet.TestResponse{Status: http.StatusNotFound},
				}))

				err := repo.UpdateServiceInstance("instance-guid", "plan-guid", nil, nil)
				Expect(testHandler).To(HaveAllRequestsCalled())
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the user passes arbitrary params", func() {
			It("passes the parameters in the correct field for the request", func() {
				setupTestServer(apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "PUT",
					Path:     "/v2/service_instances/instance-guid?accepts_incomplete=true",
					Matcher:  testnet.RequestBodyMatcher(`{"parameters": {"foo": "bar"}, "tags": null}`),
					Response: testnet.TestResponse{Status: http.StatusOK},
				}))

				paramsMap := map[string]interface{}{"foo": "bar"}

				err := repo.UpdateServiceInstance("instance-guid", "", paramsMap, nil)
				Expect(testHandler).To(HaveAllRequestsCalled())
				Expect(err).NotTo(HaveOccurred())
			})

			Context("and there is a failure during serialization", func() {
				It("returns the serialization error", func() {
					paramsMap := make(map[string]interface{})
					paramsMap["data"] = make(chan bool)

					err := repo.UpdateServiceInstance("instance-guid", "", paramsMap, nil)
					Expect(err).To(MatchError("json: unsupported type: chan bool"))
				})
			})
		})

		Context("when there are tags", func() {
			It("sends the tags as part of the request body", func() {
				setupTestServer(apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "PUT",
					Path:     "/v2/service_instances/instance-guid?accepts_incomplete=true",
					Matcher:  testnet.RequestBodyMatcher(`{"tags": ["foo", "bar"]}`),
					Response: testnet.TestResponse{Status: http.StatusOK},
				}))

				tags := []string{"foo", "bar"}

				err := repo.UpdateServiceInstance("instance-guid", "", nil, tags)
				Expect(testHandler).To(HaveAllRequestsCalled())
				Expect(err).NotTo(HaveOccurred())
			})

			It("sends empty tags", func() {
				setupTestServer(apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "PUT",
					Path:     "/v2/service_instances/instance-guid?accepts_incomplete=true",
					Matcher:  testnet.RequestBodyMatcher(`{"tags": []}`),
					Response: testnet.TestResponse{Status: http.StatusOK},
				}))

				tags := []string{}

				err := repo.UpdateServiceInstance("instance-guid", "", nil, tags)
				Expect(testHandler).To(HaveAllRequestsCalled())
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("finding service instances by name", func() {
		It("returns the service instance", func() {
			setupTestServer(findServiceInstanceReq, serviceOfferingReq)

			instance, err := repo.FindInstanceByName("my-service")

			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).NotTo(HaveOccurred())

			Expect(instance.Name).To(Equal("my-service"))
			Expect(instance.GUID).To(Equal("my-service-instance-guid"))
			Expect(instance.DashboardURL).To(Equal("my-dashboard-url"))
			Expect(instance.ServiceOffering.Label).To(Equal("mysql"))
			Expect(instance.ServiceOffering.DocumentationURL).To(Equal("http://info.example.com"))
			Expect(instance.ServiceOffering.Description).To(Equal("MySQL database"))
			Expect(instance.ServiceOffering.Requires).To(ContainElement("route_forwarding"))
			Expect(instance.ServicePlan.Name).To(Equal("plan-name"))
			Expect(len(instance.ServiceBindings)).To(Equal(2))

			binding := instance.ServiceBindings[0]
			Expect(binding.URL).To(Equal("/v2/service_bindings/service-binding-1-guid"))
			Expect(binding.GUID).To(Equal("service-binding-1-guid"))
			Expect(binding.AppGUID).To(Equal("app-1-guid"))
		})

		It("returns user provided services", func() {
			setupTestServer(apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/spaces/my-space-guid/service_instances?return_user_provided_service_instances=true&q=name%3Amy-service",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `
				{
					"resources": [
						{
						  "metadata": {
							"guid": "my-service-instance-guid"
						  },
						  "entity": {
							"name": "my-service",
							"service_bindings": [
							  {
								"metadata": {
								  "guid": "service-binding-1-guid",
								  "url": "/v2/service_bindings/service-binding-1-guid"
								},
								"entity": {
								  "app_guid": "app-1-guid"
								}
							  },
							  {
								"metadata": {
								  "guid": "service-binding-2-guid",
								  "url": "/v2/service_bindings/service-binding-2-guid"
								},
								"entity": {
								  "app_guid": "app-2-guid"
								}
							  }
							],
							"service_plan_guid": null
						  }
						}
					]
				}`}}))

			instance, err := repo.FindInstanceByName("my-service")

			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).NotTo(HaveOccurred())

			Expect(instance.Name).To(Equal("my-service"))
			Expect(instance.GUID).To(Equal("my-service-instance-guid"))
			Expect(instance.ServiceOffering.Label).To(Equal(""))
			Expect(instance.ServicePlan.Name).To(Equal(""))
			Expect(len(instance.ServiceBindings)).To(Equal(2))

			binding := instance.ServiceBindings[0]
			Expect(binding.URL).To(Equal("/v2/service_bindings/service-binding-1-guid"))
			Expect(binding.GUID).To(Equal("service-binding-1-guid"))
			Expect(binding.AppGUID).To(Equal("app-1-guid"))
		})

		It("returns a failure response when the instance doesn't exist", func() {
			setupTestServer(apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/spaces/my-space-guid/service_instances?return_user_provided_service_instances=true&q=name%3Amy-service",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{ "resources": [] }`},
			}))

			_, err := repo.FindInstanceByName("my-service")

			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).To(BeAssignableToTypeOf(&errors.ModelNotFoundError{}))
		})

		It("should not fail to parse when extra is null", func() {
			setupTestServer(findServiceInstanceReq, serviceOfferingNullExtraReq)

			_, err := repo.FindInstanceByName("my-service")

			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("DeleteService", func() {
		It("deletes the service when no apps and keys are bound", func() {
			setupTestServer(apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "DELETE",
				Path:     "/v2/service_instances/my-service-instance-guid?accepts_incomplete=true&async=true",
				Response: testnet.TestResponse{Status: http.StatusOK},
			}))

			serviceInstance := models.ServiceInstance{}
			serviceInstance.GUID = "my-service-instance-guid"

			err := repo.DeleteService(serviceInstance)
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).NotTo(HaveOccurred())
		})

		It("doesn't delete the service when apps are bound", func() {
			setupTestServer()

			serviceInstance := models.ServiceInstance{}
			serviceInstance.GUID = "my-service-instance-guid"
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

			err := repo.DeleteService(serviceInstance)
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&errors.ServiceAssociationError{}))
		})

		It("doesn't delete the service when keys are bound", func() {
			setupTestServer()

			serviceInstance := models.ServiceInstance{}
			serviceInstance.GUID = "my-service-instance-guid"
			serviceInstance.ServiceKeys = []models.ServiceKeyFields{
				{
					Name: "fake-service-key-1",
					URL:  "/v2/service_keys/service-key-1-guid",
					GUID: "service-key-1-guid",
				},
				{
					Name: "fake-service-key-2",
					URL:  "/v2/service_keys/service-key-2-guid",
					GUID: "service-key-2-guid",
				},
			}

			err := repo.DeleteService(serviceInstance)
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&errors.ServiceAssociationError{}))
		})
	})

	Describe("RenameService", func() {
		Context("when the service is not user provided", func() {

			BeforeEach(func() {
				setupTestServer(apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "PUT",
					Path:     "/v2/service_instances/my-service-instance-guid?accepts_incomplete=true",
					Matcher:  testnet.RequestBodyMatcher(`{"name":"new-name"}`),
					Response: testnet.TestResponse{Status: http.StatusCreated},
				}))
			})

			It("renames the service", func() {
				serviceInstance := models.ServiceInstance{}
				serviceInstance.GUID = "my-service-instance-guid"
				serviceInstance.ServicePlan = models.ServicePlanFields{
					GUID: "some-plan-guid",
				}

				err := repo.RenameService(serviceInstance, "new-name")
				Expect(testHandler).To(HaveAllRequestsCalled())
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the service is user provided", func() {
			BeforeEach(func() {
				setupTestServer(apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "PUT",
					Path:     "/v2/user_provided_service_instances/my-service-instance-guid",
					Matcher:  testnet.RequestBodyMatcher(`{"name":"new-name"}`),
					Response: testnet.TestResponse{Status: http.StatusCreated},
				}))
			})

			It("renames the service", func() {
				serviceInstance := models.ServiceInstance{}
				serviceInstance.GUID = "my-service-instance-guid"

				err := repo.RenameService(serviceInstance, "new-name")
				Expect(testHandler).To(HaveAllRequestsCalled())
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("FindServiceOfferingByLabelAndProvider", func() {
		Context("when the service offering can be found", func() {
			BeforeEach(func() {
				setupTestServer(testnet.TestRequest{
					Method: "GET",
					Path:   fmt.Sprintf("/v2/services?q=%s", url.QueryEscape("label:offering-1;provider:provider-1")),
					Response: testnet.TestResponse{
						Status: 200,
						Body: `
						{
							"next_url": null,
							"resources": [
								{
								  "metadata": {
									"guid": "offering-1-guid"
								  },
								  "entity": {
									"label": "offering-1",
									"provider": "provider-1",
									"description": "offering 1 description",
									"version" : "1.0",
									"service_plans": []
								  }
								}
							]
						}`}})
			})

			It("finds service offerings by label and provider", func() {
				offering, err := repo.FindServiceOfferingByLabelAndProvider("offering-1", "provider-1")
				Expect(offering.GUID).To(Equal("offering-1-guid"))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the service offering cannot be found", func() {
			BeforeEach(func() {
				setupTestServer(testnet.TestRequest{
					Method: "GET",
					Path:   fmt.Sprintf("/v2/services?q=%s", url.QueryEscape("label:offering-1;provider:provider-1")),
					Response: testnet.TestResponse{
						Status: 200,
						Body: `
						{
							"next_url": null,
							"resources": []
						}`,
					},
				})
			})
			It("returns a ModelNotFoundError", func() {
				offering, err := repo.FindServiceOfferingByLabelAndProvider("offering-1", "provider-1")

				Expect(err).To(BeAssignableToTypeOf(&errors.ModelNotFoundError{}))
				Expect(offering.GUID).To(Equal(""))
			})
		})

		It("handles api errors when finding service offerings", func() {
			setupTestServer(testnet.TestRequest{
				Method: "GET",
				Path:   fmt.Sprintf("/v2/services?q=%s", url.QueryEscape("label:offering-1;provider:provider-1")),
				Response: testnet.TestResponse{
					Status: 400,
					Body: `
					{
            			"code": 10005,
            			"description": "The query parameter is invalid"
					}`}})

			_, err := repo.FindServiceOfferingByLabelAndProvider("offering-1", "provider-1")
			Expect(err).To(HaveOccurred())
			Expect(err.(errors.HTTPError).ErrorCode()).To(Equal("10005"))
		})
	})

	Describe("FindServiceOfferingsByLabel", func() {
		Context("when the service offering can be found", func() {
			BeforeEach(func() {
				setupTestServer(testnet.TestRequest{
					Method: "GET",
					Path:   fmt.Sprintf("/v2/services?q=%s", url.QueryEscape("label:offering-1")),
					Response: testnet.TestResponse{
						Status: 200,
						Body: `
						{
							"next_url": null,
							"resources": [
								{
								  "metadata": {
									"guid": "offering-1-guid"
								  },
								  "entity": {
									"label": "offering-1",
									"provider": "provider-1",
									"description": "offering 1 description",
									"version" : "1.0",
									"service_plans": [],
                  "service_broker_guid": "broker-1-guid"
								  }
								}
							]
						}`}})
			})

			It("finds service offerings by label", func() {
				offerings, err := repo.FindServiceOfferingsByLabel("offering-1")
				Expect(offerings[0].GUID).To(Equal("offering-1-guid"))
				Expect(offerings[0].Label).To(Equal("offering-1"))
				Expect(offerings[0].Provider).To(Equal("provider-1"))
				Expect(offerings[0].Description).To(Equal("offering 1 description"))
				Expect(offerings[0].Version).To(Equal("1.0"))
				Expect(offerings[0].BrokerGUID).To(Equal("broker-1-guid"))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the service offering cannot be found", func() {
			BeforeEach(func() {
				setupTestServer(testnet.TestRequest{
					Method: "GET",
					Path:   fmt.Sprintf("/v2/services?q=%s", url.QueryEscape("label:offering-1")),
					Response: testnet.TestResponse{
						Status: 200,
						Body: `
						{
							"next_url": null,
							"resources": []
						}`,
					},
				})
			})

			It("returns a ModelNotFoundError", func() {
				offerings, err := repo.FindServiceOfferingsByLabel("offering-1")

				Expect(err).To(BeAssignableToTypeOf(&errors.ModelNotFoundError{}))
				Expect(offerings).To(Equal(models.ServiceOfferings{}))
			})
		})

		It("handles api errors when finding service offerings", func() {
			setupTestServer(testnet.TestRequest{
				Method: "GET",
				Path:   fmt.Sprintf("/v2/services?q=%s", url.QueryEscape("label:offering-1")),
				Response: testnet.TestResponse{
					Status: 400,
					Body: `
					{
            			"code": 10005,
            			"description": "The query parameter is invalid"
					}`}})

			_, err := repo.FindServiceOfferingsByLabel("offering-1")
			Expect(err).To(HaveOccurred())
			Expect(err.(errors.HTTPError).ErrorCode()).To(Equal("10005"))
		})
	})

	Describe("GetServiceOfferingByGUID", func() {
		Context("when the service offering can be found", func() {
			BeforeEach(func() {
				setupTestServer(testnet.TestRequest{
					Method: "GET",
					Path:   fmt.Sprintf("/v2/services/offering-1-guid"),
					Response: testnet.TestResponse{
						Status: 200,
						Body: `
								{
								  "metadata": {
									"guid": "offering-1-guid"
								  },
								  "entity": {
									"label": "offering-1",
									"provider": "provider-1",
									"description": "offering 1 description",
									"version" : "1.0",
									"service_plans": [],
                  "service_broker_guid": "broker-1-guid"
								  }
								}`}})
			})

			It("finds service offerings by guid", func() {
				offering, err := repo.GetServiceOfferingByGUID("offering-1-guid")
				Expect(offering.GUID).To(Equal("offering-1-guid"))
				Expect(offering.Label).To(Equal("offering-1"))
				Expect(offering.Provider).To(Equal("provider-1"))
				Expect(offering.Description).To(Equal("offering 1 description"))
				Expect(offering.Version).To(Equal("1.0"))
				Expect(offering.BrokerGUID).To(Equal("broker-1-guid"))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the service offering cannot be found", func() {
			BeforeEach(func() {
				setupTestServer(testnet.TestRequest{
					Method: "GET",
					Path:   fmt.Sprintf("/v2/services/offering-1-guid"),
					Response: testnet.TestResponse{
						Status: 404,
						Body: `
						{
							"code": 120003,
							"description": "The service could not be found: offering-1-guid",
              "error_code": "CF-ServiceNotFound"
						}`,
					},
				})
			})

			It("returns a ModelNotFoundError", func() {
				offering, err := repo.GetServiceOfferingByGUID("offering-1-guid")

				Expect(err).To(BeAssignableToTypeOf(&errors.HTTPNotFoundError{}))
				Expect(offering.GUID).To(Equal(""))
			})
		})
	})

	Describe("PurgeServiceOffering", func() {
		It("purges service offerings", func() {
			setupTestServer(testnet.TestRequest{
				Method: "DELETE",
				Path:   "/v2/services/the-service-guid?purge=true",
				Response: testnet.TestResponse{
					Status: 204,
				}})

			offering := models.ServiceOffering{ServiceOfferingFields: models.ServiceOfferingFields{
				Label:       "the-offering",
				GUID:        "the-service-guid",
				Description: "some service description",
			}}
			offering.GUID = "the-service-guid"

			err := repo.PurgeServiceOffering(offering)
			Expect(err).NotTo(HaveOccurred())
			Expect(testHandler).To(HaveAllRequestsCalled())
		})
	})

	Describe("PurgeServiceInstance", func() {
		It("purges service instances", func() {
			setupTestServer(testnet.TestRequest{
				Method: "DELETE",
				Path:   "/v2/service_instances/instance-guid?purge=true",
				Response: testnet.TestResponse{
					Status: 204,
				}})

			instance := models.ServiceInstance{ServiceInstanceFields: models.ServiceInstanceFields{
				Name: "schrodinger",
				GUID: "instance-guid",
			}}

			err := repo.PurgeServiceInstance(instance)
			Expect(err).NotTo(HaveOccurred())
			Expect(testHandler).To(HaveAllRequestsCalled())
		})
	})

	Describe("getting the count of service instances for a service plan", func() {
		var planGUID = "abc123"

		It("returns the number of service instances", func() {
			setupTestServer(apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   fmt.Sprintf("/v2/service_plans/%s/service_instances?results-per-page=1", planGUID),
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `
                    {
                      "total_results": 9,
                      "total_pages": 9,
                      "prev_url": null,
                      "next_url": "/v2/service_plans/abc123/service_instances?page=2&results-per-page=1",
                      "resources": [
                        {
                          "metadata": {
                            "guid": "def456",
                            "url": "/v2/service_instances/def456",
                            "created_at": "2013-06-06T02:42:55+00:00",
                            "updated_at": null
                          },
                          "entity": {
                            "name": "pet-db",
                            "credentials": { "name": "the_name" },
                            "service_plan_guid": "abc123",
                            "space_guid": "ghi789",
                            "dashboard_url": "https://example.com/dashboard",
                            "type": "managed_service_instance",
                            "space_url": "/v2/spaces/ghi789",
                            "service_plan_url": "/v2/service_plans/abc123",
                            "service_bindings_url": "/v2/service_instances/def456/service_bindings"
                          }
                        }
                      ]
                    }
                `},
			}))

			count, err := repo.GetServiceInstanceCountForServicePlan(planGUID)
			Expect(count).To(Equal(9))
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the API error when one occurs", func() {
			setupTestServer(apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     fmt.Sprintf("/v2/service_plans/%s/service_instances?results-per-page=1", planGUID),
				Response: testnet.TestResponse{Status: http.StatusInternalServerError},
			}))

			_, err := repo.GetServiceInstanceCountForServicePlan(planGUID)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("finding a service plan", func() {
		var planDescription resources.ServicePlanDescription

		Context("when the service is a v1 service", func() {
			BeforeEach(func() {
				planDescription = resources.ServicePlanDescription{
					ServiceLabel:    "v1-elephantsql",
					ServicePlanName: "v1-panda",
					ServiceProvider: "v1-elephantsql",
				}

				setupTestServer(apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "GET",
					Path:   fmt.Sprintf("/v2/services?inline-relations-depth=1&q=%s", url.QueryEscape("label:v1-elephantsql;provider:v1-elephantsql")),
					Response: testnet.TestResponse{Status: http.StatusOK, Body: `
                        {
                          "resources": [
                            {
                              "metadata": {
                                "guid": "offering-1-guid"
                              },
                              "entity": {
                                "label": "v1-elephantsql",
                                "provider": "v1-elephantsql",
                                "description": "Offering 1 description",
                                "version" : "1.0",
                                "service_plans": [
                                    {
                                        "metadata": {"guid": "offering-1-plan-1-guid"},
                                        "entity": {"name": "not-the-plan-youre-looking-for"}
                                    },
                                    {
                                        "metadata": {"guid": "offering-1-plan-2-guid"},
                                        "entity": {"name": "v1-panda"}
                                    }
                                ]
                              }
                            }
                          ]
                        }`}}))
			})

			It("returns the plan guid for a v1 plan", func() {
				guid, err := repo.FindServicePlanByDescription(planDescription)

				Expect(guid).To(Equal("offering-1-plan-2-guid"))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the service is a v2 service", func() {
			BeforeEach(func() {
				planDescription = resources.ServicePlanDescription{
					ServiceLabel:    "v2-elephantsql",
					ServicePlanName: "v2-panda",
				}

				setupTestServer(apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "GET",
					Path:   fmt.Sprintf("/v2/services?inline-relations-depth=1&q=%s", url.QueryEscape("label:v2-elephantsql;provider:")),
					Response: testnet.TestResponse{Status: http.StatusOK, Body: `
                        {
                          "resources": [
                            {
                              "metadata": {
                                "guid": "offering-1-guid"
                              },
                              "entity": {
                                "label": "v2-elephantsql",
                                "provider": null,
                                "description": "Offering 1 description",
                                "version" : "1.0",
                                "service_plans": [
                                    {
                                        "metadata": {"guid": "offering-1-plan-1-guid"},
                                        "entity": {"name": "not-the-plan-youre-looking-for"}
                                    },
                                    {
                                        "metadata": {"guid": "offering-1-plan-2-guid"},
                                        "entity": {"name": "v2-panda"}
                                    }
                                ]
                              }
                            }
                          ]
                        }`}}))
			})

			It("returns the plan guid for a v2 plan", func() {
				guid, err := repo.FindServicePlanByDescription(planDescription)
				Expect(err).NotTo(HaveOccurred())
				Expect(guid).To(Equal("offering-1-plan-2-guid"))
			})
		})

		Context("when no service matches the description", func() {
			BeforeEach(func() {
				planDescription = resources.ServicePlanDescription{
					ServiceLabel:    "v2-service-label",
					ServicePlanName: "v2-plan-name",
				}

				setupTestServer(apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     fmt.Sprintf("/v2/services?inline-relations-depth=1&q=%s", url.QueryEscape("label:v2-service-label;provider:")),
					Response: testnet.TestResponse{Status: http.StatusOK, Body: `{ "resources": [] }`},
				}))
			})

			It("returns an error", func() {
				_, err := repo.FindServicePlanByDescription(planDescription)
				Expect(err).To(BeAssignableToTypeOf(&errors.ModelNotFoundError{}))
				Expect(err.Error()).To(ContainSubstring("Plan"))
				Expect(err.Error()).To(ContainSubstring("v2-service-label v2-plan-name"))
			})
		})

		Context("when the described service has no matching plan", func() {
			BeforeEach(func() {
				planDescription = resources.ServicePlanDescription{
					ServiceLabel:    "v2-service-label",
					ServicePlanName: "v2-plan-name",
				}

				setupTestServer(apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "GET",
					Path:   fmt.Sprintf("/v2/services?inline-relations-depth=1&q=%s", url.QueryEscape("label:v2-service-label;provider:")),
					Response: testnet.TestResponse{Status: http.StatusOK, Body: `
                        {
                          "resources": [
                            {
                              "metadata": {
                                "guid": "offering-1-guid"
                              },
                              "entity": {
                                "label": "v2-elephantsql",
                                "provider": null,
                                "description": "Offering 1 description",
                                "version" : "1.0",
                                "service_plans": [
                                  {
                                    "metadata": {"guid": "offering-1-plan-1-guid"},
                                    "entity": {"name": "not-the-plan-youre-looking-for"}
                                  },
                                  {
                                    "metadata": {"guid": "offering-1-plan-2-guid"},
                                    "entity": {"name": "also-not-the-plan-youre-looking-for"}
                                  }
                                ]
                              }
                            }
                          ]
                        }`}}))
			})

			It("returns a ModelNotFoundError", func() {
				_, err := repo.FindServicePlanByDescription(planDescription)

				Expect(err).To(BeAssignableToTypeOf(&errors.ModelNotFoundError{}))
				Expect(err.Error()).To(ContainSubstring("Plan"))
				Expect(err.Error()).To(ContainSubstring("v2-service-label v2-plan-name"))
			})
		})

		Context("when we get an HTTP error", func() {
			BeforeEach(func() {
				planDescription = resources.ServicePlanDescription{
					ServiceLabel:    "v2-service-label",
					ServicePlanName: "v2-plan-name",
				}

				setupTestServer(apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "GET",
					Path:   fmt.Sprintf("/v2/services?inline-relations-depth=1&q=%s", url.QueryEscape("label:v2-service-label;provider:")),
					Response: testnet.TestResponse{
						Status: http.StatusInternalServerError,
					}}))
			})

			It("returns an error", func() {
				_, err := repo.FindServicePlanByDescription(planDescription)

				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(errors.NewHTTPError(500, "", "")))
			})
		})
	})

	Describe("migrating service plans", func() {
		It("makes a request to CC to migrate the instances from v1 to v2", func() {
			setupTestServer(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/service_plans/v1-guid/service_instances",
				Matcher:  testnet.RequestBodyMatcher(`{"service_plan_guid":"v2-guid"}`),
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"changed_count":3}`},
			})

			changedCount, err := repo.MigrateServicePlanFromV1ToV2("v1-guid", "v2-guid")
			Expect(err).NotTo(HaveOccurred())
			Expect(changedCount).To(Equal(3))
		})

		It("returns an error when migrating fails", func() {
			setupTestServer(apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/service_plans/v1-guid/service_instances",
				Matcher:  testnet.RequestBodyMatcher(`{"service_plan_guid":"v2-guid"}`),
				Response: testnet.TestResponse{Status: http.StatusInternalServerError},
			}))

			_, err := repo.MigrateServicePlanFromV1ToV2("v1-guid", "v2-guid")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("FindServiceOfferingsForSpaceByLabel", func() {
		It("finds service offerings within a space by label", func() {
			setupTestServer(
				testnet.TestRequest{
					Method: "GET",
					Path:   fmt.Sprintf("/v2/spaces/my-space-guid/services?q=%s", url.QueryEscape("label:offering-1")),
					Response: testnet.TestResponse{
						Status: 200,
						Body: `
						{
							"next_url": "/v2/spaces/my-space-guid/services?q=label%3Aoffering-1&page=2",
							"resources": [
								{
									"metadata": {
										"guid": "offering-1-guid"
									},
									"entity": {
										"label": "offering-1",
										"provider": "provider-1",
										"description": "offering 1 description",
										"version" : "1.0"
									  }
								}
							]
						}`}},
				testnet.TestRequest{
					Method: "GET",
					Path:   fmt.Sprintf("/v2/spaces/my-space-guid/services?q=%s", url.QueryEscape("label:offering-1")),
					Response: testnet.TestResponse{
						Status: 200,
						Body: `
						{
							"next_url": null,
							"resources": [
								{
									"metadata": {
										"guid": "offering-2-guid"
									},
									"entity": {
										"label": "offering-2",
										"provider": "provider-2",
										"description": "offering 2 description",
										"version" : "1.0"
									}
								}
							]
						}`}})

			offerings, err := repo.FindServiceOfferingsForSpaceByLabel("my-space-guid", "offering-1")
			Expect(err).ToNot(HaveOccurred())
			Expect(offerings).To(HaveLen(2))
			Expect(offerings[0].GUID).To(Equal("offering-1-guid"))
		})

		It("returns an error if the offering cannot be found", func() {
			setupTestServer(testnet.TestRequest{
				Method: "GET",
				Path:   fmt.Sprintf("/v2/spaces/my-space-guid/services?q=%s", url.QueryEscape("label:offering-1")),
				Response: testnet.TestResponse{
					Status: http.StatusOK,
					Body: `{
						"next_url": null,
						"resources": []
					}`,
				},
			})

			offerings, err := repo.FindServiceOfferingsForSpaceByLabel("my-space-guid", "offering-1")
			Expect(err).To(BeAssignableToTypeOf(&errors.ModelNotFoundError{}))
			Expect(offerings).To(HaveLen(0))
		})

		It("handles api errors when finding service offerings", func() {
			setupTestServer(testnet.TestRequest{
				Method: "GET",
				Path:   fmt.Sprintf("/v2/spaces/my-space-guid/services?q=%s", url.QueryEscape("label:offering-1")),
				Response: testnet.TestResponse{
					Status: http.StatusBadRequest,
					Body: `{
						"code": 9001,
						"description": "Something Happened"
					}`,
				},
			})

			_, err := repo.FindServiceOfferingsForSpaceByLabel("my-space-guid", "offering-1")
			Expect(err).To(BeAssignableToTypeOf(errors.NewHTTPError(400, "", "")))
		})

		Describe("when api returns query by label is invalid", func() {
			It("makes a backwards-compatible request", func() {
				failedRequestByQueryLabel := testnet.TestRequest{
					Method: "GET",
					Path:   fmt.Sprintf("/v2/spaces/my-space-guid/services?q=%s", url.QueryEscape("label:my-service-offering")),
					Response: testnet.TestResponse{
						Status: http.StatusBadRequest,
						Body:   `{"code": 10005,"description": "The query parameter is invalid"}`,
					},
				}

				firstPaginatedRequest := testnet.TestRequest{
					Method: "GET",
					Path:   fmt.Sprintf("/v2/spaces/my-space-guid/services"),
					Response: testnet.TestResponse{
						Status: http.StatusOK,
						Body: `{
							"next_url": "/v2/spaces/my-space-guid/services?page=2",
							"resources": [
								{
								  "metadata": {
									"guid": "my-service-offering-guid"
								  },
								  "entity": {
									"label": "my-service-offering",
									"provider": "some-other-provider",
									"description": "a description that does not match your provider",
									"version" : "1.0"
								  }
								}
							]
						}`,
					},
				}

				secondPaginatedRequest := testnet.TestRequest{
					Method: "GET",
					Path:   fmt.Sprintf("/v2/spaces/my-space-guid/services"),
					Response: testnet.TestResponse{
						Status: http.StatusOK,
						Body: `{"next_url": null,
									"resources": [
										{
										  "metadata": {
											"guid": "my-service-offering-guid"
										  },
										  "entity": {
											"label": "my-service-offering",
											"provider": "my-provider",
											"description": "offering 1 description",
											"version" : "1.0"
										  }
										}
									]}`,
					},
				}

				setupTestServer(failedRequestByQueryLabel, firstPaginatedRequest, secondPaginatedRequest)

				serviceOfferings, err := repo.FindServiceOfferingsForSpaceByLabel("my-space-guid", "my-service-offering")
				Expect(err).NotTo(HaveOccurred())
				Expect(len(serviceOfferings)).To(Equal(2))
			})
		})
	})
})

var firstOfferingsResponse = testnet.TestResponse{Status: http.StatusOK, Body: `
{
	"next_url": "/v2/services?page=2",
	"resources": [
	{
		"metadata": {
			"guid": "first-offering-1-guid"
		},
		"entity": {
			"label": "first-Offering 1",
			"provider": "Offering 1 provider",
			"description": "first Offering 1 description",
			"version" : "1.0"
		}
	}
  ]}`,
}

var firstOfferingsForSpaceResponse = testnet.TestResponse{Status: http.StatusOK, Body: `
{
	"next_url": "/v2/spaces/my-space-guid/services?inline-relations-depth=1&page=2",
	"resources": [
		{
			"metadata": {
				"guid": "first-offering-1-guid"
			},
			"entity": {
				"label": "first-Offering 1",
				"provider": "Offering 1 provider",
				"description": "first Offering 1 description",
				"version" : "1.0"
			}
	    }
    ]}`,
}

var multipleOfferingsResponse = testnet.TestResponse{Status: http.StatusOK, Body: `
{
	"resources": [
		{
	    	"metadata": {
	        	"guid": "offering-1-guid"
			},
      		"entity": {
		        "label": "Offering 1",
		        "provider": "Offering 1 provider",
		        "description": "Offering 1 description",
		        "version" : "1.0"
			}
	    },
    	{
      		"metadata": {
        		"guid": "offering-2-guid"
	      	},
	      	"entity": {
		        "label": "Offering 2",
		        "provider": "Offering 2 provider",
		        "description": "Offering 2 description",
		        "version" : "1.5"
	        }
    	}
	]}`,
}

var serviceOfferingReq = apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/services/the-service-guid",
	Response: testnet.TestResponse{Status: http.StatusOK, Body: `
		{
		  "metadata": {
			"guid": "15790581-a293-489b-9efc-847ecf1b1339"
		  },
		  "entity": {
			"label": "mysql",
			"provider": "mysql",
		    "extra": "{\"documentationURL\":\"http://info.example.com\"}",
			"description": "MySQL database",
			"requires": ["route_forwarding"]
		  }
		}`,
	}})

var serviceOfferingNullExtraReq = apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/services/the-service-guid",
	Response: testnet.TestResponse{Status: http.StatusOK, Body: `
		{
		  "metadata": {
			"guid": "15790581-a293-489b-9efc-847ecf1b1339"
		  },
		  "entity": {
			"label": "mysql",
			"provider": "mysql",
		    "extra": null,
			"description": "MySQL database"
		  }
		}`,
	}})

var findServiceInstanceReq = apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/spaces/my-space-guid/service_instances?return_user_provided_service_instances=true&q=name%3Amy-service",
	Response: testnet.TestResponse{Status: http.StatusOK, Body: `
	{"resources": [
        {
          "metadata": {
            "guid": "my-service-instance-guid"
          },
          "entity": {
            "name": "my-service",
			"dashboard_url":"my-dashboard-url",
            "service_bindings": [
              {
                "metadata": {
                  "guid": "service-binding-1-guid",
                  "url": "/v2/service_bindings/service-binding-1-guid"
                },
                "entity": {
                  "app_guid": "app-1-guid"
                }
              },
              {
                "metadata": {
                  "guid": "service-binding-2-guid",
                  "url": "/v2/service_bindings/service-binding-2-guid"
                },
                "entity": {
                  "app_guid": "app-2-guid"
                }
              }
            ],
            "service_plan": {
              "metadata": {
                "guid": "plan-guid"
              },
              "entity": {
                "name": "plan-name",
                "service_guid": "the-service-guid"
              }
            }
          }
        }
    ]}`}})
