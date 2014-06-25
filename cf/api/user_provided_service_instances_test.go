package api_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"

	. "github.com/cloudfoundry/cli/cf/api"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UserProvidedServiceRepository", func() {
	It("creates a user provided service with a name and credentials", func() {
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
		Expect(handler).To(HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})

	It("creates user provided service instances with syslog drains", func() {
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
		Expect(handler).To(HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})

	It("can update a user provided service, given a service instance with a guid", func() {
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
		Expect(handler).To(HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})
})

func createUserProvidedServiceInstanceRepo(req testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo UserProvidedServiceInstanceRepository) {
	ts, handler = testnet.NewServer([]testnet.TestRequest{req})
	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway(configRepo, time.Now)
	repo = NewCCUserProvidedServiceInstanceRepository(configRepo, gateway)
	return
}
