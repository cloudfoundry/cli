package api_test

import (
	. "cf/api"
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
