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

func TestServiceBrokersFindAll(t *testing.T) {
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

	expectedServiceBrokers := []cf.ServiceBroker{
		{
			Guid:     "found-guid-1",
			Name:     "found-name-1",
			Url:      "http://found.example.com-1",
			Username: "found-username-1",
			Password: "found-password-1",
		},
		{
			Guid:     "found-guid-2",
			Name:     "found-name-2",
			Url:      "http://found.example.com-2",
			Username: "found-username-2",
			Password: "found-password-2",
		},
	}

	serviceBrokers := []cf.ServiceBroker{}
	for chunk := range serviceBrokersChan {
		serviceBrokers = append(serviceBrokers, chunk...)
	}
	apiResponse := <-statusChan

	assert.Equal(t, serviceBrokers, expectedServiceBrokers)
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
	expectedBroker := cf.ServiceBroker{
		Name:     "found-name",
		Url:      "http://found.example.com",
		Username: "found-username",
		Password: "found-password",
		Guid:     "found-guid",
	}

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
	assert.Equal(t, apiResponse.Message, "Service Broker my-broker not found")
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

	serviceBroker := cf.ServiceBroker{
		Name:     "foobroker",
		Url:      "http://example.com",
		Username: "foouser",
		Password: "password",
	}
	apiResponse := repo.Create(serviceBroker)

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

	serviceBroker := cf.ServiceBroker{
		Guid:     "my-guid",
		Name:     "foobroker",
		Url:      "http://update.example.com",
		Username: "update-foouser",
		Password: "update-password",
	}
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

	serviceBroker := cf.ServiceBroker{
		Guid: "my-guid",
		Name: "update-foobroker",
	}
	apiResponse := repo.Rename(serviceBroker)

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

	serviceBroker := cf.ServiceBroker{
		Guid: "my-guid",
	}
	apiResponse := repo.Delete(serviceBroker)

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
