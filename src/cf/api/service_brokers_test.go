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

func TestServiceBrokersListServiceBrokers(t *testing.T) {
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
			Body: `{
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

	ts, handler, repo := createServiceBrokerRepo(t, firstRequest, secondRequest)
	defer ts.Close()

	stopChan := make(chan bool)
	defer close(stopChan)
	serviceBrokersChan, statusChan := repo.ListServiceBrokers(stopChan)

	serviceBrokers := []cf.ServiceBroker{}
	for chunk := range serviceBrokersChan {
		serviceBrokers = append(serviceBrokers, chunk...)
	}
	apiResponse := <-statusChan

	assert.Equal(t, len(serviceBrokers), 2)
	assert.Equal(t, serviceBrokers[0].Guid, "found-guid-1")
	assert.Equal(t, serviceBrokers[1].Guid, "found-guid-2")
	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestServiceBrokersListServiceBrokersWithNoServiceBrokers(t *testing.T) {
	emptyServiceBrokersRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/service_brokers",
		Response: testnet.TestResponse{
			Status: http.StatusOK,
			Body:   `{"resources": []}`,
		},
	})

	ts, handler, repo := createServiceBrokerRepo(t, emptyServiceBrokersRequest)
	defer ts.Close()

	stopChan := make(chan bool)
	defer close(stopChan)
	serviceBrokersChan, statusChan := repo.ListServiceBrokers(stopChan)

	_, ok := <-serviceBrokersChan
	apiResponse := <-statusChan

	assert.False(t, ok)
	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestFindServiceBrokerByName(t *testing.T) {
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

	ts, handler, repo := createServiceBrokerRepo(t, req)
	defer ts.Close()

	foundBroker, apiResponse := repo.FindByName("my-broker")
	expectedBroker := cf.ServiceBroker{}
	expectedBroker.Name = "found-name"
	expectedBroker.Url = "http://found.example.com"
	expectedBroker.Username = "found-username"
	expectedBroker.Password = "found-password"
	expectedBroker.Guid = "found-guid"

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, foundBroker, expectedBroker)
}

func TestFindServiceBrokerByNameWheNotFound(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/v2/service_brokers?q=name%3Amy-broker",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{ "resources": [ ] }`},
	})

	ts, handler, repo := createServiceBrokerRepo(t, req)
	defer ts.Close()

	_, apiResponse := repo.FindByName("my-broker")

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsNotFound())
	assert.Equal(t, apiResponse.Message, "Service Broker 'my-broker' not found")
}

func TestCreateServiceBroker(t *testing.T) {
	expectedReqBody := `{"name":"foobroker","broker_url":"http://example.com","auth_username":"foouser","auth_password":"password"}`

	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "POST",
		Path:     "/v2/service_brokers",
		Matcher:  testnet.RequestBodyMatcher(expectedReqBody),
		Response: testnet.TestResponse{Status: http.StatusCreated},
	})

	ts, handler, repo := createServiceBrokerRepo(t, req)
	defer ts.Close()

	apiResponse := repo.Create("foobroker", "http://example.com", "foouser", "password")

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestUpdateServiceBroker(t *testing.T) {
	expectedReqBody := `{"broker_url":"http://update.example.com","auth_username":"update-foouser","auth_password":"update-password"}`

	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     "/v2/service_brokers/my-guid",
		Matcher:  testnet.RequestBodyMatcher(expectedReqBody),
		Response: testnet.TestResponse{Status: http.StatusOK},
	})

	ts, handler, repo := createServiceBrokerRepo(t, req)
	defer ts.Close()
	serviceBroker := cf.ServiceBroker{}
	serviceBroker.Guid = "my-guid"
	serviceBroker.Name = "foobroker"
	serviceBroker.Url = "http://update.example.com"
	serviceBroker.Username = "update-foouser"
	serviceBroker.Password = "update-password"

	apiResponse := repo.Update(serviceBroker)

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestRenameServiceBroker(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     "/v2/service_brokers/my-guid",
		Matcher:  testnet.RequestBodyMatcher(`{"name":"update-foobroker"}`),
		Response: testnet.TestResponse{Status: http.StatusOK},
	})

	ts, handler, repo := createServiceBrokerRepo(t, req)
	defer ts.Close()

	apiResponse := repo.Rename("my-guid", "update-foobroker")

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestDeleteServiceBroker(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "DELETE",
		Path:     "/v2/service_brokers/my-guid",
		Response: testnet.TestResponse{Status: http.StatusNoContent},
	})

	ts, handler, repo := createServiceBrokerRepo(t, req)
	defer ts.Close()

	apiResponse := repo.Delete("my-guid")

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

func createServiceBrokerRepo(t *testing.T, requests ...testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo ServiceBrokerRepository) {
	ts, handler = testnet.NewTLSServer(t, requests)

	config := &configuration.Configuration{
		Target:      ts.URL,
		AccessToken: "BEARER my_access_token",
	}

	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerServiceBrokerRepository(config, gateway)
	return
}
