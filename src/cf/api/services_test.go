package api_test

import (
	"cf"
	. "cf/api"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testhelpers"
	"testing"
)

var multipleOfferingsResponse = testhelpers.TestResponse{Status: http.StatusOK, Body: `
{
  "resources": [
    {
      "metadata": {
        "guid": "offering-1-guid"
      },
      "entity": {
        "label": "Offering 1",
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

var multipleOfferingsEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/services?inline-relations-depth=1",
	nil,
	multipleOfferingsResponse,
)

func TestGetServiceOfferings(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleOfferingsEndpoint))
	defer ts.Close()

	repo := CloudControllerServiceRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	offerings, err := repo.GetServiceOfferings(config)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(offerings))

	firstOffering := offerings[0]
	assert.Equal(t, firstOffering.Label, "Offering 1")
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

var createServiceInstanceEndpoint = testhelpers.CreateEndpoint(
	"POST",
	"/v2/service_instances",
	testhelpers.RequestBodyMatcher(`{"name":"instance-name","service_plan_guid":"plan-guid","space_guid":"space-guid"}`),
	testhelpers.TestResponse{Status: http.StatusCreated},
)

func TestCreateServiceInstance(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(createServiceInstanceEndpoint))
	defer ts.Close()

	repo := CloudControllerServiceRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "space-guid"},
	}

	err := repo.CreateServiceInstance(config, "instance-name", cf.ServicePlan{Guid: "plan-guid"})
	assert.NoError(t, err)
}

var singleServiceInstanceResponse = testhelpers.TestResponse{Status: http.StatusOK, Body: `{
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
        ]
      }
    }
  ]
}`}

var findServiceInstanceEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/spaces/my-space-guid/service_instances?q=name%3Amy-service",
	nil,
	singleServiceInstanceResponse,
)

func TestFindInstanceByName(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(findServiceInstanceEndpoint))
	defer ts.Close()

	repo := CloudControllerServiceRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
	}

	instance, err := repo.FindInstanceByName(config, "my-service")
	assert.NoError(t, err)
	assert.Equal(t, instance.Name, "my-service")
	assert.Equal(t, instance.Guid, "my-service-instance-guid")
	assert.Equal(t, len(instance.ServiceBindings), 2)

	binding := instance.ServiceBindings[0]
	assert.Equal(t, binding.Url, "/v2/service_bindings/service-binding-1-guid")
	assert.Equal(t, binding.Guid, "service-binding-1-guid")
	assert.Equal(t, binding.AppGuid, "app-1-guid")
}

var bindServiceEndpoint = testhelpers.CreateEndpoint(
	"POST",
	"/v2/service_bindings",
	testhelpers.RequestBodyMatcher(`{"app_guid":"my-app-guid","service_instance_guid":"my-service-instance-guid"}`),
	testhelpers.TestResponse{Status: http.StatusCreated},
)

func TestBindService(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(bindServiceEndpoint))
	defer ts.Close()

	repo := CloudControllerServiceRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}

	serviceInstance := cf.ServiceInstance{Guid: "my-service-instance-guid"}
	app := cf.Application{Guid: "my-app-guid"}
	err := repo.BindService(config, serviceInstance, app)
	assert.NoError(t, err)
}

var deleteBindingEndpoint = testhelpers.CreateEndpoint(
	"DELETE",
	"/v2/service_bindings/service-binding-2-guid",
	nil,
	testhelpers.TestResponse{Status: http.StatusOK},
)

func TestUnbindService(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(deleteBindingEndpoint))
	defer ts.Close()

	repo := CloudControllerServiceRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}

	serviceBindings := []cf.ServiceBinding{
		cf.ServiceBinding{Url: "/v2/service_bindings/service-binding-1-guid", AppGuid: "app-1-guid"},
		cf.ServiceBinding{Url: "/v2/service_bindings/service-binding-2-guid", AppGuid: "app-2-guid"},
	}

	serviceInstance := cf.ServiceInstance{
		Guid:            "my-service-instance-guid",
		ServiceBindings: serviceBindings,
	}
	app := cf.Application{Guid: "app-2-guid"}
	err := repo.UnbindService(config, serviceInstance, app)
	assert.NoError(t, err)
}

var deleteServiceInstanceEndpoint = testhelpers.CreateEndpoint(
	"DELETE",
	"/v2/spaces/my-space-guid/service_instances/my-service-instance-guid",
	nil,
	testhelpers.TestResponse{Status: http.StatusOK},
)

func TestDeleteServiceWithoutServiceBindings(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(deleteServiceInstanceEndpoint))
	defer ts.Close()

	repo := CloudControllerServiceRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}

	serviceInstance := cf.ServiceInstance{ Guid:"my-service-instance-guid"}
	err := repo.DeleteService(config, serviceInstance)
	assert.NoError(t, err)
}

func TestDeleteServiceWithServiceBindings(t *testing.T) {
	repo := CloudControllerServiceRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
	}

	serviceBindings := []cf.ServiceBinding{
		cf.ServiceBinding{Url: "/v2/service_bindings/service-binding-1-guid", AppGuid: "app-1-guid"},
		cf.ServiceBinding{Url: "/v2/service_bindings/service-binding-2-guid", AppGuid: "app-2-guid"},
	}

	serviceInstance := cf.ServiceInstance{
		Guid:            "my-service-instance-guid",
		ServiceBindings: serviceBindings,
	}

	err := repo.DeleteService(config, serviceInstance)
	assert.Equal(t, err.Error(), "Cannot delete service instance, apps are still bound to it")
}
