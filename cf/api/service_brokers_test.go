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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Brokers Repo", func() {
	It("lists services brokers", func() {
		firstRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "GET",
			Path:   "/v2/service_brokers",
			Response: testnet.TestResponse{
				Status: http.StatusOK,
				Body: `{
				  "next_url": "/v2/service_brokers?page=2",
				  "resources": [
					{
					  "metadata": {
						"guid":"found-guid-1"
					  },
					  "entity": {
						"name": "found-name-1",
						"broker_url": "http://found.example.com-1",
						"auth_username": "found-username-1",
						"auth_password": "found-password-1"
					  }
					}
				  ]
				}`,
			},
		})

		secondRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "GET",
			Path:   "/v2/service_brokers?page=2",
			Response: testnet.TestResponse{
				Status: http.StatusOK,
				Body: `
				{
				  "resources": [
					{
					  "metadata": {
						"guid":"found-guid-2"
					  },
					  "entity": {
						"name": "found-name-2",
						"broker_url": "http://found.example.com-2",
						"auth_username": "found-username-2",
						"auth_password": "found-password-2"
					  }
					}
				  ]
				}`,
			},
		})

		ts, handler, repo := createServiceBrokerRepo(firstRequest, secondRequest)
		defer ts.Close()

		serviceBrokers := []models.ServiceBroker{}
		apiErr := repo.ListServiceBrokers(func(broker models.ServiceBroker) bool {
			serviceBrokers = append(serviceBrokers, broker)
			return true
		})

		Expect(len(serviceBrokers)).To(Equal(2))
		Expect(serviceBrokers[0].Guid).To(Equal("found-guid-1"))
		Expect(serviceBrokers[1].Guid).To(Equal("found-guid-2"))
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})

	Describe("FindByName", func() {
		It("returns the service broker with the given name", func() {
			responseBody := `
{"resources": [{
  "metadata": {"guid":"found-guid"},
  "entity": {
  	"name": "found-name",
		"broker_url": "http://found.example.com",
		"auth_username": "found-username",
		"auth_password": "found-password"
  }
}]}`

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/service_brokers?q=name%3Amy-broker",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: responseBody},
			})

			ts, handler, repo := createServiceBrokerRepo(req)
			defer ts.Close()

			foundBroker, apiErr := repo.FindByName("my-broker")
			expectedBroker := models.ServiceBroker{}
			expectedBroker.Name = "found-name"
			expectedBroker.Url = "http://found.example.com"
			expectedBroker.Username = "found-username"
			expectedBroker.Password = "found-password"
			expectedBroker.Guid = "found-guid"

			Expect(handler).To(testnet.HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
			Expect(foundBroker).To(Equal(expectedBroker))
		})

		It("returns an error when the service broker cannot be found", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/service_brokers?q=name%3Amy-broker",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{ "resources": [ ] }`},
			})

			ts, handler, repo := createServiceBrokerRepo(req)
			defer ts.Close()

			_, apiErr := repo.FindByName("my-broker")

			Expect(handler).To(testnet.HaveAllRequestsCalled())
			Expect(apiErr).To(HaveOccurred())
			Expect(apiErr.Error()).To(Equal("Service Broker my-broker not found"))
		})
	})

	Describe("Create", func() {
		It("creates the service broker with the given name, URL, username and password", func() {
			expectedReqBody := `{"name":"foobroker","broker_url":"http://example.com","auth_username":"foouser","auth_password":"password"}`

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "POST",
				Path:     "/v2/service_brokers",
				Matcher:  testnet.RequestBodyMatcher(expectedReqBody),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			ts, handler, repo := createServiceBrokerRepo(req)
			defer ts.Close()

			apiErr := repo.Create("foobroker", "http://example.com", "foouser", "password")

			Expect(handler).To(testnet.HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})
	})

	Describe("Update", func() {
		It("updates the service broker with the given guid", func() {
			expectedReqBody := `{"broker_url":"http://update.example.com","auth_username":"update-foouser","auth_password":"update-password"}`

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/service_brokers/my-guid",
				Matcher:  testnet.RequestBodyMatcher(expectedReqBody),
				Response: testnet.TestResponse{Status: http.StatusOK},
			})

			ts, handler, repo := createServiceBrokerRepo(req)
			defer ts.Close()
			serviceBroker := models.ServiceBroker{}
			serviceBroker.Guid = "my-guid"
			serviceBroker.Name = "foobroker"
			serviceBroker.Url = "http://update.example.com"
			serviceBroker.Username = "update-foouser"
			serviceBroker.Password = "update-password"

			apiErr := repo.Update(serviceBroker)

			Expect(handler).To(testnet.HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})
	})

	Describe("Rename", func() {
		It("renames the service broker with the given guid", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/service_brokers/my-guid",
				Matcher:  testnet.RequestBodyMatcher(`{"name":"update-foobroker"}`),
				Response: testnet.TestResponse{Status: http.StatusOK},
			})

			ts, handler, repo := createServiceBrokerRepo(req)
			defer ts.Close()

			apiErr := repo.Rename("my-guid", "update-foobroker")

			Expect(handler).To(testnet.HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})
	})

	Describe("Delete", func() {
		It("deletes the service broker with the given guid", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "DELETE",
				Path:     "/v2/service_brokers/my-guid",
				Response: testnet.TestResponse{Status: http.StatusNoContent},
			})

			ts, handler, repo := createServiceBrokerRepo(req)
			defer ts.Close()

			apiErr := repo.Delete("my-guid")

			Expect(handler).To(testnet.HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})
	})
})

func createServiceBrokerRepo(requests ...testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo ServiceBrokerRepository) {
	ts, handler = testnet.NewServer(requests)
	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway(configRepo, time.Now)
	repo = NewCloudControllerServiceBrokerRepository(configRepo, gateway)
	return
}
