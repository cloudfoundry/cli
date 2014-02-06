package api_test

import (
	. "cf/api"
	"cf/configuration"
	"cf/models"
	"cf/net"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
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

func testGetServiceOfferings(t mr.TestingT, req testnet.TestRequest, config *configuration.Configuration) {
	ts, handler, repo := createServiceRepoWithConfig(t, []testnet.TestRequest{req}, config)
	defer ts.Close()

	offerings, apiResponse := repo.GetServiceOfferings()

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, 2, len(offerings))

	firstOffering := offerings[0]
	assert.Equal(t, firstOffering.Label, "Offering 1")
	assert.Equal(t, firstOffering.Version, "1.0")
	assert.Equal(t, firstOffering.Description, "Offering 1 description")
	assert.Equal(t, firstOffering.Provider, "Offering 1 provider")
	assert.Equal(t, firstOffering.Guid, "offering-1-guid")
	assert.Equal(t, len(firstOffering.Plans), 2)

	plan := firstOffering.Plans[0]
	assert.Equal(t, plan.Name, "Offering 1 Plan 1")
	assert.Equal(t, plan.Guid, "offering-1-plan-1-guid")

	secondOffering := offerings[1]
	assert.Equal(t, secondOffering.Label, "Offering 2")
	assert.Equal(t, secondOffering.Guid, "offering-2-guid")
	assert.Equal(t, len(secondOffering.Plans), 1)
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

func testRenameService(t mr.TestingT, endpointPath string, serviceInstance models.ServiceInstance) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     endpointPath,
		Matcher:  testnet.RequestBodyMatcher(`{"name":"new-name"}`),
		Response: testnet.TestResponse{Status: http.StatusCreated},
	})

	ts, handler, repo := createServiceRepo(t, []testnet.TestRequest{req})
	defer ts.Close()

	apiResponse := repo.RenameService(serviceInstance, "new-name")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func createServiceRepo(t mr.TestingT, reqs []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo ServiceRepository) {
	space2 := models.SpaceFields{}
	space2.Guid = "my-space-guid"
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		SpaceFields: space2,
	}
	return createServiceRepoWithConfig(t, reqs, config)
}

func createServiceRepoWithConfig(t mr.TestingT, reqs []testnet.TestRequest, config *configuration.Configuration) (ts *httptest.Server, handler *testnet.TestHandler, repo ServiceRepository) {
	if len(reqs) > 0 {
		ts, handler = testnet.NewTLSServer(t, reqs)
		config.Target = ts.URL
	}

	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerServiceRepository(config, gateway)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestGetServiceOfferingsWhenNotTargetingASpace", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/services?inline-relations-depth=1",
				Response: multipleOfferingsResponse,
			})

			config := &configuration.Configuration{
				AccessToken: "BEARER my_access_token",
			}
			testGetServiceOfferings(mr.T(), req, config)
		})
		It("TestGetServiceOfferingsWhenTargetingASpace", func() {

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/spaces/my-space-guid/services?inline-relations-depth=1",
				Response: multipleOfferingsResponse,
			})
			space := models.SpaceFields{}
			space.Guid = "my-space-guid"
			config := &configuration.Configuration{
				AccessToken: "BEARER my_access_token",
				SpaceFields: space,
			}
			testGetServiceOfferings(mr.T(), req, config)
		})
		It("TestCreateServiceInstance", func() {

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "POST",
				Path:     "/v2/service_instances",
				Matcher:  testnet.RequestBodyMatcher(`{"name":"instance-name","service_plan_guid":"plan-guid","space_guid":"my-space-guid","async":true}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			ts, handler, repo := createServiceRepo(mr.T(), []testnet.TestRequest{req})
			defer ts.Close()

			identicalAlreadyExists, apiResponse := repo.CreateServiceInstance("instance-name", "plan-guid")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())
			assert.Equal(mr.T(), identicalAlreadyExists, false)
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

			ts, handler, repo := createServiceRepo(mr.T(), []testnet.TestRequest{errorReq, findServiceInstanceReq})
			defer ts.Close()

			identicalAlreadyExists, apiResponse := repo.CreateServiceInstance("my-service", "plan-guid")

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
			assert.Equal(mr.T(), identicalAlreadyExists, true)
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

			ts, handler, repo := createServiceRepo(mr.T(), []testnet.TestRequest{errorReq, findServiceInstanceReq})
			defer ts.Close()

			identicalAlreadyExists, apiResponse := repo.CreateServiceInstance("my-service", "different-plan-guid")

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsNotSuccessful())
			assert.Equal(mr.T(), identicalAlreadyExists, false)
		})
		It("TestFindInstanceByName", func() {

			ts, handler, repo := createServiceRepo(mr.T(), []testnet.TestRequest{findServiceInstanceReq})
			defer ts.Close()

			instance, apiResponse := repo.FindInstanceByName("my-service")

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
			assert.Equal(mr.T(), instance.Name, "my-service")
			assert.Equal(mr.T(), instance.Guid, "my-service-instance-guid")
			assert.Equal(mr.T(), instance.ServiceOffering.Label, "mysql")
			assert.Equal(mr.T(), instance.ServiceOffering.DocumentationUrl, "http://info.example.com")
			assert.Equal(mr.T(), instance.ServiceOffering.Description, "MySQL database")
			assert.Equal(mr.T(), instance.ServicePlan.Name, "plan-name")
			assert.Equal(mr.T(), len(instance.ServiceBindings), 2)

			binding := instance.ServiceBindings[0]
			assert.Equal(mr.T(), binding.Url, "/v2/service_bindings/service-binding-1-guid")
			assert.Equal(mr.T(), binding.Guid, "service-binding-1-guid")
			assert.Equal(mr.T(), binding.AppGuid, "app-1-guid")
		})
		It("TestFindInstanceByNameForNonExistentService", func() {

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/spaces/my-space-guid/service_instances?return_user_provided_service_instances=true&q=name%3Amy-service",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{ "resources": [] }`},
			})

			ts, handler, repo := createServiceRepo(mr.T(), []testnet.TestRequest{req})
			defer ts.Close()

			_, apiResponse := repo.FindInstanceByName("my-service")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsError())
			assert.True(mr.T(), apiResponse.IsNotFound())
		})
		It("TestDeleteServiceWithoutServiceBindings", func() {

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "DELETE",
				Path:     "/v2/service_instances/my-service-instance-guid",
				Response: testnet.TestResponse{Status: http.StatusOK},
			})

			ts, handler, repo := createServiceRepo(mr.T(), []testnet.TestRequest{req})
			defer ts.Close()
			serviceInstance := models.ServiceInstance{}
			serviceInstance.Guid = "my-service-instance-guid"
			apiResponse := repo.DeleteService(serviceInstance)
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
		})
		It("TestDeleteServiceWithServiceBindings", func() {

			_, _, repo := createServiceRepo(mr.T(), []testnet.TestRequest{})

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
			assert.True(mr.T(), apiResponse.IsNotSuccessful())
			assert.Equal(mr.T(), apiResponse.Message, "Cannot delete service instance, apps are still bound to it")
		})
		It("TestRenameService", func() {

			path := "/v2/service_instances/my-service-instance-guid"
			serviceInstance := models.ServiceInstance{}
			serviceInstance.Guid = "my-service-instance-guid"

			plan := models.ServicePlanFields{}
			plan.Guid = "some-plan-guid"
			serviceInstance.ServicePlan = plan

			testRenameService(mr.T(), path, serviceInstance)
		})
		It("TestRenameServiceWhenServiceIsUserProvided", func() {

			path := "/v2/user_provided_service_instances/my-service-instance-guid"
			serviceInstance := models.ServiceInstance{}
			serviceInstance.Guid = "my-service-instance-guid"
			testRenameService(mr.T(), path, serviceInstance)
		})
	})
}
