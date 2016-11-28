package api_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testnet "code.cloudfoundry.org/cli/util/testhelpers/net"

	. "code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UserProvidedServiceRepository", func() {

	Context("Create()", func() {
		It("creates a user provided service with a name and credentials", func() {
			req := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "POST",
				Path:     "/v2/user_provided_service_instances",
				Matcher:  testnet.RequestBodyMatcher(`{"name":"my-custom-service","credentials":{"host":"example.com","password":"secret","user":"me"},"space_guid":"my-space-guid","syslog_drain_url":"","route_service_url":""}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			ts, handler, repo := createUserProvidedServiceInstanceRepo([]testnet.TestRequest{req})
			defer ts.Close()

			apiErr := repo.Create("my-custom-service", "", "", map[string]interface{}{
				"host":     "example.com",
				"user":     "me",
				"password": "secret",
			})
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})

		It("creates user provided service instances with syslog drains", func() {
			req := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "POST",
				Path:     "/v2/user_provided_service_instances",
				Matcher:  testnet.RequestBodyMatcher(`{"name":"my-custom-service","credentials":{"host":"example.com","password":"secret","user":"me"},"space_guid":"my-space-guid","syslog_drain_url":"syslog://example.com","route_service_url":""}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			ts, handler, repo := createUserProvidedServiceInstanceRepo([]testnet.TestRequest{req})
			defer ts.Close()

			apiErr := repo.Create("my-custom-service", "syslog://example.com", "", map[string]interface{}{
				"host":     "example.com",
				"user":     "me",
				"password": "secret",
			})
			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})

		It("creates user provided service instances with route service url", func() {
			req := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "POST",
				Path:     "/v2/user_provided_service_instances",
				Matcher:  testnet.RequestBodyMatcher(`{"name":"my-custom-service","credentials":{"host":"example.com","password":"secret","user":"me"},"space_guid":"my-space-guid","syslog_drain_url":"","route_service_url":"https://example.com"}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			ts, handler, repo := createUserProvidedServiceInstanceRepo([]testnet.TestRequest{req})
			defer ts.Close()

			apiErr := repo.Create("my-custom-service", "", "https://example.com", map[string]interface{}{
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
			req := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/user_provided_service_instances/my-instance-guid",
				Matcher:  testnet.RequestBodyMatcher(`{"credentials":{"host":"example.com","password":"secret","user":"me"},"syslog_drain_url":"syslog://example.com","route_service_url":""}`),
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
			serviceInstance.GUID = "my-instance-guid"
			serviceInstance.Params = params
			serviceInstance.SysLogDrainURL = "syslog://example.com"
			serviceInstance.RouteServiceURL = ""

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
            "syslog_drain_url": "sample/drainURL",
            "space_url": "/v2/spaces/f36dbf3e-eff1-4336-ae5c-aff01dd8ce94",
            "service_bindings_url": "/v2/user_provided_service_instances/2d0a1eb6-b6e5-4b92-b1da-91c5e826b3b4/service_bindings"
         }
      }
   ]
}`}

			req := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
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
			Expect(summaries.Resources[1].SysLogDrainURL).To(Equal("sample/drainURL"))
		})
	})

})

func createUserProvidedServiceInstanceRepo(req []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo UserProvidedServiceInstanceRepository) {
	ts, handler = testnet.NewServer(req)
	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetAPIEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
	repo = NewCCUserProvidedServiceInstanceRepository(configRepo, gateway)
	return
}
