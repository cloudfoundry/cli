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
)

var _ = Describe("App Events Repo", func() {
	var (
		server  *httptest.Server
		handler *testnet.TestHandler
		config  configuration.ReadWriter
		gateway net.Gateway
		repo    AppEventsRepository
	)

	BeforeEach(func() {
		config = testconfig.NewRepository()
		config.SetAccessToken("BEARER my_access_token")

		gateway = net.NewCloudControllerGateway(config)
		repo = NewCloudControllerAppEventsRepository(config, gateway)
	})

	AfterEach(func() {
		server.Close()
	})

	setupTestServer := func(requests ...testnet.TestRequest) {
		server, handler = testnet.NewServer(requests)
		config.SetApiEndpoint(server.URL)
	}

	Describe("list recent events", func() {
		var recentAPIRequest testnet.TestRequest
		BeforeEach(func() {
			recentAPIRequest = newAPIRequestPage1
			recentAPIRequest.Path += "&order-direction=desc&results-per-page=2"
		})

		It("makes a request to the /v2/events endpoint", func() {
			setupTestServer(recentAPIRequest)

			list, err := repo.RecentEvents("my-app-guid", 2)
			Expect(err).ToNot(HaveOccurred())

			Expect(list).To(Equal([]models.EventFields{
				models.EventFields{
					Guid:        "event-1-guid",
					Name:        "audit.app.update",
					Timestamp:   testtime.MustParse(eventTimestampFormat, "2014-01-21T00:20:11+00:00"),
					Description: "instances: 1, memory: 256, command: PRIVATE DATA HIDDEN, environment_json: PRIVATE DATA HIDDEN",
				},
				models.EventFields{
					Guid:        "event-2-guid",
					Name:        "audit.app.update",
					Timestamp:   testtime.MustParse(eventTimestampFormat, "2014-01-21T00:20:11+00:00"),
					Description: "instances: 1, memory: 256, command: PRIVATE DATA HIDDEN, environment_json: PRIVATE DATA HIDDEN",
				},
			}))
		})

		It("makes a backwards compatible request to old events endpoint", func() {
			setupTestServer(newV2NotFoundRequest, oldAPIRequestPage1)

			_, err := repo.RecentEvents("my-app-guid", 2)
			Expect(err).ToNot(HaveOccurred())

			Expect(handler).To(testnet.HaveAllRequestsCalled())
		})
	})

	Describe("unmarshalling events", func() {
		It("unmarshals app crash events", func() {
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
			}`), &resource)

			Expect(err).NotTo(HaveOccurred())

			eventFields := resource.ToFields()
			Expect(eventFields.Guid).To(Equal("event-1-guid"))
			Expect(eventFields.Name).To(Equal("app.crash"))
			Expect(eventFields.Timestamp).To(Equal(testtime.MustParse(eventTimestampFormat, "2013-10-07T16:51:07+00:00")))
			Expect(eventFields.Description).To(Equal(`index: 3, reason: CRASHED, exit_description: unknown, exit_status: -1`))
		})

		It("unmarshals app update events", func() {
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
			}`), &resource)

			Expect(err).NotTo(HaveOccurred())

			eventFields := resource.ToFields()
			Expect(eventFields.Guid).To(Equal("event-1-guid"))
			Expect(eventFields.Name).To(Equal("audit.app.update"))
			Expect(eventFields.Timestamp).To(Equal(testtime.MustParse(eventTimestampFormat, "2014-01-21T00:20:11+00:00")))
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
			}`), &resource)

			Expect(err).NotTo(HaveOccurred())

			eventFields = resource.ToFields()
			Expect(eventFields.Guid).To(Equal("event-1-guid"))
			Expect(eventFields.Name).To(Equal("audit.app.update"))
			Expect(eventFields.Timestamp).To(Equal(testtime.MustParse(eventTimestampFormat, "2014-01-21T00:20:11+00:00")))
			Expect(eventFields.Description).To(Equal(`state: STOPPED`))
		})

		It("unmarshals app delete events", func() {
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
			}`), &resource)

			Expect(err).NotTo(HaveOccurred())

			eventFields := resource.ToFields()
			Expect(eventFields.Guid).To(Equal("event-2-guid"))
			Expect(eventFields.Name).To(Equal("audit.app.delete-request"))
			Expect(eventFields.Timestamp).To(Equal(testtime.MustParse(eventTimestampFormat, "2014-01-21T18:39:09+00:00")))
			Expect(eventFields.Description).To(Equal("recursive: true"))
		})

		It("unmarshals the new v2 app create event", func() {
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
			Expect(eventFields.Timestamp).To(Equal(testtime.MustParse(eventTimestampFormat, "2014-01-22T19:34:16+00:00")))
			Expect(eventFields.Description).To(Equal("disk_quota: 1024, instances: 1, state: STOPPED, environment_json: PRIVATE DATA HIDDEN"))
		})
	})
})

const eventTimestampFormat = "2006-01-02T15:04:05-07:00"

var oldAPIRequestPage1 = testnet.TestRequest{
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
			},
			{
			  "entity": {
				"instance_index": 2,
				"exit_status": 2,
				"exit_description": "app instance exited",
				"timestamp": "2013-11-07T16:51:07+00:00"
			  }
			}
		  ]
		}`}}

var newAPIRequestPage1 = testnet.TestRequest{
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
			},
			{
			  "metadata": {
				"guid": "event-2-guid"
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

var newV2NotFoundRequest = testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/events?q=actee%3Amy-app-guid",
	Response: testnet.TestResponse{
		Status: http.StatusNotFound,
	},
}
