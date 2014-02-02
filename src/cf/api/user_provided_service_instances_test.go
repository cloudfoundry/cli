package api_test

import (
	"cf"
	. "cf/api"
	"cf/configuration"
	"cf/net"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testnet "testhelpers/net"
)

func createUserProvidedServiceInstanceRepo(t mr.TestingT, req testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo UserProvidedServiceInstanceRepository) {
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
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestCreateUserProvidedServiceInstance", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "POST",
				Path:     "/v2/user_provided_service_instances",
				Matcher:  testnet.RequestBodyMatcher(`{"name":"my-custom-service","credentials":{"host":"example.com","password":"secret","user":"me"},"space_guid":"my-space-guid","syslog_drain_url":""}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			ts, handler, repo := createUserProvidedServiceInstanceRepo(mr.T(), req)
			defer ts.Close()

			apiResponse := repo.Create("my-custom-service", "", map[string]string{
				"host":     "example.com",
				"user":     "me",
				"password": "secret",
			})
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
		})
		It("TestCreateUserProvidedServiceInstanceWithSyslogDrain", func() {

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "POST",
				Path:     "/v2/user_provided_service_instances",
				Matcher:  testnet.RequestBodyMatcher(`{"name":"my-custom-service","credentials":{"host":"example.com","password":"secret","user":"me"},"space_guid":"my-space-guid","syslog_drain_url":"syslog://example.com"}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			ts, handler, repo := createUserProvidedServiceInstanceRepo(mr.T(), req)
			defer ts.Close()

			apiResponse := repo.Create("my-custom-service", "syslog://example.com", map[string]string{
				"host":     "example.com",
				"user":     "me",
				"password": "secret",
			})
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
		})
		It("TestUpdateUserProvidedServiceInstance", func() {

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/user_provided_service_instances/my-instance-guid",
				Matcher:  testnet.RequestBodyMatcher(`{"credentials":{"host":"example.com","password":"secret","user":"me"},"syslog_drain_url":"syslog://example.com"}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			ts, handler, repo := createUserProvidedServiceInstanceRepo(mr.T(), req)
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
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
		})
		It("TestUpdateUserProvidedServiceInstanceWithOnlyParams", func() {

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/user_provided_service_instances/my-instance-guid",
				Matcher:  testnet.RequestBodyMatcher(`{"credentials":{"host":"example.com","password":"secret","user":"me"}}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			ts, handler, repo := createUserProvidedServiceInstanceRepo(mr.T(), req)
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
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
		})
		It("TestUpdateUserProvidedServiceInstanceWithOnlySysLogDrainUrl", func() {

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/user_provided_service_instances/my-instance-guid",
				Matcher:  testnet.RequestBodyMatcher(`{"syslog_drain_url":"syslog://example.com"}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			ts, handler, repo := createUserProvidedServiceInstanceRepo(mr.T(), req)
			defer ts.Close()
			serviceInstance := cf.ServiceInstanceFields{}
			serviceInstance.Guid = "my-instance-guid"
			serviceInstance.SysLogDrainUrl = "syslog://example.com"
			apiResponse := repo.Update(serviceInstance)
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
		})
	})
}
