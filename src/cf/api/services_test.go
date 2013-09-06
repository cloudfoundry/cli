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

var multipleOfferingsEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/services?inline-relations-depth=1",
	nil,
	multipleOfferingsResponse,
)

func TestGetServiceOfferings(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleOfferingsEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	repo := NewCloudControllerServiceRepository(config)
	offerings, err := repo.GetServiceOfferings()

	assert.NoError(t, err)
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

var createServiceInstanceEndpoint = testhelpers.CreateEndpoint(
	"POST",
	"/v2/service_instances",
	testhelpers.RequestBodyMatcher(`{"name":"instance-name","service_plan_guid":"plan-guid","space_guid":"space-guid"}`),
	testhelpers.TestResponse{Status: http.StatusCreated},
)

func TestCreateServiceInstance(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(createServiceInstanceEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "space-guid"},
	}
	repo := NewCloudControllerServiceRepository(config)

	err := repo.CreateServiceInstance("instance-name", cf.ServicePlan{Guid: "plan-guid"})
	assert.NoError(t, err)
}

var createUserProvidedServiceInstanceEndpoint = testhelpers.CreateEndpoint(
	"POST",
	"/v2/user_provided_service_instances",
	testhelpers.RequestBodyMatcher(`{"name":"my-custom-service","credentials":{"host":"example.com","password":"secret","user":"me"},"space_guid":"some-space-guid"}`),
	testhelpers.TestResponse{Status: http.StatusCreated},
)

func TestCreateUserProvidedServiceInstance(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(createUserProvidedServiceInstanceEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "some-space-guid"},
	}
	repo := NewCloudControllerServiceRepository(config)

	params := map[string]string{
		"host":     "example.com",
		"user":     "me",
		"password": "secret",
	}
	err := repo.CreateUserProvidedServiceInstance("my-custom-service", params)
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
	"/v2/spaces/my-space-guid/service_instances?return_user_provided_service_instances=true&q=name%3Amy-service",
	nil,
	singleServiceInstanceResponse,
)

func TestFindInstanceByName(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(findServiceInstanceEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
	}
	repo := NewCloudControllerServiceRepository(config)

	instance, err := repo.FindInstanceByName("my-service")
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

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	repo := NewCloudControllerServiceRepository(config)

	serviceInstance := cf.ServiceInstance{Guid: "my-service-instance-guid"}
	app := cf.Application{Guid: "my-app-guid"}
	_, err := repo.BindService(serviceInstance, app)
	assert.NoError(t, err)
}

var bindServiceErrorEndpoint = testhelpers.CreateEndpoint(
	"POST",
	"/v2/service_bindings",
	testhelpers.RequestBodyMatcher(`{"app_guid":"my-app-guid","service_instance_guid":"my-service-instance-guid"}`),
	testhelpers.TestResponse{
		Status: http.StatusBadRequest,
		Body:   `{"code":90003,"description":"The app space binding to service is taken: 7b959018-110a-4913-ac0a-d663e613cdea 346bf237-7eef-41a7-b892-68fb08068f09"}`,
	},
)

func TestBindServiceIfError(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(bindServiceErrorEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	repo := NewCloudControllerServiceRepository(config)

	serviceInstance := cf.ServiceInstance{Guid: "my-service-instance-guid"}
	app := cf.Application{Guid: "my-app-guid"}
	errorCode, err := repo.BindService(serviceInstance, app)

	assert.Error(t, err)
	assert.Equal(t, errorCode, 90003)
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

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	repo := NewCloudControllerServiceRepository(config)

	serviceBindings := []cf.ServiceBinding{
		cf.ServiceBinding{Url: "/v2/service_bindings/service-binding-1-guid", AppGuid: "app-1-guid"},
		cf.ServiceBinding{Url: "/v2/service_bindings/service-binding-2-guid", AppGuid: "app-2-guid"},
	}

	serviceInstance := cf.ServiceInstance{
		Guid:            "my-service-instance-guid",
		ServiceBindings: serviceBindings,
	}
	app := cf.Application{Guid: "app-2-guid"}
	err := repo.UnbindService(serviceInstance, app)
	assert.NoError(t, err)
}

var deleteServiceInstanceEndpoint = testhelpers.CreateEndpoint(
	"DELETE",
	"/v2/service_instances/my-service-instance-guid",
	nil,
	testhelpers.TestResponse{Status: http.StatusOK},
)

func TestDeleteServiceWithoutServiceBindings(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(deleteServiceInstanceEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	repo := NewCloudControllerServiceRepository(config)

	serviceInstance := cf.ServiceInstance{Guid: "my-service-instance-guid"}
	err := repo.DeleteService(serviceInstance)
	assert.NoError(t, err)
}

func TestDeleteServiceWithServiceBindings(t *testing.T) {
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
	}
	repo := NewCloudControllerServiceRepository(config)

	serviceBindings := []cf.ServiceBinding{
		cf.ServiceBinding{Url: "/v2/service_bindings/service-binding-1-guid", AppGuid: "app-1-guid"},
		cf.ServiceBinding{Url: "/v2/service_bindings/service-binding-2-guid", AppGuid: "app-2-guid"},
	}

	serviceInstance := cf.ServiceInstance{
		Guid:            "my-service-instance-guid",
		ServiceBindings: serviceBindings,
	}

	err := repo.DeleteService(serviceInstance)
	assert.Equal(t, err.Error(), "Cannot delete service instance, apps are still bound to it")
}
