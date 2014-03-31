package api_test

import (
	. "cf/api"
	"cf/api/resources"
	"cf/configuration"
	"cf/errors"
	"cf/models"
	"cf/net"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"net/url"
	testapi "testhelpers/api"
	testconfig "testhelpers/configuration"
	"testhelpers/maker"
	testnet "testhelpers/net"
)

var _ = Describe("Services Repo", func() {
	It("gets all public service offerings", func() {
		config := testconfig.NewRepository()
		config.SetAccessToken("BEARER my_access_token")

		testServer, handler, repo := createServiceRepoWithConfig([]testnet.TestRequest{
			testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/services?inline-relations-depth=1",
				Response: multipleOfferingsResponse,
			}),
		}, config)
		defer testServer.Close()

		offerings, apiErr := repo.GetAllServiceOfferings()

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
		expectMultipleServiceOfferings(offerings)
	})

	It("gets all service offerings in a given space", func() {
		config := testconfig.NewRepository()
		config.SetAccessToken("BEARER my_access_token")

		testServer, handler, repo := createServiceRepoWithConfig([]testnet.TestRequest{
			testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/spaces/my-space-guid/services?inline-relations-depth=1",
				Response: multipleOfferingsResponse,
			}),
		}, config)
		defer testServer.Close()

		offerings, apiErr := repo.GetServiceOfferingsForSpace("my-space-guid")

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
		expectMultipleServiceOfferings(offerings)
	})

	Describe("creating a service instance", func() {
		It("makes the right request", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "POST",
				Path:     "/v2/service_instances",
				Matcher:  testnet.RequestBodyMatcher(`{"name":"instance-name","service_plan_guid":"plan-guid","space_guid":"my-space-guid","async":true}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			testServer, handler, repo := createServiceRepo([]testnet.TestRequest{req})
			defer testServer.Close()

			apiErr := repo.CreateServiceInstance("instance-name", "plan-guid")
			Expect(handler).To(testnet.HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})

		It("returns a successful response when an identical service instance already exists", func() {
			errorReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:  "POST",
				Path:    "/v2/service_instances",
				Matcher: testnet.RequestBodyMatcher(`{"name":"my-service","service_plan_guid":"plan-guid","space_guid":"my-space-guid","async":true}`),
				Response: testnet.TestResponse{
					Status: http.StatusBadRequest,
					Body:   `{"code":60002,"description":"The service instance name is taken: my-service"}`,
				}})

			testServer, handler, repo := createServiceRepo([]testnet.TestRequest{errorReq, findServiceInstanceReq, serviceOfferingReq})
			defer testServer.Close()

			apiErr := repo.CreateServiceInstance("my-service", "plan-guid")

			Expect(handler).To(testnet.HaveAllRequestsCalled())

			_, ok := apiErr.(*errors.ServiceInstanceAlreadyExistsError)
			Expect(ok).To(BeTrue())
		})

		It("fails when a different service instance with the same name already exists", func() {
			errorReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:  "POST",
				Path:    "/v2/service_instances",
				Matcher: testnet.RequestBodyMatcher(`{"name":"my-service","service_plan_guid":"different-plan-guid","space_guid":"my-space-guid","async":true}`),
				Response: testnet.TestResponse{
					Status: http.StatusBadRequest,
					Body:   `{"code":60002,"description":"The service instance name is taken: my-service"}`,
				}})

			testServer, handler, repo := createServiceRepo([]testnet.TestRequest{errorReq, findServiceInstanceReq, serviceOfferingReq})
			defer testServer.Close()

			apiErr := repo.CreateServiceInstance("my-service", "different-plan-guid")

			Expect(handler).To(testnet.HaveAllRequestsCalled())
			_, ok := apiErr.(*errors.ServiceInstanceAlreadyExistsError)
			Expect(ok).To(BeFalse())
		})
	})

	Describe("finding service instances by name", func() {
		It("returns the service instance", func() {
			testServer, handler, repo := createServiceRepo([]testnet.TestRequest{findServiceInstanceReq, serviceOfferingReq})
			defer testServer.Close()

			instance, apiErr := repo.FindInstanceByName("my-service")

			Expect(handler).To(testnet.HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
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
			testServer, handler, repo := createServiceRepo([]testnet.TestRequest{findUserProvidedServiceInstanceReq})
			defer testServer.Close()

			instance, apiErr := repo.FindInstanceByName("my-service")

			Expect(handler).To(testnet.HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
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
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/spaces/my-space-guid/service_instances?return_user_provided_service_instances=true&q=name%3Amy-service",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{ "resources": [] }`},
			})

			testServer, handler, repo := createServiceRepo([]testnet.TestRequest{req})
			defer testServer.Close()

			_, apiErr := repo.FindInstanceByName("my-service")
			Expect(handler).To(testnet.HaveAllRequestsCalled())

			Expect(apiErr.(*errors.ModelNotFoundError)).NotTo(BeNil())
		})
	})

	It("TestDeleteServiceWithoutServiceBindings", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "DELETE",
			Path:     "/v2/service_instances/my-service-instance-guid",
			Response: testnet.TestResponse{Status: http.StatusOK},
		})
		testServer, handler, repo := createServiceRepo([]testnet.TestRequest{req})
		defer testServer.Close()

		serviceInstance := models.ServiceInstance{}
		serviceInstance.Guid = "my-service-instance-guid"
		apiErr := repo.DeleteService(serviceInstance)
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})

	It("TestDeleteServiceWithServiceBindings", func() {
		testServer, _, repo := createServiceRepo([]testnet.TestRequest{})
		defer testServer.Close()

		serviceInstance := models.ServiceInstance{}
		serviceInstance.Guid = "my-service-instance-guid"

		binding := models.ServiceBindingFields{}
		binding.Url = "/v2/service_bindings/service-binding-1-guid"
		binding.AppGuid = "app-1-guid"

		binding2 := models.ServiceBindingFields{}
		binding2.Url = "/v2/service_bindings/service-binding-2-guid"
		binding2.AppGuid = "app-2-guid"

		serviceInstance.ServiceBindings = []models.ServiceBindingFields{binding, binding2}

		apiErr := repo.DeleteService(serviceInstance)
		Expect(apiErr).NotTo(BeNil())
		Expect(apiErr.Error()).To(Equal("Cannot delete service instance, apps are still bound to it"))
	})

	It("TestRenameService", func() {
		path := "/v2/service_instances/my-service-instance-guid"
		serviceInstance := models.ServiceInstance{}
		serviceInstance.Guid = "my-service-instance-guid"

		plan := models.ServicePlanFields{}
		plan.Guid = "some-plan-guid"
		serviceInstance.ServicePlan = plan

		testRenameService(path, serviceInstance)
	})

	It("TestRenameServiceWhenServiceIsUserProvided", func() {
		path := "/v2/user_provided_service_instances/my-service-instance-guid"
		serviceInstance := models.ServiceInstance{}
		serviceInstance.Guid = "my-service-instance-guid"
		testRenameService(path, serviceInstance)
	})

	It("finds service offerings by label and provider", func() {
		testServer, _, repo := createServiceRepo([]testnet.TestRequest{{
			Method: "GET",
			Path:   fmt.Sprintf("/v2/services?q=%s", url.QueryEscape("label:offering-1;provider:provider-1")),
			Response: testnet.TestResponse{
				Status: 200,
				Body: `{
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
        }`,
			},
		}})
		defer testServer.Close()

		offering, apiErr := repo.FindServiceOfferingByLabelAndProvider("offering-1", "provider-1")
		Expect(offering.Guid).To(Equal("offering-1-guid"))
		Expect(apiErr).NotTo(HaveOccurred())
	})

	It("returns an error if the offering cannot be found", func() {
		testServer, _, repo := createServiceRepo([]testnet.TestRequest{{
			Method: "GET",
			Path:   fmt.Sprintf("/v2/services?q=%s", url.QueryEscape("label:offering-1;provider:provider-1")),
			Response: testnet.TestResponse{
				Status: 200,
				Body: `{
            "next_url": null,
            "resources": []
        }`,
			},
		}})
		defer testServer.Close()

		offering, apiErr := repo.FindServiceOfferingByLabelAndProvider("offering-1", "provider-1")
		Expect(apiErr.(*errors.ModelNotFoundError)).NotTo(BeNil())
		Expect(offering.Guid).To(Equal(""))
	})

	It("handles api errors when finding service offerings", func() {
		testServer, _, repo := createServiceRepo([]testnet.TestRequest{{
			Method: "GET",
			Path:   fmt.Sprintf("/v2/services?q=%s", url.QueryEscape("label:offering-1;provider:provider-1")),
			Response: testnet.TestResponse{
				Status: 400,
				Body: `{
            "code": 10005,
            "description": "The query parameter is invalid"
        }`,
			},
		}})
		defer testServer.Close()

		_, apiErr := repo.FindServiceOfferingByLabelAndProvider("offering-1", "provider-1")
		Expect(apiErr).To(HaveOccurred())
		Expect(apiErr.(errors.HttpError).ErrorCode()).To(Equal("10005"))
	})

	It("purges service offerings", func() {
		testServer, handler, repo := createServiceRepo([]testnet.TestRequest{{
			Method: "DELETE",
			Path:   "/v2/services/the-service-guid?purge=true",
			Response: testnet.TestResponse{
				Status: 204,
			},
		}})
		defer testServer.Close()

		offering := maker.NewServiceOffering("the-offering")
		offering.Guid = "the-service-guid"

		apiErr := repo.PurgeServiceOffering(offering)
		Expect(apiErr).NotTo(HaveOccurred())
		Expect(handler).To(testnet.HaveAllRequestsCalled())
	})

	Describe("getting the count of service instances for a service plan", func() {
		var planGuid = "abc123"

		It("returns the number of service instances", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
			})
			testServer, _, repo := createServiceRepo([]testnet.TestRequest{req})
			defer testServer.Close()

			count, apiErr := repo.GetServiceInstanceCountForServicePlan(planGuid)
			Expect(count).To(Equal(9))
			Expect(apiErr).NotTo(HaveOccurred())
		})

		It("returns the API error when one occurs", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     fmt.Sprintf("/v2/service_plans/%s/service_instances?results-per-page=1", planGuid),
				Response: testnet.TestResponse{Status: http.StatusInternalServerError},
			})
			testServer, _, repo := createServiceRepo([]testnet.TestRequest{req})
			defer testServer.Close()

			_, apiErr := repo.GetServiceInstanceCountForServicePlan(planGuid)

			Expect(apiErr).To(HaveOccurred())
		})
	})

	Describe("finding a service plan", func() {
		Context("when we find a matching plan", func() {
			It("returns the plan guid for a v1 plan", func() {
				req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
                        }`}})

				testServer, _, repo := createServiceRepo([]testnet.TestRequest{req})
				defer testServer.Close()

				v1 := resources.ServicePlanDescription{
					ServiceLabel:    "v1-elephantsql",
					ServicePlanName: "v1-panda",
					ServiceProvider: "v1-elephantsql",
				}

				v1Guid, apiErr := repo.FindServicePlanByDescription(v1)

				Expect(v1Guid).To(Equal("offering-1-plan-2-guid"))
				Expect(apiErr).NotTo(HaveOccurred())
			})

			It("returns the plan guid for a v2 plan", func() {
				req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
                        }`}})

				testServer, _, repo := createServiceRepo([]testnet.TestRequest{req})
				defer testServer.Close()

				v2 := resources.ServicePlanDescription{
					ServiceLabel:    "v2-elephantsql",
					ServicePlanName: "v2-panda",
				}

				v2Guid, apiErr := repo.FindServicePlanByDescription(v2)

				Expect(apiErr).NotTo(HaveOccurred())
				Expect(v2Guid).To(Equal("offering-1-plan-2-guid"))
			})
		})

		Context("when no service matches the description", func() {
			It("returns an apiErr error", func() {
				req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     fmt.Sprintf("/v2/services?inline-relations-depth=1&q=%s", url.QueryEscape("label:v2-service-label;provider:")),
					Response: testnet.TestResponse{Status: http.StatusOK, Body: `{ "resources": [] }`},
				})

				testServer, _, repo := createServiceRepo([]testnet.TestRequest{req})
				defer testServer.Close()

				v2 := resources.ServicePlanDescription{
					ServiceLabel:    "v2-service-label",
					ServicePlanName: "v2-plan-name",
				}

				_, apiErr := repo.FindServicePlanByDescription(v2)

				Expect(apiErr).To(HaveOccurred())
				Expect(apiErr.(*errors.ModelNotFoundError)).NotTo(BeNil())
				Expect(apiErr.Error()).To(ContainSubstring("Plan"))
				Expect(apiErr.Error()).To(ContainSubstring("v2-service-label v2-plan-name"))
				Expect(apiErr.Error()).To(ContainSubstring("not found"))
			})
		})

		Context("when the described service has no matching plan", func() {
			It("returns apiErr error", func() {
				req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
                        }`}})

				testServer, _, repo := createServiceRepo([]testnet.TestRequest{req})
				defer testServer.Close()

				v2 := resources.ServicePlanDescription{
					ServiceLabel:    "v2-service-label",
					ServicePlanName: "v2-plan-name",
				}

				_, apiErr := repo.FindServicePlanByDescription(v2)

				Expect(apiErr).To(HaveOccurred())
				Expect(apiErr.(*errors.ModelNotFoundError)).NotTo(BeNil())
				Expect(apiErr.Error()).To(ContainSubstring("Plan"))
				Expect(apiErr.Error()).To(ContainSubstring("v2-service-label v2-plan-name"))
				Expect(apiErr.Error()).To(ContainSubstring("not found"))
			})
		})

		Context("when we get an HTTP error", func() {
			It("returns that apiErr error", func() {
				req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "GET",
					Path:     fmt.Sprintf("/v2/services?inline-relations-depth=1&q=%s", url.QueryEscape("label:v2-service-label;provider:")),
					Response: testnet.TestResponse{Status: http.StatusInternalServerError}})

				testServer, _, repo := createServiceRepo([]testnet.TestRequest{req})
				defer testServer.Close()

				v2 := resources.ServicePlanDescription{
					ServiceLabel:    "v2-service-label",
					ServicePlanName: "v2-plan-name",
				}

				_, apiErr := repo.FindServicePlanByDescription(v2)

				Expect(apiErr).To(HaveOccurred())
				_, ok := apiErr.(errors.HttpError)
				Expect(ok).To(BeTrue())
			})
		})
	})

	Describe("migrating service plans", func() {
		It("makes a request to CC to migrate the instances from v1 to v2", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/service_plans/v1-guid/service_instances",
				Matcher:  testnet.RequestBodyMatcher(`{"service_plan_guid":"v2-guid"}`),
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"changed_count":3}`},
			})

			testServer, _, repo := createServiceRepo([]testnet.TestRequest{req})
			defer testServer.Close()

			changedCount, apiErr := repo.MigrateServicePlanFromV1ToV2("v1-guid", "v2-guid")
			Expect(apiErr).NotTo(HaveOccurred())
			Expect(changedCount).To(Equal(3))
		})

		It("returns an error when migrating fails", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/service_plans/v1-guid/service_instances",
				Matcher:  testnet.RequestBodyMatcher(`{"service_plan_guid":"v2-guid"}`),
				Response: testnet.TestResponse{Status: http.StatusInternalServerError},
			})

			testServer, _, repo := createServiceRepo([]testnet.TestRequest{req})
			defer testServer.Close()

			_, apiErr := repo.MigrateServicePlanFromV1ToV2("v1-guid", "v2-guid")
			Expect(apiErr).To(HaveOccurred())
		})
	})

	Describe("FindServiceOfferingsForSpaceByLabel", func() {
		It("finds service offerings within a space by label", func() {
			_, _, repo := createServiceRepo([]testnet.TestRequest{{
				Method: "GET",
				Path:   fmt.Sprintf("/v2/spaces/my-space-guid/services?q=%s&inline-relations-depth=1", url.QueryEscape("label:offering-1")),
				Response: testnet.TestResponse{
					Status: 200,
					Body: `{
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
        }`,
				},
			}})

			offerings, apiErr := repo.FindServiceOfferingsForSpaceByLabel("my-space-guid", "offering-1")
			Expect(apiErr).ToNot(HaveOccurred())
			Expect(offerings).To(HaveLen(1))
			Expect(offerings[0].Guid).To(Equal("offering-1-guid"))
		})

		It("returns an error if the offering cannot be found", func() {
			_, _, repo := createServiceRepo([]testnet.TestRequest{{
				Method: "GET",
				Path:   fmt.Sprintf("/v2/spaces/my-space-guid/services?q=%s&inline-relations-depth=1", url.QueryEscape("label:offering-1")),
				Response: testnet.TestResponse{
					Status: 200,
					Body: `{
						"next_url": null,
						"resources": []
					}`,
				},
			}})

			offerings, apiErr := repo.FindServiceOfferingsForSpaceByLabel("my-space-guid", "offering-1")
			Expect(apiErr.(*errors.ModelNotFoundError)).NotTo(BeNil())
			Expect(offerings).To(HaveLen(0))
		})

		It("handles api errors when finding service offerings", func() {
			_, _, repo := createServiceRepo([]testnet.TestRequest{{
				Method: "GET",
				Path:   fmt.Sprintf("/v2/spaces/my-space-guid/services?q=%s&inline-relations-depth=1", url.QueryEscape("label:offering-1")),
				Response: testnet.TestResponse{
					Status: 400,
					Body: `{
						"code": 9001,
						"description": "Something Happened"
					}`,
				},
			}})

			_, apiErr := repo.FindServiceOfferingsForSpaceByLabel("my-space-guid", "offering-1")
			Expect(apiErr.(errors.HttpError)).NotTo(BeNil())
		})

		Describe("when api returns query by label is invalid", func() {
			It("makes a backwards-compatible request", func() {
				failedRequestByQueryLabel := testnet.TestRequest{
					Method: "GET",
					Path:   fmt.Sprintf("/v2/spaces/my-space-guid/services?q=%s&inline-relations-depth=1", url.QueryEscape("label:my-service-offering")),
					Response: testnet.TestResponse{
						Status: 400,
						Body:   `{"code": 10005,"description": "The query parameter is invalid"}`,
					},
				}

				firstPaginatedRequest := testnet.TestRequest{
					Method: "GET",
					Path:   fmt.Sprintf("/v2/spaces/my-space-guid/services?inline-relations-depth=1"),
					Response: testnet.TestResponse{
						Status: 200,
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
						Status: 200,
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

				testServer, _, repo := createServiceRepo([]testnet.TestRequest{failedRequestByQueryLabel, firstPaginatedRequest, secondPaginatedRequest})
				defer testServer.Close()

				serviceOfferings, apiErr := repo.FindServiceOfferingsForSpaceByLabel("my-space-guid", "my-service-offering")
				Expect(apiErr).NotTo(HaveOccurred())
				Expect(len(serviceOfferings)).To(Equal(2))
			})
		})
	})
})

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
  ]
}`}

func expectMultipleServiceOfferings(offerings []models.ServiceOffering) {
	Expect(len(offerings)).To(Equal(2))

	firstOffering := offerings[0]
	Expect(firstOffering.Label).To(Equal("Offering 1"))
	Expect(firstOffering.Version).To(Equal("1.0"))
	Expect(firstOffering.Description).To(Equal("Offering 1 description"))
	Expect(firstOffering.Provider).To(Equal("Offering 1 provider"))
	Expect(firstOffering.Guid).To(Equal("offering-1-guid"))
	Expect(len(firstOffering.Plans)).To(Equal(2))

	plan := firstOffering.Plans[0]
	Expect(plan.Name).To(Equal("Offering 1 Plan 1"))
	Expect(plan.Guid).To(Equal("offering-1-plan-1-guid"))

	secondOffering := offerings[1]
	Expect(secondOffering.Label).To(Equal("Offering 2"))
	Expect(secondOffering.Guid).To(Equal("offering-2-guid"))
	Expect(len(secondOffering.Plans)).To(Equal(1))
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

var findUserProvidedServiceInstanceReq = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
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
            "service_plan_guid": null
          }
        }
    ]}`}})

func testRenameService(endpointPath string, serviceInstance models.ServiceInstance) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     endpointPath,
		Matcher:  testnet.RequestBodyMatcher(`{"name":"new-name"}`),
		Response: testnet.TestResponse{Status: http.StatusCreated},
	})

	testServer, handler, repo := createServiceRepo([]testnet.TestRequest{req})
	defer testServer.Close()

	apiErr := repo.RenameService(serviceInstance, "new-name")
	Expect(handler).To(testnet.HaveAllRequestsCalled())
	Expect(apiErr).NotTo(HaveOccurred())
}

func createServiceRepo(reqs []testnet.TestRequest) (testServer *httptest.Server, handler *testnet.TestHandler, repo ServiceRepository) {
	config := testconfig.NewRepository()
	config.SetAccessToken("BEARER my_access_token")
	config.SetSpaceFields(models.SpaceFields{Guid: "my-space-guid"})
	return createServiceRepoWithConfig(reqs, config)
}

func createServiceRepoWithConfig(reqs []testnet.TestRequest, config configuration.ReadWriter) (testServer *httptest.Server, handler *testnet.TestHandler, repo ServiceRepository) {
	testServer, handler = testnet.NewServer(reqs)
	config.SetApiEndpoint(testServer.URL)

	gateway := net.NewCloudControllerGateway(config)
	repo = NewCloudControllerServiceRepository(config, gateway)
	return
}
