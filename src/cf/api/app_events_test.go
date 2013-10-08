package api_test

import (
	"cf"
	. "cf/api"
	"cf/configuration"
	"cf/net"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var listEventsResponse = `
{
  "total_results": 58,
  "total_pages": 2,
  "prev_url": null,
  "next_url": "/v2/apps/8138e7f7-1f45-4d1b-ba70-160525dfe107/events?inline-relations-depth=1&page=2&results-per-page=50",
  "resources": [
    {
      "entity": {
        "instance_index": 0,
        "exit_status": 1,
        "exit_description": "app instance exited",
        "timestamp": "2013-10-07T16:51:07+00:00"
      }
    }
  ]
}
`
var listEventsEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	methodMatches := request.Method == "GET"
	pathMatches := request.URL.Path == "/v2/apps/my-app-guid/events"

	if !methodMatches || !pathMatches {
		fmt.Printf("One of the matchers did not match. Method [%t] Path [%t]",
			methodMatches, pathMatches)

		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	fmt.Fprint(writer, listEventsResponse)
}

func TestListEvents(t *testing.T) {

	listEventsServer := httptest.NewTLSServer(http.HandlerFunc(listEventsEndpoint))
	defer listEventsServer.Close()

	config := &configuration.Configuration{
		Target:      listEventsServer.URL,
		AccessToken: "BEARER my_access_token",
	}
	repo := NewCloudControllerAppEventsRepository(config, net.NewCloudControllerGateway())

	list, apiErr := repo.ListEvents(cf.Application{Guid: "my-app-guid"})

	expectedTime, err := time.Parse(APP_EVENT_TIMESTAMP_FORMAT, "2013-10-07T16:51:07+00:00")
	expectedEvents := []cf.Event{
		{
			InstanceIndex:   0,
			ExitStatus:      1,
			ExitDescription: "app instance exited",
			Timestamp:       expectedTime,
		},
	}

	assert.NoError(t, err)
	assert.False(t, apiErr.IsNotSuccessful())
	assert.Equal(t, list, expectedEvents)
}
