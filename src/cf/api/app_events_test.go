package api_test

import (
	"cf"
	. "cf/api"
	"cf/configuration"
	"cf/net"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	testnet "testhelpers/net"
	testtime "testhelpers/time"

	. "github.com/onsi/ginkgo"
	mr "github.com/tjarratt/mr_t"
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

var newV2NotFoundRequest = testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/events?q=actee%3Amy-app-guid",
	Response: testnet.TestResponse{
		Status: http.StatusNotFound,
	},
}

type eventTestDependencies struct {
	server  *httptest.Server
	handler *testnet.TestHandler
	config  *configuration.Configuration
	gateway net.Gateway
}

func setupEventTest(t mr.TestingT, requests []testnet.TestRequest) (deps eventTestDependencies) {
	deps.server, deps.handler = testnet.NewTLSServer(t, requests)

	deps.config = &configuration.Configuration{
		Target:      deps.server.URL,
		AccessToken: "BEARER my_access_token",
	}

	deps.gateway = net.NewCloudControllerGateway()

	return
}

func teardownEventTest(deps eventTestDependencies) {
	deps.server.Close()
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestListOldV2EventsWhenNewV2ApiNotFound", func() {
			deps := setupEventTest(mr.T(), []testnet.TestRequest{
				newV2NotFoundRequest,
				firstPageOldV2EventsRequest,
				secondPageOldV2EventsRequest,
			})
			defer teardownEventTest(deps)

			repo := NewCloudControllerAppEventsRepository(deps.config, deps.gateway)

			event1 := cf.EventFields{}
			event1.Name = "app crashed"
			event1.Description = "instance: 1, reason: app instance exited, exit_status: 1"
			event1.Timestamp = testtime.MustParse(APP_EVENT_TIMESTAMP_FORMAT, "2013-10-07T16:51:07+00:00")

			event2 := cf.EventFields{}
			event2.Name = "app crashed"
			event2.Description = "instance: 2, reason: app instance was stopped, exit_status: 2"
			event2.Timestamp = testtime.MustParse(APP_EVENT_TIMESTAMP_FORMAT, "2013-10-07T17:51:07+00:00")

			expectedEvents := []cf.EventFields{
				event1,
				event2,
			}

			list := []cf.EventFields{}
			apiResponse := repo.ListEvents("my-app-guid", ListEventsCallback(func(events []cf.EventFields) (fetchNext bool) {
				list = append(list, events...)
				return true
			}))

			assert.True(mr.T(), apiResponse.IsSuccessful())
			assert.Equal(mr.T(), list, expectedEvents)
			assert.True(mr.T(), deps.handler.AllRequestsCalled())
		})
		It("TestListEventsUsingNewEndpoint", func() {

			pageOneNewV2Request := testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/events?q=actee%3Amy-app-guid",
				Response: testnet.TestResponse{
					Status: http.StatusOK,
					Body: `{
			  "total_results": 1,
			  "total_pages": 1,
			  "prev_url": null,
			  "next_url": "/v2/events?q=actee%3Amy-app-guid&page=2",
			  "resources": [
    			{
				  "metadata": {
					"guid": "event-1-guid"
				  },
				  "entity": {
					"type": "audit.app.update",
					"timestamp": "2014-01-21T00:20:11+00:00",
					"metadata": {
					  "request": {
						"command": "PRIVATE DATA HIDDEN",
						"instances": 1,
						"memory": 256,
						"name": "dora",
						"environment_json": "PRIVATE DATA HIDDEN"
					  }
					}
				  }
				}
			  ]
			}`}}

			pageTwoNewV2Request := testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/events?q=actee%3Amy-app-guid&page=2",
				Response: testnet.TestResponse{
					Status: http.StatusOK,
					Body: `{
			  "total_results": 1,
			  "total_pages": 1,
			  "prev_url": null,
			  "next_url": "",
			  "resources": [
				{
				  "metadata":{
				    "guid":"event-2-guid"
				  },
				  "entity": {
					"type": "app.crash",
					"timestamp": "2013-10-07T17:51:07+00:00",
					"metadata": {}
				  }
				}
			  ]
			}`}}

			deps := setupEventTest(mr.T(), []testnet.TestRequest{
				pageOneNewV2Request,
				pageTwoNewV2Request,
			})
			defer teardownEventTest(deps)

			repo := NewCloudControllerAppEventsRepository(deps.config, deps.gateway)

			events := []cf.EventFields{}
			apiResponse := repo.ListEvents("my-app-guid", ListEventsCallback(func(eventPage []cf.EventFields) (fetchNext bool) {
				events = append(events, eventPage...)
				return true
			}))

			assert.True(mr.T(), apiResponse.IsSuccessful())
			assert.True(mr.T(), deps.handler.AllRequestsCalled())

			assert.Equal(mr.T(), len(events), 2)
			assert.Equal(mr.T(), events[0].Guid, "event-1-guid")
			assert.Equal(mr.T(), events[0].Name, "audit.app.update")
			assert.Equal(mr.T(), events[1].Guid, "event-2-guid")
			assert.Equal(mr.T(), events[1].Name, "app.crash")
		})
		It("TestListOldV2EventsApiError", func() {

			deps := setupEventTest(mr.T(), []testnet.TestRequest{
				newV2NotFoundRequest,
				firstPageOldV2EventsRequest,
				oldV2NotFoundRequest,
			})
			defer teardownEventTest(deps)

			repo := NewCloudControllerAppEventsRepository(deps.config, deps.gateway)

			list := []cf.EventFields{}
			apiResponse := repo.ListEvents("my-app-guid", ListEventsCallback(func(events []cf.EventFields) (fetchNext bool) {
				list = append(list, events...)
				return true
			}))

			firstExpectedTime, err := time.Parse(APP_EVENT_TIMESTAMP_FORMAT, "2013-10-07T16:51:07+00:00")
			assert.NoError(mr.T(), err)

			event1 := cf.EventFields{}
			event1.Name = "app crashed"
			event1.Description = "instance: 1, reason: app instance exited, exit_status: 1"
			event1.Timestamp = firstExpectedTime

			expectedEvents := []cf.EventFields{
				event1,
			}

			assert.Equal(mr.T(), list, expectedEvents)
			assert.True(mr.T(), apiResponse.IsNotSuccessful())
			assert.True(mr.T(), deps.handler.AllRequestsCalled())
		})
		It("TestUnmarshalNewCrashEvent", func() {

			resource := new(EventResourceNewV2)
			err := json.Unmarshal([]byte(`
	{
	  "metadata": {
	  	"guid":"event-1-guid"
	  },
	  "entity": {
	  	"timestamp": "2013-10-07T16:51:07+00:00",
	  	"type": "app.crash",
	  	"metadata": {
	  	  "instance": "50dd66d3f8874b35988d23a25d19bfa0",
	  	  "index": 3,
	  	  "exit_status": -1,
	  	  "exit_description": "unknown",
	  	  "reason": "CRASHED"
	  	}
	  }
	}
	`), &resource)

			assert.NoError(mr.T(), err)

			eventFields := resource.ToFields().(cf.EventFields)
			assert.Equal(mr.T(), eventFields.Guid, "event-1-guid")
			assert.Equal(mr.T(), eventFields.Name, "app.crash")
			assert.Equal(mr.T(), eventFields.Timestamp, testtime.MustParse(APP_EVENT_TIMESTAMP_FORMAT, "2013-10-07T16:51:07+00:00"))
			assert.Equal(mr.T(), eventFields.Description, `index: 3, reason: CRASHED, exit_description: unknown, exit_status: -1`)
		})
		It("TestUnmarshalUpdateAppEvent", func() {

			resource := new(EventResourceNewV2)
			err := json.Unmarshal([]byte(`
    {
      "metadata": {
        "guid": "event-1-guid"
      },
      "entity": {
        "type": "audit.app.update",
        "timestamp": "2014-01-21T00:20:11+00:00",
        "metadata": {
          "request": {
            "command": "PRIVATE DATA HIDDEN",
            "instances": 1,
            "memory": 256,
            "environment_json": "PRIVATE DATA HIDDEN"
          }
        }
      }
    }
	`), &resource)

			assert.NoError(mr.T(), err)

			eventFields := resource.ToFields().(cf.EventFields)
			assert.Equal(mr.T(), eventFields.Guid, "event-1-guid")
			assert.Equal(mr.T(), eventFields.Name, "audit.app.update")
			assert.Equal(mr.T(), eventFields.Timestamp, testtime.MustParse(APP_EVENT_TIMESTAMP_FORMAT, "2014-01-21T00:20:11+00:00"))
			assert.Equal(mr.T(), eventFields.Description, "instances: 1, memory: 256, command: PRIVATE DATA HIDDEN, environment_json: PRIVATE DATA HIDDEN")

			resource = new(EventResourceNewV2)
			err = json.Unmarshal([]byte(`
    {
      "metadata": {
        "guid": "event-1-guid"
      },
      "entity": {
        "type": "audit.app.update",
        "timestamp": "2014-01-21T00:20:11+00:00",
        "metadata": {
          "request": {
          	"state": "STOPPED"
          }
        }
      }
    }
	`), &resource)

			assert.NoError(mr.T(), err)

			eventFields = resource.ToFields().(cf.EventFields)
			assert.Equal(mr.T(), eventFields.Guid, "event-1-guid")
			assert.Equal(mr.T(), eventFields.Name, "audit.app.update")
			assert.Equal(mr.T(), eventFields.Timestamp, testtime.MustParse(APP_EVENT_TIMESTAMP_FORMAT, "2014-01-21T00:20:11+00:00"))
			assert.Equal(mr.T(), eventFields.Description, `state: STOPPED`)
		})
		It("TestUnmarshalDeleteAppEvent", func() {

			resource := new(EventResourceNewV2)
			err := json.Unmarshal([]byte(`
    {
      "metadata": {
        "guid": "event-2-guid"
      },
      "entity": {
        "type": "audit.app.delete-request",
        "timestamp": "2014-01-21T18:39:09+00:00",
        "metadata": {
          "request": {
            "recursive": true
          }
        }
      }
    }
	`), &resource)

			assert.NoError(mr.T(), err)

			eventFields := resource.ToFields().(cf.EventFields)
			assert.Equal(mr.T(), eventFields.Guid, "event-2-guid")
			assert.Equal(mr.T(), eventFields.Name, "audit.app.delete-request")
			assert.Equal(mr.T(), eventFields.Timestamp, testtime.MustParse(APP_EVENT_TIMESTAMP_FORMAT, "2014-01-21T18:39:09+00:00"))
			assert.Equal(mr.T(), eventFields.Description, "recursive: true")
		})
		It("TestUnmarshalNewV2CreateEvent", func() {

			resource := new(EventResourceNewV2)
			err := json.Unmarshal([]byte(`
	{
      "metadata": {
        "guid": "event-1-guid"
      },
      "entity": {
        "type": "audit.app.create",
        "timestamp": "2014-01-22T19:34:16+00:00",
        "metadata": {
          "request": {
            "name": "java-warz",
            "space_guid": "6cc20fec-0dee-4843-b875-b124bfee791a",
            "production": false,
            "environment_json": "PRIVATE DATA HIDDEN",
            "instances": 1,
            "disk_quota": 1024,
            "state": "STOPPED",
            "console": false
          }
        }
      }
	}`), &resource)

			assert.NoError(mr.T(), err)

			eventFields := resource.ToFields().(cf.EventFields)
			assert.Equal(mr.T(), eventFields.Guid, "event-1-guid")
			assert.Equal(mr.T(), eventFields.Name, "audit.app.create")
			assert.Equal(mr.T(), eventFields.Timestamp, testtime.MustParse(APP_EVENT_TIMESTAMP_FORMAT, "2014-01-22T19:34:16+00:00"))
			assert.Equal(mr.T(), eventFields.Description, "disk_quota: 1024, instances: 1, state: STOPPED, environment_json: PRIVATE DATA HIDDEN")
		})
	})
}
