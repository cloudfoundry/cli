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

var firstPageOldV2EventsRequest = testnet.TestRequest{
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

var secondPageOldV2EventsRequest = testnet.TestRequest{
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

var oldV2NotFoundRequest = testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/apps/my-app-guid/events",
	Response: testnet.TestResponse{
		Status: http.StatusNotFound,
	},
}

func TestListEvents(t *testing.T) {
	listEventsServer, handler := testnet.NewTLSServer(t, []testnet.TestRequest{
		firstPageOldV2EventsRequest,
		secondPageOldV2EventsRequest,
	})
	defer listEventsServer.Close()

	config := &configuration.Configuration{
		Target:      listEventsServer.URL,
		AccessToken: "BEARER my_access_token",
	}
	repo := NewCloudControllerAppEventsRepository(config, net.NewCloudControllerGateway())

	time1, _ := time.Parse(APP_EVENT_TIMESTAMP_FORMAT, "2013-10-07T16:51:07+00:00")
	time2, _ := time.Parse(APP_EVENT_TIMESTAMP_FORMAT, "2013-10-07T17:51:07+00:00")

	event1 := cf.EventFields{}
	event1.Name = "app crashed"
	event1.InstanceIndex = 1
	event1.Description = "reason: app instance exited, exit_status: 1"
	event1.Timestamp = time1

	event2 := cf.EventFields{}
	event2.Name = "app crashed"
	event2.InstanceIndex = 2
	event2.Description = "reason: app instance was stopped, exit_status: 2"
	event2.Timestamp = time2

	expectedEvents := []cf.EventFields{
		event1,
		event2,
	}

	list := []cf.EventFields{}
	apiResponse := repo.ListEvents("my-app-guid", ListEventsCallback(func(events []cf.EventFields) (fetchNext bool) {
		list = append(list, events...)
		return true
	}))

	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, list, expectedEvents)
	assert.True(t, handler.AllRequestsCalled())
}

func TestListEventsNotFound(t *testing.T) {

	listEventsServer, handler := testnet.NewTLSServer(t, []testnet.TestRequest{
		firstPageOldV2EventsRequest,
		oldV2NotFoundRequest,
	})
	defer listEventsServer.Close()

	config := &configuration.Configuration{
		Target:      listEventsServer.URL,
		AccessToken: "BEARER my_access_token",
	}
	repo := NewCloudControllerAppEventsRepository(config, net.NewCloudControllerGateway())

	firstExpectedTime, err := time.Parse(APP_EVENT_TIMESTAMP_FORMAT, "2013-10-07T16:51:07+00:00")

	event1 := cf.EventFields{}
	event1.Name = "app crashed"
	event1.InstanceIndex = 1
	event1.Description = "reason: app instance exited, exit_status: 1"
	event1.Timestamp = firstExpectedTime

	expectedEvents := []cf.EventFields{
		event1,
	}

	list := []cf.EventFields{}
	apiResponse := repo.ListEvents("my-app-guid", ListEventsCallback(func(events []cf.EventFields) (fetchNext bool) {
		list = append(list, events...)
		return true
	}))

	assert.NoError(t, err)
	assert.Equal(t, list, expectedEvents)
	assert.True(t, apiResponse.IsNotSuccessful())
	assert.True(t, handler.AllRequestsCalled())
}
