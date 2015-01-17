package api_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/api"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UserProvidedServiceRepository", func() {

	Context("Create()", func() {
		It("creates a user provided service with a name and credentials", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "POST",
				Path:     "/v2/user_provided_service_instances",
				Matcher:  testnet.RequestBodyMatcher(`{"name":"my-custom-service","credentials":{"host":"example.com","password":"secret","user":"me"},"space_guid":"my-space-guid","syslog_drain_url":""}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			ts, handler, repo := createUserProvidedServiceInstanceRepo([]testnet.TestRequest{req})
			defer ts.Close()

			apiErr := repo.Create("my-custom-service", "", map[string]interface{}{
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

			ts, handler, repo := createUserProvidedServiceInstanceRepo([]testnet.TestRequest{req})
			defer ts.Close()

			apiErr := repo.Create("my-custom-service", "syslog://example.com", map[string]interface{}{
				"host":     "example.com",
				"user":     "me",
				"password": "secret",
			})
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})
	})

	Context("Update()", func() {
		It("can update a user provided service, given a service instance with a guid", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/user_provided_service_instances/my-instance-guid",
				Matcher:  testnet.RequestBodyMatcher(`{"credentials":{"host":"example.com","password":"secret","user":"me"},"syslog_drain_url":"syslog://example.com"}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			ts, handler, repo := createUserProvidedServiceInstanceRepo([]testnet.TestRequest{req})
			defer ts.Close()

			params := map[string]interface{}{
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

	Context("GetSummaries()", func() {
		It("returns all user created service in []models.UserProvidedService", func() {
			responseStr := testnet.TestResponse{Status: http.StatusOK, Body: `
{
   "total_results": 2,
   "total_pages": 1,
   "prev_url": null,
   "next_url": null,
   "resources": [
      {
         "metadata": {
            "guid": "2d0a1eb6-b6e5-4b92-b1da-91c5e826b3b4",
            "url": "/v2/user_provided_service_instances/2d0a1eb6-b6e5-4b92-b1da-91c5e826b3b4",
            "created_at": "2015-01-15T22:57:08Z",
            "updated_at": null
         },
         "entity": {
            "name": "test_service",
            "credentials": {},
            "space_guid": "f36dbf3e-eff1-4336-ae5c-aff01dd8ce94",
            "type": "user_provided_service_instance",
            "syslog_drain_url": "",
            "space_url": "/v2/spaces/f36dbf3e-eff1-4336-ae5c-aff01dd8ce94",
            "service_bindings_url": "/v2/user_provided_service_instances/2d0a1eb6-b6e5-4b92-b1da-91c5e826b3b4/service_bindings"
         }
      },
      {
         "metadata": {
            "guid": "9d0a1eb6-b6e5-4b92-b1da-91c5ed26b3b4",
            "url": "/v2/user_provided_service_instances/9d0a1eb6-b6e5-4b92-b1da-91c5e826b3b4",
            "created_at": "2015-01-15T22:57:08Z",
            "updated_at": null
         },
         "entity": {
            "name": "test_service2",
            "credentials": {
							"password": "admin",
              "username": "admin"
						},
            "space_guid": "f36dbf3e-eff1-4336-ae5c-aff01dd8ce94",
            "type": "user_provided_service_instance",
            "syslog_drain_url": "sample/drainUrl",
            "space_url": "/v2/spaces/f36dbf3e-eff1-4336-ae5c-aff01dd8ce94",
            "service_bindings_url": "/v2/user_provided_service_instances/2d0a1eb6-b6e5-4b92-b1da-91c5e826b3b4/service_bindings"
         }
      }
   ]
}`}

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/user_provided_service_instances",
				Response: responseStr,
			})

			ts, handler, repo := createUserProvidedServiceInstanceRepo([]testnet.TestRequest{req})
			defer ts.Close()

			summaries, apiErr := repo.GetSummaries()
			Expect(apiErr).NotTo(HaveOccurred())
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(len(summaries.Resources)).To(Equal(2))

			Expect(summaries.Resources[0].Name).To(Equal("test_service"))
			Expect(summaries.Resources[1].Name).To(Equal("test_service2"))
			Expect(summaries.Resources[1].Credentials["username"]).To(Equal("admin"))
			Expect(summaries.Resources[1].SysLogDrainUrl).To(Equal("sample/drainUrl"))
		})
	})

})

func createUserProvidedServiceInstanceRepo(req []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo UserProvidedServiceInstanceRepository) {
	ts, handler = testnet.NewServer(req)
	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway(configRepo, time.Now, &testterm.FakeUI{})
	repo = NewCCUserProvidedServiceInstanceRepository(configRepo, gateway)
	return
}
