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
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Service Brokers Repo", func() {
	It("lists services brokers", func() {
		firstRequest := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
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

		secondRequest := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
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
		Expect(serviceBrokers[0].GUID).To(Equal("found-guid-1"))
		Expect(serviceBrokers[1].GUID).To(Equal("found-guid-2"))
		Expect(handler).To(HaveAllRequestsCalled())
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

			req := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/service_brokers?q=name%3Amy-broker",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: responseBody},
			})

			ts, handler, repo := createServiceBrokerRepo(req)
			defer ts.Close()

			foundBroker, apiErr := repo.FindByName("my-broker")
			expectedBroker := models.ServiceBroker{}
			expectedBroker.Name = "found-name"
			expectedBroker.URL = "http://found.example.com"
			expectedBroker.Username = "found-username"
			expectedBroker.Password = "found-password"
			expectedBroker.GUID = "found-guid"

			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
			Expect(foundBroker).To(Equal(expectedBroker))
		})

		It("returns an error when the service broker cannot be found", func() {
			req := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/service_brokers?q=name%3Amy-broker",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{ "resources": [ ] }`},
			})

			ts, handler, repo := createServiceBrokerRepo(req)
			defer ts.Close()

			_, apiErr := repo.FindByName("my-broker")

			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).To(HaveOccurred())
			Expect(apiErr.Error()).To(Equal("Service Broker my-broker not found"))
		})

		It("returns an error when listing service brokers returns an api error", func() {
			req := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/service_brokers?q=name%3Amy-broker",
				Response: testnet.TestResponse{Status: http.StatusForbidden, Body: `{
				  "code": 10003,
				  "description": "You are not authorized to perform the requested action",
				  "error_code": "CF-NotAuthorized"
			  }`},
			})

			ts, handler, repo := createServiceBrokerRepo(req)
			defer ts.Close()

			_, apiErr := repo.FindByName("my-broker")

			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).To(HaveOccurred())
			Expect(apiErr.Error()).To(Equal("Server error, status code: 403, error code: 10003, message: You are not authorized to perform the requested action"))
		})
	})

	Describe("FindByGUID", func() {
		It("returns the service broker with the given guid", func() {
			responseBody := `
{
   "metadata": {
      "guid": "found-guid",
      "url": "/v2/service_brokers/found-guid",
      "created_at": "2014-07-24T21:21:54+00:00",
      "updated_at": "2014-07-25T17:03:40+00:00"
   },
   "entity": {
      "name": "found-name",
      "broker_url": "http://found.example.com",
      "auth_username": "found-username"
   }
}
`

			req := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/service_brokers/found-guid",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: responseBody},
			})

			ts, handler, repo := createServiceBrokerRepo(req)
			defer ts.Close()

			foundBroker, apiErr := repo.FindByGUID("found-guid")
			expectedBroker := models.ServiceBroker{}
			expectedBroker.Name = "found-name"
			expectedBroker.URL = "http://found.example.com"
			expectedBroker.Username = "found-username"
			expectedBroker.GUID = "found-guid"

			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
			Expect(foundBroker).To(Equal(expectedBroker))
		})

		It("returns an error when the service broker cannot be found", func() {
			req := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/service_brokers/bogus-guid",
				//This error code may not reflect reality.  Check it, change the code to match, and remove this comment.
				Response: testnet.TestResponse{Status: http.StatusNotFound, Body: `{"error_code":"ServiceBrokerNotFound","description":"Service Broker bogus-guid not found","code":270042}`},
			})

			ts, handler, repo := createServiceBrokerRepo(req)
			defer ts.Close()

			_, apiErr := repo.FindByGUID("bogus-guid")

			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).To(HaveOccurred())
			Expect(apiErr.Error()).To(Equal("Server error, status code: 404, error code: 270042, message: Service Broker bogus-guid not found"))
		})
	})

	Describe("Create", func() {
		var (
			ccServer *ghttp.Server
			repo     CloudControllerServiceBrokerRepository
		)

		BeforeEach(func() {
			ccServer = ghttp.NewServer()

			configRepo := testconfig.NewRepositoryWithDefaults()
			configRepo.SetAPIEndpoint(ccServer.URL())
			gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
			repo = NewCloudControllerServiceBrokerRepository(configRepo, gateway)
		})

		AfterEach(func() {
			ccServer.Close()
		})

		It("creates the service broker with the given name, URL, username and password", func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v2/service_brokers"),
					ghttp.VerifyJSON(`
						{
						    "name": "foobroker",
						    "broker_url": "http://example.com",
						    "auth_username": "foouser",
						    "auth_password": "password"
						}
					`),
					ghttp.RespondWith(http.StatusCreated, nil),
				),
			)

			err := repo.Create("foobroker", "http://example.com", "foouser", "password", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("creates the service broker with the correct params when given a space GUID", func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v2/service_brokers"),
					ghttp.VerifyJSON(`
						{
								"name": "foobroker",
								"broker_url": "http://example.com",
								"auth_username": "foouser",
								"auth_password": "password",
								"space_guid": "space-guid"
						}
					`),
					ghttp.RespondWith(http.StatusCreated, nil),
				),
			)

			err := repo.Create("foobroker", "http://example.com", "foouser", "password", "space-guid")
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})
	})

	Describe("Update", func() {
		It("updates the service broker with the given guid", func() {
			expectedReqBody := `{"broker_url":"http://update.example.com","auth_username":"update-foouser","auth_password":"update-password"}`

			req := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/service_brokers/my-guid",
				Matcher:  testnet.RequestBodyMatcher(expectedReqBody),
				Response: testnet.TestResponse{Status: http.StatusOK},
			})

			ts, handler, repo := createServiceBrokerRepo(req)
			defer ts.Close()
			serviceBroker := models.ServiceBroker{}
			serviceBroker.GUID = "my-guid"
			serviceBroker.Name = "foobroker"
			serviceBroker.URL = "http://update.example.com"
			serviceBroker.Username = "update-foouser"
			serviceBroker.Password = "update-password"

			apiErr := repo.Update(serviceBroker)

			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})
	})

	Describe("Rename", func() {
		It("renames the service broker with the given guid", func() {
			req := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/service_brokers/my-guid",
				Matcher:  testnet.RequestBodyMatcher(`{"name":"update-foobroker"}`),
				Response: testnet.TestResponse{Status: http.StatusOK},
			})

			ts, handler, repo := createServiceBrokerRepo(req)
			defer ts.Close()

			apiErr := repo.Rename("my-guid", "update-foobroker")

			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})
	})

	Describe("Delete", func() {
		It("deletes the service broker with the given guid", func() {
			req := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "DELETE",
				Path:     "/v2/service_brokers/my-guid",
				Response: testnet.TestResponse{Status: http.StatusNoContent},
			})

			ts, handler, repo := createServiceBrokerRepo(req)
			defer ts.Close()

			apiErr := repo.Delete("my-guid")

			Expect(handler).To(HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})
	})
})

func createServiceBrokerRepo(requests ...testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo ServiceBrokerRepository) {
	ts, handler = testnet.NewServer(requests)
	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetAPIEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
	repo = NewCloudControllerServiceBrokerRepository(configRepo, gateway)
	return
}
