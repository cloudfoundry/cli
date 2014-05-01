/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/github.com/cloudfoundry/cli/cf/commands/application/delete_app_test.go
   src/github.com/cloudfoundry/cli/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package api_test

import (
	. "github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
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
