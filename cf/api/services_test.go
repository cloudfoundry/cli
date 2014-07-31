package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	"github.com/cloudfoundry/cli/testhelpers/maker"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"

	. "github.com/cloudfoundry/cli/cf/api"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Services Repo", func() {
	var (
		testServer  *httptest.Server
		testHandler *testnet.TestHandler
		configRepo  configuration.ReadWriter
		repo        ServiceRepository
	)

	setupTestServer := func(reqs ...testnet.TestRequest) {
		testServer, testHandler = testnet.NewServer(reqs)
		configRepo.SetApiEndpoint(testServer.URL)
	}

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		configRepo.SetAccessToken("BEARER my_access_token")

		gateway := net.NewCloudControllerGateway((configRepo), time.Now)
		repo = NewCloudControllerServiceRepository(configRepo, gateway)
	})

	AfterEach(func() {
		testServer.Close()
	})

	Describe("GetAllServiceOfferings", func() {
		BeforeEach(func() {
			setupTestServer(
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     "/v2/services?inline-relations-depth=1",
					Response: firstOfferingsResponse,
				}),
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     "/v2/services?inline-relations-depth=1",
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
			Expect(firstOffering.Guid).To(Equal("first-offering-1-guid"))
		})
	})

	Describe("GetServiceOfferingsForSpace", func() {
		It("gets all service offerings in a given space", func() {
			setupTestServer(
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     "/v2/spaces/my-space-guid/services?inline-relations-depth=1",
					Response: firstOfferingsForSpaceResponse,
				}),
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     "/v2/spaces/my-space-guid/services?inline-relations-depth=1",
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
			Expect(firstOffering.Guid).To(Equal("first-offering-1-guid"))
			Expect(len(firstOffering.Plans)).To(Equal(2))

			secondOffering := offerings[1]
			Expect(secondOffering.Label).To(Equal("Offering 1"))
			Expect(secondOffering.Version).To(Equal("1.0"))
			Expect(secondOffering.Description).To(Equal("Offering 1 description"))
			Expect(secondOffering.Provider).To(Equal("Offering 1 provider"))
			Expect(secondOffering.Guid).To(Equal("offering-1-guid"))
			Expect(len(secondOffering.Plans)).To(Equal(2))
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
            "version": "damageableness-preheat",
            "documentation_url": "YESWECAN.com"
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
            "version": "seraphine-lowdah",
            "documentation_url": "YESWECAN.com"
         }
      }
   ]
}`

			setupTestServer(
				testapi.NewCloudControllerTestRequest(
					testnet.TestRequest{
						Method:   "GET",
						Path:     "/v2/services?q=service_broker_guid%3Amy-service-broker-guid",
						Response: testnet.TestResponse{Status: http.StatusOK, Body: body1},
					}),
				testapi.NewCloudControllerTestRequest(
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

			Expect(services[0].Guid).To(Equal("my-service-guid"))
			Expect(services[1].Guid).To(Equal("my-service-guid2"))
		})
	})

	Describe("creating a service instance", func() {
		It("makes the right request", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "POST",
				Path:     "/v2/service_instances",
				Matcher:  testnet.RequestBodyMatcher(`{"name":"instance-name","service_plan_guid":"plan-guid","space_guid":"my-space-guid","async":true}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			}))

			err := repo.CreateServiceInstance("instance-name", "plan-guid")
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the name is taken but an identical service exists", func() {
			BeforeEach(func() {
				setupTestServer(
					testapi.NewCloudControllerTestRequest(testnet.TestRequest{
						Method:  "POST",
						Path:    "/v2/service_instances",
						Matcher: testnet.RequestBodyMatcher(`{"name":"my-service","service_plan_guid":"plan-guid","space_guid":"my-space-guid","async":true}`),
						Response: testnet.TestResponse{
							Status: http.StatusBadRequest,
							Body:   `{"code":60002,"description":"The service instance name is taken: my-service"}`,
						}}),
					findServiceInstanceReq,
					serviceOfferingReq)
			})

			It("returns a ModelAlreadyExistsError if the plan is the same", func() {
				err := repo.CreateServiceInstance("my-service", "plan-guid")
				Expect(testHandler).To(HaveAllRequestsCalled())
				Expect(err).To(BeAssignableToTypeOf(&errors.ModelAlreadyExistsError{}))
			})
		})

		Context("when the name is taken and no identical service instance exists", func() {
			BeforeEach(func() {
				setupTestServer(
					testapi.NewCloudControllerTestRequest(testnet.TestRequest{
						Method:  "POST",
						Path:    "/v2/service_instances",
						Matcher: testnet.RequestBodyMatcher(`{"name":"my-service","service_plan_guid":"different-plan-guid","space_guid":"my-space-guid","async":true}`),
						Response: testnet.TestResponse{
							Status: http.StatusBadRequest,
							Body:   `{"code":60002,"description":"The service instance name is taken: my-service"}`,
						}}),
					findServiceInstanceReq,
					serviceOfferingReq)
			})

			It("fails if the plan is different", func() {
				err := repo.CreateServiceInstance("my-service", "different-plan-guid")

				Expect(testHandler).To(HaveAllRequestsCalled())
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(errors.NewHttpError(400, "", "")))
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
			Expect(instance.Guid).To(Equal("my-service-instance-guid"))
			Expect(instance.ServiceOffering.Label).To(Equal("mysql"))
			Expect(instance.ServiceOffering.DocumentationUrl).To(Equal("http://info.example.com"))
			Expect(instance.ServiceOffering.Description).To(Equal("MySQL database"))
			Expect(instance.ServicePlan.Name).To(Equal("plan-name"))
			Expect(len(instance.ServiceBindings)).To(Equal(2))

			binding := instance.ServiceBindings[0]
			Expect(binding.Url).To(Equal("/v2/service_bindings/service-binding-1-guid"))
			Expect(binding.Guid).To(Equal("service-binding-1-guid"))
			Expect(binding.AppGuid).To(Equal("app-1-guid"))
		})

		It("returns user provided services", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
			Expect(instance.Guid).To(Equal("my-service-instance-guid"))
			Expect(instance.ServiceOffering.Label).To(Equal(""))
			Expect(instance.ServicePlan.Name).To(Equal(""))
			Expect(len(instance.ServiceBindings)).To(Equal(2))

			binding := instance.ServiceBindings[0]
			Expect(binding.Url).To(Equal("/v2/service_bindings/service-binding-1-guid"))
			Expect(binding.Guid).To(Equal("service-binding-1-guid"))
			Expect(binding.AppGuid).To(Equal("app-1-guid"))
		})

		It("it returns a failure response when the instance doesn't exist", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/spaces/my-space-guid/service_instances?return_user_provided_service_instances=true&q=name%3Amy-service",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{ "resources": [] }`},
			}))

			_, err := repo.FindInstanceByName("my-service")

			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).To(BeAssignableToTypeOf(&errors.ModelNotFoundError{}))
		})
	})

	Describe("DeleteService", func() {
		It("it deletes the service when no apps are bound", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "DELETE",
				Path:     "/v2/service_instances/my-service-instance-guid",
				Response: testnet.TestResponse{Status: http.StatusOK},
			}))

			serviceInstance := models.ServiceInstance{}
			serviceInstance.Guid = "my-service-instance-guid"

			err := repo.DeleteService(serviceInstance)
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).NotTo(HaveOccurred())
		})

		It("doesn't delete the service when apps are bound", func() {
			setupTestServer()

			serviceInstance := models.ServiceInstance{}
			serviceInstance.Guid = "my-service-instance-guid"
			serviceInstance.ServiceBindings = []models.ServiceBindingFields{
				{
					Url:     "/v2/service_bindings/service-binding-1-guid",
					AppGuid: "app-1-guid",
				},
				{
					Url:     "/v2/service_bindings/service-binding-2-guid",
					AppGuid: "app-2-guid",
				},
			}

			err := repo.DeleteService(serviceInstance)
			Expect(err.Error()).To(Equal("Cannot delete service instance, apps are still bound to it"))
		})
	})

	Describe("RenameService", func() {
		Context("when the service is not user provided", func() {

			BeforeEach(func() {
				setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "PUT",
					Path:     "/v2/service_instances/my-service-instance-guid",
					Matcher:  testnet.RequestBodyMatcher(`{"name":"new-name"}`),
					Response: testnet.TestResponse{Status: http.StatusCreated},
				}))
			})

			It("renames the service", func() {
				serviceInstance := models.ServiceInstance{}
				serviceInstance.Guid = "my-service-instance-guid"
				serviceInstance.ServicePlan = models.ServicePlanFields{
					Guid: "some-plan-guid",
				}

				err := repo.RenameService(serviceInstance, "new-name")
				Expect(testHandler).To(HaveAllRequestsCalled())
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the service is user provided", func() {
			BeforeEach(func() {
				setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "PUT",
					Path:     "/v2/user_provided_service_instances/my-service-instance-guid",
					Matcher:  testnet.RequestBodyMatcher(`{"name":"new-name"}`),
					Response: testnet.TestResponse{Status: http.StatusCreated},
				}))
			})

			It("renames the service", func() {
				serviceInstance := models.ServiceInstance{}
				serviceInstance.Guid = "my-service-instance-guid"

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
				Expect(offering.Guid).To(Equal("offering-1-guid"))
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
				Expect(offering.Guid).To(Equal(""))
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
			Expect(err.(errors.HttpError).ErrorCode()).To(Equal("10005"))
		})
	})

	Describe("FindServiceOfferingByLabel", func() {
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
				offering, err := repo.FindServiceOfferingByLabel("offering-1")
				Expect(offering.Guid).To(Equal("offering-1-guid"))
				Expect(offering.Label).To(Equal("offering-1"))
				Expect(offering.Provider).To(Equal("provider-1"))
				Expect(offering.Description).To(Equal("offering 1 description"))
				Expect(offering.Version).To(Equal("1.0"))
				Expect(offering.BrokerGuid).To(Equal("broker-1-guid"))
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
				offering, err := repo.FindServiceOfferingByLabel("offering-1")

				Expect(err).To(BeAssignableToTypeOf(&errors.ModelNotFoundError{}))
				Expect(offering.Guid).To(Equal(""))
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

			_, err := repo.FindServiceOfferingByLabel("offering-1")
			Expect(err).To(HaveOccurred())
			Expect(err.(errors.HttpError).ErrorCode()).To(Equal("10005"))
		})
	})

	Describe("GetServiceOfferingByGuid", func() {
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
				offering, err := repo.GetServiceOfferingByGuid("offering-1-guid")
				Expect(offering.Guid).To(Equal("offering-1-guid"))
				Expect(offering.Label).To(Equal("offering-1"))
				Expect(offering.Provider).To(Equal("provider-1"))
				Expect(offering.Description).To(Equal("offering 1 description"))
				Expect(offering.Version).To(Equal("1.0"))
				Expect(offering.BrokerGuid).To(Equal("broker-1-guid"))
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
				offering, err := repo.GetServiceOfferingByGuid("offering-1-guid")

				Expect(err).To(BeAssignableToTypeOf(&errors.HttpNotFoundError{}))
				Expect(offering.Guid).To(Equal(""))
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

			offering := maker.NewServiceOffering("the-offering")
			offering.Guid = "the-service-guid"

			err := repo.PurgeServiceOffering(offering)
			Expect(err).NotTo(HaveOccurred())
			Expect(testHandler).To(HaveAllRequestsCalled())
		})
	})

	Describe("getting the count of service instances for a service plan", func() {
		var planGuid = "abc123"

		It("returns the number of service instances", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   fmt.Sprintf("/v2/service_plans/%s/service_instances?results-per-page=1", planGuid),
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

			count, err := repo.GetServiceInstanceCountForServicePlan(planGuid)
			Expect(count).To(Equal(9))
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the API error when one occurs", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     fmt.Sprintf("/v2/service_plans/%s/service_instances?results-per-page=1", planGuid),
				Response: testnet.TestResponse{Status: http.StatusInternalServerError},
			}))

			_, err := repo.GetServiceInstanceCountForServicePlan(planGuid)
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

				setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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

				setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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

				setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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

				setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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

				setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "GET",
					Path:   fmt.Sprintf("/v2/services?inline-relations-depth=1&q=%s", url.QueryEscape("label:v2-service-label;provider:")),
					Response: testnet.TestResponse{
						Status: http.StatusInternalServerError,
					}}))
			})

			It("returns an error", func() {
				_, err := repo.FindServicePlanByDescription(planDescription)

				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(errors.NewHttpError(500, "", "")))
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
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
					Path:   fmt.Sprintf("/v2/spaces/my-space-guid/services?q=%s&inline-relations-depth=1", url.QueryEscape("label:offering-1")),
					Response: testnet.TestResponse{
						Status: 200,
						Body: `
						{
							"next_url": "/v2/spaces/my-space-guid/services?q=label%3Aoffering-1&inline-relations-depth=1&page=2",
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
						}`}},
				testnet.TestRequest{
					Method: "GET",
					Path:   fmt.Sprintf("/v2/spaces/my-space-guid/services?q=%s&inline-relations-depth=1", url.QueryEscape("label:offering-1")),
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
										"version" : "1.0",
										"service_plans": []
									}
								}
							]
						}`}})

			offerings, err := repo.FindServiceOfferingsForSpaceByLabel("my-space-guid", "offering-1")
			Expect(err).ToNot(HaveOccurred())
			Expect(offerings).To(HaveLen(2))
			Expect(offerings[0].Guid).To(Equal("offering-1-guid"))
		})

		It("returns an error if the offering cannot be found", func() {
			setupTestServer(testnet.TestRequest{
				Method: "GET",
				Path:   fmt.Sprintf("/v2/spaces/my-space-guid/services?q=%s&inline-relations-depth=1", url.QueryEscape("label:offering-1")),
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
				Path:   fmt.Sprintf("/v2/spaces/my-space-guid/services?q=%s&inline-relations-depth=1", url.QueryEscape("label:offering-1")),
				Response: testnet.TestResponse{
					Status: http.StatusBadRequest,
					Body: `{
						"code": 9001,
						"description": "Something Happened"
					}`,
				},
			})

			_, err := repo.FindServiceOfferingsForSpaceByLabel("my-space-guid", "offering-1")
			Expect(err).To(BeAssignableToTypeOf(errors.NewHttpError(400, "", "")))
		})

		Describe("when api returns query by label is invalid", func() {
			It("makes a backwards-compatible request", func() {
				failedRequestByQueryLabel := testnet.TestRequest{
					Method: "GET",
					Path:   fmt.Sprintf("/v2/spaces/my-space-guid/services?q=%s&inline-relations-depth=1", url.QueryEscape("label:my-service-offering")),
					Response: testnet.TestResponse{
						Status: http.StatusBadRequest,
						Body:   `{"code": 10005,"description": "The query parameter is invalid"}`,
					},
				}

				firstPaginatedRequest := testnet.TestRequest{
					Method: "GET",
					Path:   fmt.Sprintf("/v2/spaces/my-space-guid/services?inline-relations-depth=1"),
					Response: testnet.TestResponse{
						Status: http.StatusOK,
						Body: `{
							"next_url": "/v2/spaces/my-space-guid/services?page=2&inline-relations-depth=1",
							"resources": [
								{
								  "metadata": {
									"guid": "my-service-offering-guid"
								  },
								  "entity": {
									"label": "my-service-offering",
									"provider": "some-other-provider",
									"description": "a description that does not match your provider",
									"version" : "1.0",
									"service_plans": []
								  }
								}
							]
						}`,
					},
				}

				secondPaginatedRequest := testnet.TestRequest{
					Method: "GET",
					Path:   fmt.Sprintf("/v2/spaces/my-space-guid/services?inline-relations-depth=1"),
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
											"version" : "1.0",
											"service_plans": []
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
	"next_url": "/v2/services?inline-relations-depth=1&page=2",
	"resources": [
	{
		"metadata": {
			"guid": "first-offering-1-guid"
		},
		"entity": {
			"label": "first-Offering 1",
			"provider": "Offering 1 provider",
			"description": "first Offering 1 description",
			"version" : "1.0",
			"service_plans": [
				{
					"metadata": {"guid": "first-offering-1-plan-1-guid"},
					"entity": {"name": "first Offering 1 Plan 1"}
				},
				{
					"metadata": {"guid": "first-offering-1-plan-2-guid"},
					"entity": {"name": "first Offering 1 Plan 2"}
				}
	        ]
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
				"version" : "1.0",
				"service_plans": [
				{
					"metadata": {"guid": "first-offering-1-plan-1-guid"},
					"entity": {"name": "first Offering 1 Plan 1"}
				},
				{
					"metadata": {"guid": "first-offering-1-plan-2-guid"},
					"entity": {"name": "first Offering 1 Plan 2"}
				}]
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
		        "version" : "1.0",
		        "service_plans": [
		            {
		                "metadata": {"guid": "offering-1-plan-1-guid"},
		                "entity": {"name": "Offering 1 Plan 1"}
		            },
		            {
		                "metadata": {"guid": "offering-1-plan-2-guid"},
		                "entity": {"name": "Offering 1 Plan 2"}
		            }
		        ]
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
		        "version" : "1.5",
		        "service_plans": [
		            {
		                "metadata": {"guid": "offering-2-plan-1-guid"},
		                "entity": {"name": "Offering 2 Plan 1"}
		            }
	        	]
	        }
    	}
	]}`,
}

var serviceOfferingReq = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
			"documentation_url": "http://info.example.com",
			"description": "MySQL database"
		  }
		}`,
	}})

var findServiceInstanceReq = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
