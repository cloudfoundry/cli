package api_test

import (
	. "cf/api"
	"cf/configuration"
	"cf/models"
	"cf/net"
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
	testtime "testhelpers/time"
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
	config  configuration.Reader
	gateway net.Gateway
}

func setupEventTest(requests []testnet.TestRequest) (deps eventTestDependencies) {
	deps.server, deps.handler = testnet.NewTLSServer(requests)

	configRepo := testconfig.NewRepository()
	configRepo.SetApiEndpoint(deps.server.URL)
	configRepo.SetAccessToken("BEARER my_access_token")

	deps.config = configRepo
	deps.gateway = net.NewCloudControllerGateway()

	return
}

func teardownEventTest(deps eventTestDependencies) {
	deps.server.Close()
}

var _ = Describe("App Events Repo", func() {
	It("TestListOldV2EventsWhenNewV2ApiNotFound", func() {
		deps := setupEventTest([]testnet.TestRequest{
			newV2NotFoundRequest,
			firstPageOldV2EventsRequest,
			secondPageOldV2EventsRequest,
		})
		defer teardownEventTest(deps)

		repo := NewCloudControllerAppEventsRepository(deps.config, deps.gateway)

		expectedEvents := []models.EventFields{
			models.EventFields{
				Name:        "app crashed",
				Description: "instance: 1, reason: app instance exited, exit_status: 1",
				Timestamp:   testtime.MustParse(APP_EVENT_TIMESTAMP_FORMAT, "2013-10-07T16:51:07+00:00"),
			},
			models.EventFields{
				Name:        "app crashed",
				Description: "instance: 2, reason: app instance was stopped, exit_status: 2",
				Timestamp:   testtime.MustParse(APP_EVENT_TIMESTAMP_FORMAT, "2013-10-07T17:51:07+00:00"),
			},
		}

		list := []models.EventFields{}
		apiResponse := repo.ListEvents("my-app-guid", func(event models.EventFields) bool {
			list = append(list, event)
			return true
		})

		Expect(apiResponse.IsSuccessful()).To(BeTrue())
		Expect(list).To(Equal(expectedEvents))
		Expect(deps.handler.AllRequestsCalled()).To(BeTrue())
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

		deps := setupEventTest([]testnet.TestRequest{
			pageOneNewV2Request,
			pageTwoNewV2Request,
		})
		defer teardownEventTest(deps)

		repo := NewCloudControllerAppEventsRepository(deps.config, deps.gateway)

		events := []models.EventFields{}
		apiResponse := repo.ListEvents("my-app-guid", func(e models.EventFields) bool {
			events = append(events, e)
			return true
		})

		Expect(apiResponse.IsSuccessful()).To(BeTrue())
		Expect(deps.handler.AllRequestsCalled()).To(BeTrue())

		Expect(len(events)).To(Equal(2))
		Expect(events[0].Guid).To(Equal("event-1-guid"))
		Expect(events[0].Name).To(Equal("audit.app.update"))
		Expect(events[1].Guid).To(Equal("event-2-guid"))
		Expect(events[1].Name).To(Equal("app.crash"))
	})

	It("TestListOldV2EventsApiError", func() {
		deps := setupEventTest([]testnet.TestRequest{
			newV2NotFoundRequest,
			firstPageOldV2EventsRequest,
			oldV2NotFoundRequest,
		})
		defer teardownEventTest(deps)

		repo := NewCloudControllerAppEventsRepository(deps.config, deps.gateway)

		list := []models.EventFields{}
		apiResponse := repo.ListEvents("my-app-guid", func(e models.EventFields) bool {
			list = append(list, e)
			return true
		})

		firstExpectedTime, err := time.Parse(APP_EVENT_TIMESTAMP_FORMAT, "2013-10-07T16:51:07+00:00")
		Expect(err).NotTo(HaveOccurred())

		expectedEvents := []models.EventFields{
			models.EventFields{
				Name:        "app crashed",
				Description: "instance: 1, reason: app instance exited, exit_status: 1",
				Timestamp:   firstExpectedTime,
			},
		}

		Expect(list).To(Equal(expectedEvents))
		Expect(apiResponse.IsNotSuccessful()).To(BeTrue())
		Expect(deps.handler.AllRequestsCalled()).To(BeTrue())
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

		Expect(err).NotTo(HaveOccurred())

		eventFields := resource.ToFields()
		Expect(eventFields.Guid).To(Equal("event-1-guid"))
		Expect(eventFields.Name).To(Equal("app.crash"))
		Expect(eventFields.Timestamp).To(Equal(testtime.MustParse(APP_EVENT_TIMESTAMP_FORMAT, "2013-10-07T16:51:07+00:00")))
		Expect(eventFields.Description).To(Equal(`index: 3, reason: CRASHED, exit_description: unknown, exit_status: -1`))
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

		Expect(err).NotTo(HaveOccurred())

		eventFields := resource.ToFields()
		Expect(eventFields.Guid).To(Equal("event-1-guid"))
		Expect(eventFields.Name).To(Equal("audit.app.update"))
		Expect(eventFields.Timestamp).To(Equal(testtime.MustParse(APP_EVENT_TIMESTAMP_FORMAT, "2014-01-21T00:20:11+00:00")))
		Expect(eventFields.Description).To(Equal("instances: 1, memory: 256, command: PRIVATE DATA HIDDEN, environment_json: PRIVATE DATA HIDDEN"))

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

		Expect(err).NotTo(HaveOccurred())

		eventFields = resource.ToFields()
		Expect(eventFields.Guid).To(Equal("event-1-guid"))
		Expect(eventFields.Name).To(Equal("audit.app.update"))
		Expect(eventFields.Timestamp).To(Equal(testtime.MustParse(APP_EVENT_TIMESTAMP_FORMAT, "2014-01-21T00:20:11+00:00")))
		Expect(eventFields.Description).To(Equal(`state: STOPPED`))
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

		Expect(err).NotTo(HaveOccurred())

		eventFields := resource.ToFields()
		Expect(eventFields.Guid).To(Equal("event-2-guid"))
		Expect(eventFields.Name).To(Equal("audit.app.delete-request"))
		Expect(eventFields.Timestamp).To(Equal(testtime.MustParse(APP_EVENT_TIMESTAMP_FORMAT, "2014-01-21T18:39:09+00:00")))
		Expect(eventFields.Description).To(Equal("recursive: true"))
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

		Expect(err).NotTo(HaveOccurred())

		eventFields := resource.ToFields()
		Expect(eventFields.Guid).To(Equal("event-1-guid"))
		Expect(eventFields.Name).To(Equal("audit.app.create"))
		Expect(eventFields.Timestamp).To(Equal(testtime.MustParse(APP_EVENT_TIMESTAMP_FORMAT, "2014-01-22T19:34:16+00:00")))
		Expect(eventFields.Description).To(Equal("disk_quota: 1024, instances: 1, state: STOPPED, environment_json: PRIVATE DATA HIDDEN"))
	})
})
