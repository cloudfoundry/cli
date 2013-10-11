package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var listEventsResponse = `
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
`
var secondPageEventsResponse = `
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
`

var listEventsEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	methodMatches := request.Method == "GET"
	pathMatches := request.URL.Path == "/v2/apps/my-app-guid/events"
	isSecondPage := strings.Contains(request.URL.RawQuery, "page=2")

	if !methodMatches || !pathMatches {
		fmt.Printf("One of the matchers did not match. Method [%t] Path [%t]",
			methodMatches, pathMatches)

		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	if isSecondPage {
		fmt.Fprint(writer, secondPageEventsResponse)
	} else {
		fmt.Fprint(writer, listEventsResponse)
	}
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

	assert.NoError(t, err)
	assert.True(t, apiErr.IsSuccessful())
	assert.Equal(t, list, expectedEvents)
}
