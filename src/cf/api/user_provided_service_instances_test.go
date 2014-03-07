package api_test

import (
	. "cf/api"
	"cf/models"
	"cf/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestCreateUserProvidedServiceInstance", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "POST",
			Path:     "/v2/user_provided_service_instances",
			Matcher:  testnet.RequestBodyMatcher(`{"name":"my-custom-service","credentials":{"host":"example.com","password":"secret","user":"me"},"space_guid":"my-space-guid","syslog_drain_url":""}`),
			Response: testnet.TestResponse{Status: http.StatusCreated},
		})

		ts, handler, repo := createUserProvidedServiceInstanceRepo(req)
		defer ts.Close()

		apiErr := repo.Create("my-custom-service", "", map[string]string{
			"host":     "example.com",
			"user":     "me",
			"password": "secret",
		})
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})

	It("TestCreateUserProvidedServiceInstanceWithSyslogDrain", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "POST",
			Path:     "/v2/user_provided_service_instances",
			Matcher:  testnet.RequestBodyMatcher(`{"name":"my-custom-service","credentials":{"host":"example.com","password":"secret","user":"me"},"space_guid":"my-space-guid","syslog_drain_url":"syslog://example.com"}`),
			Response: testnet.TestResponse{Status: http.StatusCreated},
		})

		ts, handler, repo := createUserProvidedServiceInstanceRepo(req)
		defer ts.Close()

		apiErr := repo.Create("my-custom-service", "syslog://example.com", map[string]string{
			"host":     "example.com",
			"user":     "me",
			"password": "secret",
		})
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})

	It("TestUpdateUserProvidedServiceInstance", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "PUT",
			Path:     "/v2/user_provided_service_instances/my-instance-guid",
			Matcher:  testnet.RequestBodyMatcher(`{"credentials":{"host":"example.com","password":"secret","user":"me"},"syslog_drain_url":"syslog://example.com"}`),
			Response: testnet.TestResponse{Status: http.StatusCreated},
		})

		ts, handler, repo := createUserProvidedServiceInstanceRepo(req)
		defer ts.Close()

		params := map[string]string{
			"host":     "example.com",
			"user":     "me",
			"password": "secret",
		}
		serviceInstance := models.ServiceInstanceFields{}
		serviceInstance.Guid = "my-instance-guid"
		serviceInstance.Params = params
		serviceInstance.SysLogDrainUrl = "syslog://example.com"

		apiErr := repo.Update(serviceInstance)
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})

	It("TestUpdateUserProvidedServiceInstanceWithOnlyParams", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "PUT",
			Path:     "/v2/user_provided_service_instances/my-instance-guid",
			Matcher:  testnet.RequestBodyMatcher(`{"credentials":{"host":"example.com","password":"secret","user":"me"}}`),
			Response: testnet.TestResponse{Status: http.StatusCreated},
		})

		ts, handler, repo := createUserProvidedServiceInstanceRepo(req)
		defer ts.Close()

		params := map[string]string{
			"host":     "example.com",
			"user":     "me",
			"password": "secret",
		}
		serviceInstance := models.ServiceInstanceFields{}
		serviceInstance.Guid = "my-instance-guid"
		serviceInstance.Params = params
		apiErr := repo.Update(serviceInstance)
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})

	It("TestUpdateUserProvidedServiceInstanceWithOnlySysLogDrainUrl", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "PUT",
			Path:     "/v2/user_provided_service_instances/my-instance-guid",
			Matcher:  testnet.RequestBodyMatcher(`{"syslog_drain_url":"syslog://example.com"}`),
			Response: testnet.TestResponse{Status: http.StatusCreated},
		})

		ts, handler, repo := createUserProvidedServiceInstanceRepo(req)
		defer ts.Close()
		serviceInstance := models.ServiceInstanceFields{}
		serviceInstance.Guid = "my-instance-guid"
		serviceInstance.SysLogDrainUrl = "syslog://example.com"
		apiErr := repo.Update(serviceInstance)
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})
})

func createUserProvidedServiceInstanceRepo(req testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo UserProvidedServiceInstanceRepository) {
	ts, handler = testnet.NewServer([]testnet.TestRequest{req})
	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway(configRepo)
	repo = NewCCUserProvidedServiceInstanceRepository(configRepo, gateway)
	return
}
