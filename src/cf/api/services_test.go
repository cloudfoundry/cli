package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	testapi "testhelpers/api"
	"testing"
)

var multipleOfferingsResponse = testapi.TestResponse{Status: http.StatusOK, Body: `
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

func testGetServiceOfferings(t *testing.T, endpoint http.HandlerFunc, status *testapi.RequestStatus, config *configuration.Configuration) {
	ts, repo := createServiceRepoWithConfig(endpoint, config)
	defer ts.Close()

	offerings, apiResponse := repo.GetServiceOfferings()

	assert.True(t, status.Called())
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

func TestGetServiceOfferingsWhenNotTargetingASpace(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"GET",
		"/v2/services?inline-relations-depth=1",
		nil,
		multipleOfferingsResponse,
	)

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
	}
	testGetServiceOfferings(t, endpoint, status, config)
}

func TestGetServiceOfferingsWhenTargetingASpace(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"GET",
		"/v2/spaces/my-space-guid/services?inline-relations-depth=1",
		nil,
		multipleOfferingsResponse,
	)

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Space:       cf.Space{Guid: "my-space-guid"},
	}
	testGetServiceOfferings(t, endpoint, status, config)
}

func TestCreateServiceInstance(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"POST",
		"/v2/service_instances",
		testapi.RequestBodyMatcher(`{"name":"instance-name","service_plan_guid":"plan-guid","space_guid":"my-space-guid"}`),
		testapi.TestResponse{Status: http.StatusCreated},
	)

	ts, repo := createServiceRepo(endpoint)
	defer ts.Close()

	identicalAlreadyExists, apiResponse := repo.CreateServiceInstance("instance-name", cf.ServicePlan{Guid: "plan-guid"})
	assert.True(t, status.Called())
	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, identicalAlreadyExists, false)
}

func TestCreateServiceInstanceWhenIdenticalServiceAlreadyExists(t *testing.T) {
	findServiceInstanceEndpointStatus.Reset()

	errorEndpoint, errorEndpointStatus := testapi.CreateCheckableEndpoint(
		"POST",
		"/v2/service_instances",
		testapi.RequestBodyMatcher(`{"name":"my-service","service_plan_guid":"plan-guid","space_guid":"my-space-guid"}`),
		testapi.TestResponse{
			Status: http.StatusBadRequest,
			Body:   `{"code":60002,"description":"The service instance name is taken: my-service"}`,
		},
	)

	endpoints := func(res http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.RequestURI, "/v2/service_instances") {
			errorEndpoint(res, req)
		} else {
			findServiceInstanceEndpoint(res, req)
		}
	}

	ts, repo := createServiceRepo(http.HandlerFunc(endpoints))
	defer ts.Close()

	servicePlan := cf.ServicePlan{Guid: "plan-guid", Name: "plan-name"}
	identicalAlreadyExists, apiResponse := repo.CreateServiceInstance("my-service", servicePlan)

	assert.True(t, findServiceInstanceEndpointStatus.Called())
	assert.True(t, errorEndpointStatus.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, identicalAlreadyExists, true)
}

func TestCreateServiceInstanceWhenDifferentServiceAlreadyExists(t *testing.T) {
	findServiceInstanceEndpointStatus.Reset()

	errorEndpoint, errorEndpointStatus := testapi.CreateCheckableEndpoint(
		"POST",
		"/v2/service_instances",
		testapi.RequestBodyMatcher(`{"name":"my-service","service_plan_guid":"different-plan-guid","space_guid":"my-space-guid"}`),
		testapi.TestResponse{
			Status: http.StatusBadRequest,
			Body:   `{"code":60002,"description":"The service instance name is taken: my-service"}`,
		},
	)

	endpoints := func(res http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.RequestURI, "/v2/service_instances") {
			errorEndpoint(res, req)
		} else {
			findServiceInstanceEndpoint(res, req)
		}
	}

	ts, repo := createServiceRepo(http.HandlerFunc(endpoints))
	defer ts.Close()

	servicePlan := cf.ServicePlan{Guid: "different-plan-guid", Name: "plan-name"}
	identicalAlreadyExists, apiResponse := repo.CreateServiceInstance("my-service", servicePlan)

	assert.True(t, findServiceInstanceEndpointStatus.Called())
	assert.True(t, errorEndpointStatus.Called())
	assert.True(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, identicalAlreadyExists, false)
}

var singleServiceInstanceResponse = testapi.TestResponse{Status: http.StatusOK, Body: `{
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
  ]
}`}

var findServiceInstanceEndpoint, findServiceInstanceEndpointStatus = testapi.CreateCheckableEndpoint(
	"GET",
	"/v2/spaces/my-space-guid/service_instances?return_user_provided_service_instances=true&q=name%3Amy-service",
	nil,
	singleServiceInstanceResponse,
)

func TestFindInstanceByName(t *testing.T) {
	findServiceInstanceEndpointStatus.Reset()
	ts, repo := createServiceRepo(findServiceInstanceEndpoint)
	defer ts.Close()

	instance, apiResponse := repo.FindInstanceByName("my-service")

	assert.True(t, findServiceInstanceEndpointStatus.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, instance.Name, "my-service")
	assert.Equal(t, instance.Guid, "my-service-instance-guid")
	assert.Equal(t, instance.ServiceOffering().Label, "mysql")
	assert.Equal(t, instance.ServiceOffering().DocumentationUrl, "http://info.example.com")
	assert.Equal(t, instance.ServiceOffering().Description, "MySQL database")
	assert.Equal(t, instance.ServicePlan.Name, "plan-name")
	assert.Equal(t, len(instance.ServiceBindings), 2)

	binding := instance.ServiceBindings[0]
	assert.Equal(t, binding.Url, "/v2/service_bindings/service-binding-1-guid")
	assert.Equal(t, binding.Guid, "service-binding-1-guid")
	assert.Equal(t, binding.AppGuid, "app-1-guid")
}

func TestFindInstanceByNameForNonExistentService(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"GET",
		"/v2/spaces/my-space-guid/service_instances?return_user_provided_service_instances=true&q=name%3Amy-service",
		nil,
		testapi.TestResponse{Status: http.StatusOK, Body: `{ "resources": [] }`},
	)

	ts, repo := createServiceRepo(endpoint)
	defer ts.Close()

	_, apiResponse := repo.FindInstanceByName("my-service")
	assert.True(t, status.Called())
	assert.False(t, apiResponse.IsError())
	assert.True(t, apiResponse.IsNotFound())
}

func TestDeleteServiceWithoutServiceBindings(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"DELETE",
		"/v2/service_instances/my-service-instance-guid",
		nil,
		testapi.TestResponse{Status: http.StatusOK},
	)

	ts, repo := createServiceRepo(endpoint)
	defer ts.Close()

	serviceInstance := cf.ServiceInstance{Guid: "my-service-instance-guid"}
	apiResponse := repo.DeleteService(serviceInstance)
	assert.True(t, status.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestDeleteServiceWithServiceBindings(t *testing.T) {
	_, repo := createServiceRepo(nil)

	serviceBindings := []cf.ServiceBinding{
		cf.ServiceBinding{Url: "/v2/service_bindings/service-binding-1-guid", AppGuid: "app-1-guid"},
		cf.ServiceBinding{Url: "/v2/service_bindings/service-binding-2-guid", AppGuid: "app-2-guid"},
	}

	serviceInstance := cf.ServiceInstance{
		Guid:            "my-service-instance-guid",
		ServiceBindings: serviceBindings,
	}

	apiResponse := repo.DeleteService(serviceInstance)
	assert.True(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, apiResponse.Message, "Cannot delete service instance, apps are still bound to it")
}

func TestRenameService(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"PUT",
		"/v2/service_instances/my-service-instance-guid",
		testapi.RequestBodyMatcher(`{"name":"new-name"}`),
		testapi.TestResponse{Status: http.StatusCreated},
	)

	ts, repo := createServiceRepo(endpoint)
	defer ts.Close()

	serviceInstance := cf.ServiceInstance{Guid: "my-service-instance-guid"}
	apiResponse := repo.RenameService(serviceInstance, "new-name")
	assert.True(t, status.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func createServiceRepo(endpoint http.HandlerFunc) (ts *httptest.Server, repo ServiceRepository) {
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Space:       cf.Space{Guid: "my-space-guid"},
	}
	return createServiceRepoWithConfig(endpoint, config)
}

func createServiceRepoWithConfig(endpoint http.HandlerFunc, config *configuration.Configuration) (ts *httptest.Server, repo ServiceRepository) {
	if endpoint != nil {
		ts = httptest.NewTLSServer(endpoint)
		config.Target = ts.URL
	}

	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerServiceRepository(config, gateway)
	return
}
