package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testnet "testhelpers/net"
	"testing"
)

func TestCreateUserProvidedServiceInstance(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "POST",
		Path:     "/v2/user_provided_service_instances",
		Matcher:  testnet.RequestBodyMatcher(`{"name":"my-custom-service","credentials":{"host":"example.com","password":"secret","user":"me"},"space_guid":"my-space-guid","syslog_drain_url":""}`),
		Response: testnet.TestResponse{Status: http.StatusCreated},
	})

	ts, handler, repo := createUserProvidedServiceInstanceRepo(t, req)
	defer ts.Close()

	apiResponse := repo.Create("my-custom-service", "", map[string]string{
		"host":     "example.com",
		"user":     "me",
		"password": "secret",
	})
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestCreateUserProvidedServiceInstanceWithSyslogDrain(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "POST",
		Path:     "/v2/user_provided_service_instances",
		Matcher:  testnet.RequestBodyMatcher(`{"name":"my-custom-service","credentials":{"host":"example.com","password":"secret","user":"me"},"space_guid":"my-space-guid","syslog_drain_url":"syslog://example.com"}`),
		Response: testnet.TestResponse{Status: http.StatusCreated},
	})

	ts, handler, repo := createUserProvidedServiceInstanceRepo(t, req)
	defer ts.Close()

	apiResponse := repo.Create("my-custom-service", "syslog://example.com", map[string]string{
		"host":     "example.com",
		"user":     "me",
		"password": "secret",
	})
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestUpdateUserProvidedServiceInstance(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     "/v2/user_provided_service_instances/my-instance-guid",
		Matcher:  testnet.RequestBodyMatcher(`{"credentials":{"host":"example.com","password":"secret","user":"me"},"syslog_drain_url":"syslog://example.com"}`),
		Response: testnet.TestResponse{Status: http.StatusCreated},
	})

	ts, handler, repo := createUserProvidedServiceInstanceRepo(t, req)
	defer ts.Close()

	params := map[string]string{
		"host":     "example.com",
		"user":     "me",
		"password": "secret",
	}
	serviceInstance := cf.ServiceInstanceFields{}
	serviceInstance.Guid = "my-instance-guid"
	serviceInstance.Params = params
	serviceInstance.SysLogDrainUrl = "syslog://example.com"

	apiResponse := repo.Update(serviceInstance)
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestUpdateUserProvidedServiceInstanceWithOnlyParams(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     "/v2/user_provided_service_instances/my-instance-guid",
		Matcher:  testnet.RequestBodyMatcher(`{"credentials":{"host":"example.com","password":"secret","user":"me"}}`),
		Response: testnet.TestResponse{Status: http.StatusCreated},
	})

	ts, handler, repo := createUserProvidedServiceInstanceRepo(t, req)
	defer ts.Close()

	params := map[string]string{
		"host":     "example.com",
		"user":     "me",
		"password": "secret",
	}
	serviceInstance := cf.ServiceInstanceFields{}
	serviceInstance.Guid = "my-instance-guid"
	serviceInstance.Params = params
	apiResponse := repo.Update(serviceInstance)
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestUpdateUserProvidedServiceInstanceWithOnlySysLogDrainUrl(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     "/v2/user_provided_service_instances/my-instance-guid",
		Matcher:  testnet.RequestBodyMatcher(`{"syslog_drain_url":"syslog://example.com"}`),
		Response: testnet.TestResponse{Status: http.StatusCreated},
	})

	ts, handler, repo := createUserProvidedServiceInstanceRepo(t, req)
	defer ts.Close()
	serviceInstance := cf.ServiceInstanceFields{}
	serviceInstance.Guid = "my-instance-guid"
	serviceInstance.SysLogDrainUrl = "syslog://example.com"
	apiResponse := repo.Update(serviceInstance)
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func createUserProvidedServiceInstanceRepo(t *testing.T, req testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo UserProvidedServiceInstanceRepository) {
	ts, handler = testnet.NewTLSServer(t, []testnet.TestRequest{req})
	space := cf.SpaceFields{}
	space.Guid = "my-space-guid"
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		SpaceFields: space,
		Target:      ts.URL,
	}

	gateway := net.NewCloudControllerGateway()
	repo = NewCCUserProvidedServiceInstanceRepository(config, gateway)
	return
}
