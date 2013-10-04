package api_test

import (
	"cf"
	. "cf/api"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
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
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticationRepository{})
	repo := NewCloudControllerServiceRepository(config, gateway)
	offerings, apiStatus := repo.GetServiceOfferings()

	assert.False(t, apiStatus.NotSuccessful())
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
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticationRepository{})
	repo := NewCloudControllerServiceRepository(config, gateway)

	identicalAlreadyExists, apiStatus := repo.CreateServiceInstance("instance-name", cf.ServicePlan{Guid: "plan-guid"})
	assert.False(t, apiStatus.NotSuccessful())
	assert.Equal(t, identicalAlreadyExists, false)
}

var identicalServiceInstanceAlreadyExistsEndpoint = testhelpers.CreateEndpoint(
	"POST",
	"/v2/service_instances",
	testhelpers.RequestBodyMatcher(`{"name":"my-service","service_plan_guid":"plan-guid","space_guid":"my-space-guid"}`),
	testhelpers.TestResponse{
		Status: http.StatusBadRequest,
		Body:   `{"code":60002,"description":"The service instance name is taken: my-service"}`,
	},
)

var identicalInstanceAlreadyExistsEndpoints = func(res http.ResponseWriter, req *http.Request) {
	if strings.Contains(req.RequestURI, "/v2/service_instances") {
		identicalServiceInstanceAlreadyExistsEndpoint(res, req)
	} else {
		findServiceInstanceEndpoint(res, req)
	}
}

func TestCreateServiceInstanceWhenIdenticalServiceAlreadyExists(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(identicalInstanceAlreadyExistsEndpoints))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
	}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticationRepository{})
	repo := NewCloudControllerServiceRepository(config, gateway)

	servicePlan := cf.ServicePlan{Guid: "plan-guid", Name: "plan-name"}
	identicalAlreadyExists, apiStatus := repo.CreateServiceInstance("my-service", servicePlan)

	assert.False(t, apiStatus.NotSuccessful())
	assert.Equal(t, identicalAlreadyExists, true)
}

var differentServiceInstanceAlreadyExistsEndpoint = testhelpers.CreateEndpoint(
	"POST",
	"/v2/service_instances",
	testhelpers.RequestBodyMatcher(`{"name":"my-service","service_plan_guid":"different-plan-guid","space_guid":"my-space-guid"}`),
	testhelpers.TestResponse{
		Status: http.StatusBadRequest,
		Body:   `{"code":60002,"description":"The service instance name is taken: my-service"}`,
	},
)

var differentInstanceAlreadyExistsEndpoints = func(res http.ResponseWriter, req *http.Request) {
	if strings.Contains(req.RequestURI, "/v2/service_instances") {
		differentServiceInstanceAlreadyExistsEndpoint(res, req)
	} else {
		findServiceInstanceEndpoint(res, req)
	}
}

func TestCreateServiceInstanceWhenDifferentServiceAlreadyExists(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(differentInstanceAlreadyExistsEndpoints))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
	}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticationRepository{})
	repo := NewCloudControllerServiceRepository(config, gateway)

	servicePlan := cf.ServicePlan{Guid: "different-plan-guid", Name: "plan-name"}
	identicalAlreadyExists, apiStatus := repo.CreateServiceInstance("my-service", servicePlan)

	assert.True(t, apiStatus.NotSuccessful())
	assert.Equal(t, identicalAlreadyExists, false)
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
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticationRepository{})
	repo := NewCloudControllerServiceRepository(config, gateway)

	params := map[string]string{
		"host":     "example.com",
		"user":     "me",
		"password": "secret",
	}
	apiStatus := repo.CreateUserProvidedServiceInstance("my-custom-service", params)
	assert.False(t, apiStatus.NotSuccessful())
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
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticationRepository{})
	repo := NewCloudControllerServiceRepository(config, gateway)

	instance, apiStatus := repo.FindInstanceByName("my-service")
	assert.False(t, apiStatus.NotSuccessful())
	assert.Equal(t, instance.Name, "my-service")
	assert.Equal(t, instance.Guid, "my-service-instance-guid")
	assert.Equal(t, instance.ServiceOffering.Label, "mysql")
	assert.Equal(t, instance.ServiceOffering.DocumentationUrl, "http://info.example.com")
	assert.Equal(t, instance.ServiceOffering.Description, "MySQL database")
	assert.Equal(t, instance.ServicePlan.Name, "plan-name")
	assert.Equal(t, len(instance.ServiceBindings), 2)

	binding := instance.ServiceBindings[0]
	assert.Equal(t, binding.Url, "/v2/service_bindings/service-binding-1-guid")
	assert.Equal(t, binding.Guid, "service-binding-1-guid")
	assert.Equal(t, binding.AppGuid, "app-1-guid")
}

var serviceNotFoundResponse = testhelpers.TestResponse{Status: http.StatusOK, Body: `{
  "resources": []
}`}

var serviceNotFoundEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/spaces/my-space-guid/service_instances?return_user_provided_service_instances=true&q=name%3Amy-service",
	nil,
	serviceNotFoundResponse,
)

func TestFindInstanceByNameForNonExistentService(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(serviceNotFoundEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
	}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticationRepository{})
	repo := NewCloudControllerServiceRepository(config, gateway)

	_, apiStatus := repo.FindInstanceByName("my-service")
	assert.False(t, apiStatus.IsError())
	assert.True(t, apiStatus.IsNotFound())
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
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticationRepository{})
	repo := NewCloudControllerServiceRepository(config, gateway)

	serviceInstance := cf.ServiceInstance{Guid: "my-service-instance-guid"}
	app := cf.Application{Guid: "my-app-guid"}
	apiStatus := repo.BindService(serviceInstance, app)
	assert.False(t, apiStatus.NotSuccessful())
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
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticationRepository{})
	repo := NewCloudControllerServiceRepository(config, gateway)

	serviceInstance := cf.ServiceInstance{Guid: "my-service-instance-guid"}
	app := cf.Application{Guid: "my-app-guid"}
	apiStatus := repo.BindService(serviceInstance, app)

	assert.True(t, apiStatus.NotSuccessful())
	assert.Equal(t, apiStatus.ErrorCode, "90003")
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
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticationRepository{})
	repo := NewCloudControllerServiceRepository(config, gateway)

	serviceBindings := []cf.ServiceBinding{
		cf.ServiceBinding{Url: "/v2/service_bindings/service-binding-1-guid", AppGuid: "app-1-guid"},
		cf.ServiceBinding{Url: "/v2/service_bindings/service-binding-2-guid", AppGuid: "app-2-guid"},
	}

	serviceInstance := cf.ServiceInstance{
		Guid:            "my-service-instance-guid",
		ServiceBindings: serviceBindings,
	}
	app := cf.Application{Guid: "app-2-guid"}
	found, apiStatus := repo.UnbindService(serviceInstance, app)
	assert.False(t, apiStatus.NotSuccessful())
	assert.True(t, found)
}

func TestUnbindServiceWhenBindingDoesNotExist(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(deleteBindingEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticationRepository{})
	repo := NewCloudControllerServiceRepository(config, gateway)

	serviceBindings := []cf.ServiceBinding{}

	serviceInstance := cf.ServiceInstance{
		Guid:            "my-service-instance-guid",
		ServiceBindings: serviceBindings,
	}
	app := cf.Application{Guid: "app-2-guid"}
	found, apiStatus := repo.UnbindService(serviceInstance, app)
	assert.False(t, apiStatus.NotSuccessful())
	assert.False(t, found)
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
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticationRepository{})
	repo := NewCloudControllerServiceRepository(config, gateway)

	serviceInstance := cf.ServiceInstance{Guid: "my-service-instance-guid"}
	apiStatus := repo.DeleteService(serviceInstance)
	assert.False(t, apiStatus.NotSuccessful())
}

func TestDeleteServiceWithServiceBindings(t *testing.T) {
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
	}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticationRepository{})
	repo := NewCloudControllerServiceRepository(config, gateway)

	serviceBindings := []cf.ServiceBinding{
		cf.ServiceBinding{Url: "/v2/service_bindings/service-binding-1-guid", AppGuid: "app-1-guid"},
		cf.ServiceBinding{Url: "/v2/service_bindings/service-binding-2-guid", AppGuid: "app-2-guid"},
	}

	serviceInstance := cf.ServiceInstance{
		Guid:            "my-service-instance-guid",
		ServiceBindings: serviceBindings,
	}

	apiStatus := repo.DeleteService(serviceInstance)
	assert.True(t, apiStatus.NotSuccessful())
	assert.Equal(t, apiStatus.Message, "Cannot delete service instance, apps are still bound to it")
}

var renameServiceInstanceEndpoint = testhelpers.CreateEndpoint(
	"PUT",
	"/v2/service_instances/my-service-instance-guid",
	testhelpers.RequestBodyMatcher(`{"name":"new-name"}`),
	testhelpers.TestResponse{Status: http.StatusCreated},
)

func TestRenameService(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(renameServiceInstanceEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}

	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticationRepository{})
	repo := NewCloudControllerServiceRepository(config, gateway)

	serviceInstance := cf.ServiceInstance{Guid: "my-service-instance-guid"}
	apiStatus := repo.RenameService(serviceInstance, "new-name")
	assert.False(t, apiStatus.NotSuccessful())
}
