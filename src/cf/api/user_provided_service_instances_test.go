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

func TestCreateUserProvidedServiceInstance(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"POST",
		"/v2/user_provided_service_instances",
		testapi.RequestBodyMatcher(`{"name":"my-custom-service","credentials":{"host":"example.com","password":"secret","user":"me"},"space_guid":"my-space-guid"}`),
		testapi.TestResponse{Status: http.StatusCreated},
	)

	ts, repo := createUserProvidedServiceInstanceRepo(endpoint)
	defer ts.Close()

	params := map[string]string{
		"host":     "example.com",
		"user":     "me",
		"password": "secret",
	}
	apiResponse := repo.Create("my-custom-service", params)
	assert.True(t, status.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestUpdateUserProvidedServiceInstance(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"PUT",
		"/v2/user_provided_service_instances/my-instance-guid",
		testapi.RequestBodyMatcher(`{"credentials":{"host":"example.com","password":"secret","user":"me"}}`),
		testapi.TestResponse{Status: http.StatusCreated},
	)

	ts, repo := createUserProvidedServiceInstanceRepo(endpoint)
	defer ts.Close()

	params := map[string]string{
		"host":     "example.com",
		"user":     "me",
		"password": "secret",
	}
	apiResponse := repo.Update(cf.ServiceInstance{Guid: "my-instance-guid"}, params)
	assert.True(t, status.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func createUserProvidedServiceInstanceRepo(endpoint http.HandlerFunc) (ts *httptest.Server, repo UserProvidedServiceInstanceRepository) {
	ts = httptest.NewTLSServer(endpoint)

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Space:       cf.Space{Guid: "my-space-guid"},
		Target:      ts.URL,
	}

	gateway := net.NewCloudControllerGateway()
	repo = NewCCUserProvidedServiceInstanceRepository(config, gateway)
	return
}
