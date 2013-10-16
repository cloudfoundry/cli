package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	"testing"
)

func TestCreateServiceBinding(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"POST",
		"/v2/service_bindings",
		testapi.RequestBodyMatcher(`{"app_guid":"my-app-guid","service_instance_guid":"my-service-instance-guid"}`),
		testapi.TestResponse{Status: http.StatusCreated},
	)

	ts, repo := createServiceBindingRepo(endpoint)
	defer ts.Close()

	serviceInstance := cf.ServiceInstance{Guid: "my-service-instance-guid"}
	app := cf.Application{Guid: "my-app-guid"}
	apiResponse := repo.Create(serviceInstance, app)
	assert.True(t, status.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestCreateServiceBindingIfError(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"POST",
		"/v2/service_bindings",
		testapi.RequestBodyMatcher(`{"app_guid":"my-app-guid","service_instance_guid":"my-service-instance-guid"}`),
		testapi.TestResponse{
			Status: http.StatusBadRequest,
			Body:   `{"code":90003,"description":"The app space binding to service is taken: 7b959018-110a-4913-ac0a-d663e613cdea 346bf237-7eef-41a7-b892-68fb08068f09"}`,
		},
	)

	ts, repo := createServiceBindingRepo(endpoint)
	defer ts.Close()

	serviceInstance := cf.ServiceInstance{Guid: "my-service-instance-guid"}
	app := cf.Application{Guid: "my-app-guid"}
	apiResponse := repo.Create(serviceInstance, app)

	assert.True(t, status.Called())
	assert.True(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, apiResponse.ErrorCode, "90003")
}

var deleteBindingEndpoint, deleteBindingEndpointStatus = testapi.CreateCheckableEndpoint(
	"DELETE",
	"/v2/service_bindings/service-binding-2-guid",
	nil,
	testapi.TestResponse{Status: http.StatusOK},
)

func TestDeleteServiceBinding(t *testing.T) {
	deleteBindingEndpointStatus.Reset()
	ts, repo := createServiceBindingRepo(deleteBindingEndpoint)
	defer ts.Close()

	serviceBindings := []cf.ServiceBinding{
		cf.ServiceBinding{Url: "/v2/service_bindings/service-binding-1-guid", AppGuid: "app-1-guid"},
		cf.ServiceBinding{Url: "/v2/service_bindings/service-binding-2-guid", AppGuid: "app-2-guid"},
	}

	serviceInstance := cf.ServiceInstance{
		Guid:            "my-service-instance-guid",
		ServiceBindings: serviceBindings,
	}
	app := cf.Application{Guid: "app-2-guid"}
	found, apiResponse := repo.Delete(serviceInstance, app)
	assert.True(t, deleteBindingEndpointStatus.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.True(t, found)
}

func TestDeleteServiceBindingWhenBindingDoesNotExist(t *testing.T) {
	ts, repo := createServiceBindingRepo(deleteBindingEndpoint)
	defer ts.Close()

	serviceBindings := []cf.ServiceBinding{}

	serviceInstance := cf.ServiceInstance{
		Guid:            "my-service-instance-guid",
		ServiceBindings: serviceBindings,
	}
	app := cf.Application{Guid: "app-2-guid"}
	found, apiResponse := repo.Delete(serviceInstance, app)
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.False(t, found)
}

func createServiceBindingRepo(endpoint http.HandlerFunc) (ts *httptest.Server, repo ServiceBindingRepository) {
	ts = httptest.NewTLSServer(endpoint)

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Space:       cf.Space{Guid: "my-space-guid"},
		Target:      ts.URL,
	}

	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerServiceBindingRepository(config, gateway)
	return
}
