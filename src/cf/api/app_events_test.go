package api_test

import (
	. "cf/api"
	"cf/api/strategy"
	"cf/configuration"
	"cf/models"
	"cf/net"
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
		repo    AppEventsRepository
	)

	BeforeEach(func() {
		config = testconfig.NewRepository()
		config.SetAccessToken("BEARER my_access_token")
		config.SetApiVersion("2.2.0")
	})

	JustBeforeEach(func() {
		strategy := strategy.NewEndpointStrategy(config.ApiVersion())
		gateway := net.NewCloudControllerGateway(config)
		repo = NewCloudControllerAppEventsRepository(config, gateway, strategy)
	})

	AfterEach(func() {
		server.Close()
	})

	setupTestServer := func(requests ...testnet.TestRequest) {
		server, handler = testnet.NewServer(requests)
		config.SetApiEndpoint(server.URL)
	}

	Describe("list recent events", func() {
		It("returns the most recent events", func() {
			setupTestServer(eventsRequest)

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
	})
})

const eventTimestampFormat = "2006-01-02T15:04:05-07:00"

var eventsRequest = testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/events?q=actee%3Amy-app-guid&order-direction=desc&results-per-page=2",
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
