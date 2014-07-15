package api_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"

	. "github.com/cloudfoundry/cli/cf/api"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Plan Repository", func() {
	var (
		testServer  *httptest.Server
		testHandler *testnet.TestHandler
		configRepo  configuration.ReadWriter
		repo        CloudControllerServicePlanRepository
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		gateway := net.NewCloudControllerGateway((configRepo), time.Now)
		repo = NewCloudControllerServicePlanRepository(configRepo, gateway)
	})

	AfterEach(func() {
		testServer.Close()
	})

	setupTestServer := func(reqs ...testnet.TestRequest) {
		testServer, testHandler = testnet.NewServer(reqs)
		configRepo.SetApiEndpoint(testServer.URL)
	}

	Describe(".Search", func() {
		Context("No query parameters", func() {
			BeforeEach(func() {
				setupTestServer(firstPlanRequest, secondPlanRequest)
			})

			It("returns service plans", func() {
				servicePlansFields, err := repo.Search(map[string]string{})

				Expect(err).NotTo(HaveOccurred())
				Expect(testHandler).To(HaveAllRequestsCalled())
				Expect(len(servicePlansFields)).To(Equal(2))
				Expect(servicePlansFields[0].Name).To(Equal("The big one"))
				Expect(servicePlansFields[0].Guid).To(Equal("the-big-guid"))
				Expect(servicePlansFields[0].Free).To(BeTrue())
				Expect(servicePlansFields[0].Public).To(BeTrue())
				Expect(servicePlansFields[0].Active).To(BeTrue())
				Expect(servicePlansFields[1].Name).To(Equal("The small second"))
				Expect(servicePlansFields[1].Guid).To(Equal("the-small-second"))
				Expect(servicePlansFields[1].Free).To(BeTrue())
				Expect(servicePlansFields[1].Public).To(BeFalse())
				Expect(servicePlansFields[1].Active).To(BeFalse())
			})
		})
		Context("With query parameters", func() {
			BeforeEach(func() {
				setupTestServer(firstPlanRequestWithParams, secondPlanRequestWithParams)
			})

			It("returns service plans", func() {
				servicePlansFields, err := repo.Search(map[string]string{"service_guid": "Foo"})

				Expect(err).NotTo(HaveOccurred())
				Expect(testHandler).To(HaveAllRequestsCalled())
				Expect(len(servicePlansFields)).To(Equal(2))
				Expect(servicePlansFields[0].Name).To(Equal("The big one"))
				Expect(servicePlansFields[0].Guid).To(Equal("the-big-guid"))
				Expect(servicePlansFields[0].Free).To(BeTrue())
				Expect(servicePlansFields[0].Public).To(BeTrue())
				Expect(servicePlansFields[0].Active).To(BeTrue())
				Expect(servicePlansFields[1].Name).To(Equal("The small second"))
				Expect(servicePlansFields[1].Guid).To(Equal("the-small-second"))
				Expect(servicePlansFields[1].Free).To(BeTrue())
				Expect(servicePlansFields[1].Public).To(BeFalse())
				Expect(servicePlansFields[1].Active).To(BeFalse())
			})
		})
	})
})

var firstPlanRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/service_plans",
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `{
  "total_results": 2,
  "total_pages": 2,
  "next_url": "/v2/service_plans?page=2",
  "resources": [
    {
      "metadata": {
        "guid": "the-big-guid"
      },
      "entity": {
        "name": "The big one",
        "free": true,
        "public": true,
        "active": true
      }
    }
  ]
}`,
	},
})

var secondPlanRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/service_plans?page=2",
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `{
  "total_results": 2,
  "total_pages": 2,
  "resources": [
    {
      "metadata": {
        "guid": "the-small-second"
      },
      "entity": {
        "name": "The small second",
        "free": true,
        "public": false,
        "active": false
      }
    }
  ]
}`,
	},
})

var firstPlanRequestWithParams = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/service_plans?q=service_guid%3AFoo",
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `{
  "total_results": 2,
  "total_pages": 2,
  "next_url": "/v2/service_plans?q=service_guid%3AFoo&page=2",
  "resources": [
    {
      "metadata": {
        "guid": "the-big-guid"
      },
      "entity": {
        "name": "The big one",
        "free": true,
        "public": true,
        "active": true
      }
    }
  ]
}`,
	},
})

var secondPlanRequestWithParams = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/service_plans??q=service_guid%3AFoo&page=2",
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `{
  "total_results": 2,
  "total_pages": 2,
  "resources": [
    {
      "metadata": {
        "guid": "the-small-second"
      },
      "entity": {
        "name": "The small second",
        "free": true,
        "public": false,
        "active": false
      }
    }
  ]
}`,
	},
})
