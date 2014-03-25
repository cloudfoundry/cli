package api_test

import (
	. "cf/api"
	"cf/errors"
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

var _ = Describe("Service Brokers Repo", func() {
	It("TestServiceBrokersListServiceBrokers", func() {
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

	It("TestFindServiceBrokerByName", func() {
		responseBody := `{
"resources": [
{
  "metadata": {
	"guid":"found-guid"
  },
  "entity": {
	"name": "found-name",
	"broker_url": "http://found.example.com",
	"auth_username": "found-username",
	"auth_password": "found-password"
  }
}
]
}`

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

	It("TestFindServiceBrokerByNameWheNotFound", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "GET",
			Path:     "/v2/service_brokers?q=name%3Amy-broker",
			Response: testnet.TestResponse{Status: http.StatusOK, Body: `{ "resources": [ ] }`},
		})

		ts, handler, repo := createServiceBrokerRepo(req)
		defer ts.Close()

		_, apiErr := repo.FindByName("my-broker")

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr.(*errors.ModelNotFoundError)).NotTo(BeNil())
		Expect(apiErr.Error()).To(Equal("Service Broker my-broker not found"))
	})

	It("TestCreateServiceBroker", func() {
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

	It("TestUpdateServiceBroker", func() {
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

	It("TestRenameServiceBroker", func() {
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

	It("TestDeleteServiceBroker", func() {
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

func createServiceBrokerRepo(requests ...testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo ServiceBrokerRepository) {
	ts, handler = testnet.NewServer(requests)
	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway(configRepo)
	repo = NewCloudControllerServiceBrokerRepository(configRepo, gateway)
	return
}
