package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	testnet "testhelpers/net"
	"testing"
	"time"
)

var firstPageEventsRequest = testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/apps/my-app-guid/events",
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `
{
  "total_results": 58,
  "total_pages": 2,
  "prev_url": null,
  "next_url": "/v2/apps/my-app-guid/events?inline-relations-depth=1&page=2&results-per-page=50",
  "resources": [
    {
      "entity": {
        "instance_index": 1,
        "exit_status": 1,
        "exit_description": "app instance exited",
        "timestamp": "2013-10-07T16:51:07+00:00"
      }
    }
  ]
}
`},
}
var secondPageEventsRequest = testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/apps/my-app-guid/events",
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `
{
  "total_results": 58,
  "total_pages": 2,
  "prev_url": null,
  "next_url": "",
  "resources": [
    {
      "entity": {
        "instance_index": 2,
        "exit_status": 2,
        "exit_description": "app instance was stopped",
        "timestamp": "2013-10-07T17:51:07+00:00"
      }
    }
  ]
}
`},
}

var notFoundRequest = testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/apps/my-app-guid/events",
	Response: testnet.TestResponse{
		Status: http.StatusNotFound,
	},
}

func TestListEvents(t *testing.T) {
	listEventsServer, handler := testnet.NewTLSServer(t, []testnet.TestRequest{
		firstPageEventsRequest,
		secondPageEventsRequest,
	})
	defer listEventsServer.Close()

	config := &configuration.Configuration{
		Target:      listEventsServer.URL,
		AccessToken: "BEARER my_access_token",
	}
	repo := NewCloudControllerAppEventsRepository(config, net.NewCloudControllerGateway())

	eventChan, apiErr := repo.ListEvents(cf.Application{Guid: "my-app-guid"})

	firstExpectedTime, err := time.Parse(APP_EVENT_TIMESTAMP_FORMAT, "2013-10-07T16:51:07+00:00")
	secondExpectedTime, err := time.Parse(APP_EVENT_TIMESTAMP_FORMAT, "2013-10-07T17:51:07+00:00")
	expectedEvents := []cf.Event{
		{
			InstanceIndex:   1,
			ExitStatus:      1,
			ExitDescription: "app instance exited",
			Timestamp:       firstExpectedTime,
		},
		{
			InstanceIndex:   2,
			ExitStatus:      2,
			ExitDescription: "app instance was stopped",
			Timestamp:       secondExpectedTime,
		},
	}

	list := []cf.Event{}
	for events := range eventChan {
		list = append(list, events...)
	}

	_, open := <-apiErr

	assert.NoError(t, err)
	assert.False(t, open)
	assert.Equal(t, list, expectedEvents)
	assert.True(t, handler.AllRequestsCalled())
}

func TestListEventsNotFound(t *testing.T) {

	listEventsServer, handler := testnet.NewTLSServer(t, []testnet.TestRequest{
		firstPageEventsRequest,
		notFoundRequest,
	})
	defer listEventsServer.Close()

	config := &configuration.Configuration{
		Target:      listEventsServer.URL,
		AccessToken: "BEARER my_access_token",
	}
	repo := NewCloudControllerAppEventsRepository(config, net.NewCloudControllerGateway())

	eventChan, apiErr := repo.ListEvents(cf.Application{Guid: "my-app-guid"})

	firstExpectedTime, err := time.Parse(APP_EVENT_TIMESTAMP_FORMAT, "2013-10-07T16:51:07+00:00")
	expectedEvents := []cf.Event{
		{
			InstanceIndex:   1,
			ExitStatus:      1,
			ExitDescription: "app instance exited",
			Timestamp:       firstExpectedTime,
		},
	}

	list := []cf.Event{}
	for events := range eventChan {
		list = append(list, events...)
	}

	apiResponse := <-apiErr

	assert.NoError(t, err)
	assert.Equal(t, list, expectedEvents)
	assert.True(t, apiResponse.IsNotSuccessful())
	assert.True(t, handler.AllRequestsCalled())
}
