package appevents_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/cloudfoundry/cli/cf/api/appevents"
	"github.com/cloudfoundry/cli/cf/api/strategy"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/testhelpers/cloudcontrollergateway"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	testtime "github.com/cloudfoundry/cli/testhelpers/time"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("App Events Repo", func() {
	var (
		server  *httptest.Server
		handler *testnet.TestHandler
		config  coreconfig.ReadWriter
		repo    Repository
	)

	BeforeEach(func() {
		config = testconfig.NewRepository()
		config.SetAccessToken("BEARER my_access_token")
		config.SetAPIVersion("2.2.0")
	})

	JustBeforeEach(func() {
		strategy := strategy.NewEndpointStrategy(config.APIVersion())
		gateway := cloudcontrollergateway.NewTestCloudControllerGateway(config)
		repo = NewCloudControllerAppEventsRepository(config, gateway, strategy)
	})

	AfterEach(func() {
		server.Close()
	})

	setupTestServer := func(requests ...testnet.TestRequest) {
		server, handler = testnet.NewServer(requests)
		config.SetAPIEndpoint(server.URL)
	}

	Describe("list recent events", func() {
		It("returns the most recent events", func() {
			setupTestServer(eventsRequest)

			list, err := repo.RecentEvents("my-app-guid", 2)
			Expect(err).ToNot(HaveOccurred())

			Expect(list).To(ConsistOf([]models.EventFields{
				{
					GUID:        "event-1-guid",
					Name:        "audit.app.update",
					Timestamp:   testtime.MustParse(eventTimestampFormat, "2014-01-21T00:20:11+00:00"),
					Description: "instances: 1, memory: 256, command: PRIVATE DATA HIDDEN, environment_json: PRIVATE DATA HIDDEN",
					Actor:       "cf-1-client",
					ActorName:   "somebody@pivotallabs.com",
				},
				{
					GUID:        "event-2-guid",
					Name:        "audit.app.update",
					Timestamp:   testtime.MustParse(eventTimestampFormat, "2014-01-21T00:20:11+00:00"),
					Description: "instances: 1, memory: 256, command: PRIVATE DATA HIDDEN, environment_json: PRIVATE DATA HIDDEN",
					Actor:       "cf-2-client",
					ActorName:   "nobody@pivotallabs.com",
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
				"actor": "cf-1-client",
				"actor_name": "somebody@pivotallabs.com",
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
				"actor": "cf-2-client",
				"actor_name": "nobody@pivotallabs.com",
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
