package api_test

import (
	. "cf/api"
	"cf/configuration"
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

func testGetServiceOfferings(req testnet.TestRequest, config configuration.ReadWriter) {
	ts, handler, repo := createServiceRepoWithConfig([]testnet.TestRequest{req}, config)
	defer ts.Close()

	offerings, apiResponse := repo.GetServiceOfferings()

	Expect(handler.AllRequestsCalled()).To(BeTrue())
	Expect(apiResponse.IsNotSuccessful()).To(BeFalse())
	Expect(2).To(Equal(len(offerings)))

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

var findServiceInstanceReq = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/spaces/my-space-guid/service_instances?return_user_provided_service_instances=true&q=name%3Amy-service",
	Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": [
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
				"service": {
				  "metadata": {
					"guid": "service-guid"
				  },
				  "entity": {
					"label": "mysql",
					"description": "MySQL database",
					"documentation_url": "http://info.example.com"
				  }
				}
			  }
			}
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

	ts, handler, repo := createServiceRepo([]testnet.TestRequest{req})
	defer ts.Close()

	apiResponse := repo.RenameService(serviceInstance, "new-name")
	Expect(handler.AllRequestsCalled()).To(BeTrue())
	Expect(apiResponse.IsNotSuccessful()).To(BeFalse())
}

func createServiceRepo(reqs []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo ServiceRepository) {
	space := models.SpaceFields{}
	space.Guid = "my-space-guid"
	config := testconfig.NewRepository()
	config.SetAccessToken("BEARER my_access_token")
	config.SetSpaceFields(space)
	return createServiceRepoWithConfig(reqs, config)
}

func createServiceRepoWithConfig(reqs []testnet.TestRequest, config configuration.ReadWriter) (ts *httptest.Server, handler *testnet.TestHandler, repo ServiceRepository) {
	if len(reqs) > 0 {
		ts, handler = testnet.NewTLSServer(reqs)
		config.SetApiEndpoint(ts.URL)
	}

	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerServiceRepository(config, gateway)
	return
}

var _ = Describe("Testing with ginkgo", func() {

	It("TestGetServiceOfferingsWhenNotTargetingASpace", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "GET",
			Path:     "/v2/services?inline-relations-depth=1",
			Response: multipleOfferingsResponse,
		})

		config := testconfig.NewRepository()
		config.SetAccessToken("BEARER my_access_token")

		testGetServiceOfferings(req, config)
	})

	It("TestGetServiceOfferingsWhenTargetingASpace", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "GET",
			Path:     "/v2/spaces/my-space-guid/services?inline-relations-depth=1",
			Response: multipleOfferingsResponse,
		})

		space := models.SpaceFields{}
		space.Guid = "my-space-guid"

		config := testconfig.NewRepository()
		config.SetAccessToken("BEARER my_access_token")
		config.SetSpaceFields(space)

		testGetServiceOfferings(req, config)
	})
	It("TestCreateServiceInstance", func() {

		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "POST",
			Path:     "/v2/service_instances",
			Matcher:  testnet.RequestBodyMatcher(`{"name":"instance-name","service_plan_guid":"plan-guid","space_guid":"my-space-guid","async":true}`),
			Response: testnet.TestResponse{Status: http.StatusCreated},
		})

		ts, handler, repo := createServiceRepo([]testnet.TestRequest{req})
		defer ts.Close()

		identicalAlreadyExists, apiResponse := repo.CreateServiceInstance("instance-name", "plan-guid")
		Expect(handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsSuccessful()).To(BeTrue())
		Expect(identicalAlreadyExists).To(Equal(false))
	})
	It("TestCreateServiceInstanceWhenIdenticalServiceAlreadyExists", func() {

		errorReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:  "POST",
			Path:    "/v2/service_instances",
			Matcher: testnet.RequestBodyMatcher(`{"name":"my-service","service_plan_guid":"plan-guid","space_guid":"my-space-guid","async":true}`),
			Response: testnet.TestResponse{
				Status: http.StatusBadRequest,
				Body:   `{"code":60002,"description":"The service instance name is taken: my-service"}`,
			},
		})

		ts, handler, repo := createServiceRepo([]testnet.TestRequest{errorReq, findServiceInstanceReq})
		defer ts.Close()

		identicalAlreadyExists, apiResponse := repo.CreateServiceInstance("my-service", "plan-guid")

		Expect(handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsNotSuccessful()).To(BeFalse())
		Expect(identicalAlreadyExists).To(Equal(true))
	})

	It("TestCreateServiceInstanceWhenDifferentServiceAlreadyExists", func() {
		errorReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:  "POST",
			Path:    "/v2/service_instances",
			Matcher: testnet.RequestBodyMatcher(`{"name":"my-service","service_plan_guid":"different-plan-guid","space_guid":"my-space-guid","async":true}`),
			Response: testnet.TestResponse{
				Status: http.StatusBadRequest,
				Body:   `{"code":60002,"description":"The service instance name is taken: my-service"}`,
			},
		})

		ts, handler, repo := createServiceRepo([]testnet.TestRequest{errorReq, findServiceInstanceReq})
		defer ts.Close()

		identicalAlreadyExists, apiResponse := repo.CreateServiceInstance("my-service", "different-plan-guid")

		Expect(handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsNotSuccessful()).To(BeTrue())
		Expect(identicalAlreadyExists).To(Equal(false))
	})

	It("TestFindInstanceByName", func() {
		ts, handler, repo := createServiceRepo([]testnet.TestRequest{findServiceInstanceReq})
		defer ts.Close()

		instance, apiResponse := repo.FindInstanceByName("my-service")

		Expect(handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsNotSuccessful()).To(BeFalse())
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

	It("TestFindInstanceByNameForNonExistentService", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "GET",
			Path:     "/v2/spaces/my-space-guid/service_instances?return_user_provided_service_instances=true&q=name%3Amy-service",
			Response: testnet.TestResponse{Status: http.StatusOK, Body: `{ "resources": [] }`},
		})

		ts, handler, repo := createServiceRepo([]testnet.TestRequest{req})
		defer ts.Close()

		_, apiResponse := repo.FindInstanceByName("my-service")
		Expect(handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsError()).To(BeFalse())
		Expect(apiResponse.IsNotFound()).To(BeTrue())
	})

	It("TestDeleteServiceWithoutServiceBindings", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "DELETE",
			Path:     "/v2/service_instances/my-service-instance-guid",
			Response: testnet.TestResponse{Status: http.StatusOK},
		})
		ts, handler, repo := createServiceRepo([]testnet.TestRequest{req})
		defer ts.Close()
		serviceInstance := models.ServiceInstance{}
		serviceInstance.Guid = "my-service-instance-guid"
		apiResponse := repo.DeleteService(serviceInstance)
		Expect(handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsNotSuccessful()).To(BeFalse())
	})

	It("TestDeleteServiceWithServiceBindings", func() {
		_, _, repo := createServiceRepo([]testnet.TestRequest{})

		serviceInstance := models.ServiceInstance{}
		serviceInstance.Guid = "my-service-instance-guid"

		binding := models.ServiceBindingFields{}
		binding.Url = "/v2/service_bindings/service-binding-1-guid"
		binding.AppGuid = "app-1-guid"

		binding2 := models.ServiceBindingFields{}
		binding2.Url = "/v2/service_bindings/service-binding-2-guid"
		binding2.AppGuid = "app-2-guid"

		serviceInstance.ServiceBindings = []models.ServiceBindingFields{binding, binding2}

		apiResponse := repo.DeleteService(serviceInstance)
		Expect(apiResponse.IsNotSuccessful()).To(BeTrue())
		Expect(apiResponse.Message).To(Equal("Cannot delete service instance, apps are still bound to it"))
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
		_, _, repo := createServiceRepo([]testnet.TestRequest{{
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

		offering, apiResponse := repo.FindServiceOfferingByLabelAndProvider("offering-1", "provider-1")
		Expect(offering.Guid).To(Equal("offering-1-guid"))
		Expect(apiResponse.IsSuccessful()).To(BeTrue())
	})

	It("returns an error if the offering cannot be found", func() {
		_, _, repo := createServiceRepo([]testnet.TestRequest{{
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

		offering, apiResponse := repo.FindServiceOfferingByLabelAndProvider("offering-1", "provider-1")
		Expect(apiResponse.IsNotFound()).To(BeTrue())
		Expect(offering.Guid).To(Equal(""))
	})

	It("handles api errors when finding service offerings", func() {
		_, _, repo := createServiceRepo([]testnet.TestRequest{{
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

		_, apiResponse := repo.FindServiceOfferingByLabelAndProvider("offering-1", "provider-1")
		Expect(apiResponse.IsError()).To(BeTrue())
		Expect(apiResponse.ErrorCode).To(Equal("10005"))
	})

	It("purges service offerings", func() {
		_, handler, repo := createServiceRepo([]testnet.TestRequest{{
			Method: "DELETE",
			Path:   "/v2/services/the-service-guid?purge=true",
			Response: testnet.TestResponse{
				Status: 204,
			},
		}})

		offering := maker.NewServiceOffering("the-offering")
		offering.Guid = "the-service-guid"

		apiResponse := repo.PurgeServiceOffering(offering)
		Expect(apiResponse.IsSuccessful()).To(BeTrue())
		Expect(handler.AllRequestsCalled()).To(BeTrue())
	})

	Describe("migrating service plans", func() {
		It("returns an error when it cannot find a service plan by name", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/service_plans",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{ "resources": [] }`},
			})

			ts, _, repo := createServiceRepo([]testnet.TestRequest{req})
			defer ts.Close()

			v1 := V1ServicePlanDescription{}
			v2 := V2ServicePlanDescription{}
			_, _, apiResponse := repo.FindServicePlanToMigrateByDescription(v1, v2)
			Expect(apiResponse.IsNotSuccessful()).To(BeTrue())
			Expect(apiResponse.Message).To(ContainSubstring("not found"))
		})

		It("returns the service plan guid when finding service plan by name", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/service_plans",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources":[{"metadata": {"guid": "v1-guid"},"entity": {"name": "v1-panda","service": {"entity": {"label": "v1-elephantsql","provider": "v1-elephantsql"}}}}, {"metadata": {"guid": "v2-guid"},"entity": {"name": "v2-panda","service": {"entity": {"label": "v2-elephantsql","provider": null}}}}]}`},
			})

			ts, _, repo := createServiceRepo([]testnet.TestRequest{req})
			defer ts.Close()

			v1 := V1ServicePlanDescription{
				ServiceName:     "v1-elephantsql",
				ServicePlanName: "v1-panda",
				ServiceProvider: "v1-elephantsql",
			}
			v2 := V2ServicePlanDescription{
				ServiceName:     "v2-elephantsql",
				ServicePlanName: "v2-panda",
			}
			v1Guid, v2Guid, apiResponse := repo.FindServicePlanToMigrateByDescription(v1, v2)

			Expect(apiResponse.IsSuccessful()).To(BeTrue())
			Expect(v1Guid).To(Equal("v1-guid"))
			Expect(v2Guid).To(Equal("v2-guid"))
		})

		It("migrates service plans from v1 to v2", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/service_plans/v1-guid/service_instances",
				Matcher:  testnet.RequestBodyMatcher(`{"service_plan_guid":"v2-guid"}`),
				Response: testnet.TestResponse{Status: http.StatusOK},
			})

			ts, _, repo := createServiceRepo([]testnet.TestRequest{req})
			defer ts.Close()

			apiResponse := repo.MigrateServicePlanFromV1ToV2("v1-guid", "v2-guid")
			Expect(apiResponse.IsSuccessful()).To(BeTrue())
		})

		It("returns an error when migrating fails", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/service_plans/v1-guid/service_instances",
				Matcher:  testnet.RequestBodyMatcher(`{"service_plan_guid":"v2-guid"}`),
				Response: testnet.TestResponse{Status: http.StatusInternalServerError},
			})

			ts, _, repo := createServiceRepo([]testnet.TestRequest{req})
			defer ts.Close()

			apiResponse := repo.MigrateServicePlanFromV1ToV2("v1-guid", "v2-guid")
			Expect(apiResponse.IsSuccessful()).To(BeFalse())
		})
	})
})
