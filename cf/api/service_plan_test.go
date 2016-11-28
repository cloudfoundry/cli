package api_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testnet "code.cloudfoundry.org/cli/util/testhelpers/net"

	. "code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Plan Repository", func() {
	var (
		testServer  *httptest.Server
		testHandler *testnet.TestHandler
		configRepo  coreconfig.ReadWriter
		repo        CloudControllerServicePlanRepository
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
		repo = NewCloudControllerServicePlanRepository(configRepo, gateway)
	})

	AfterEach(func() {
		testServer.Close()
	})

	setupTestServer := func(reqs ...testnet.TestRequest) {
		testServer, testHandler = testnet.NewServer(reqs)
		configRepo.SetAPIEndpoint(testServer.URL)
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
				Expect(servicePlansFields[0].GUID).To(Equal("the-big-guid"))
				Expect(servicePlansFields[0].Free).To(BeTrue())
				Expect(servicePlansFields[0].Public).To(BeTrue())
				Expect(servicePlansFields[0].Active).To(BeTrue())
				Expect(servicePlansFields[1].Name).To(Equal("The small second"))
				Expect(servicePlansFields[1].GUID).To(Equal("the-small-second"))
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
				Expect(servicePlansFields[0].GUID).To(Equal("the-big-guid"))
				Expect(servicePlansFields[0].Free).To(BeTrue())
				Expect(servicePlansFields[0].Public).To(BeTrue())
				Expect(servicePlansFields[0].Active).To(BeTrue())
				Expect(servicePlansFields[1].Name).To(Equal("The small second"))
				Expect(servicePlansFields[1].GUID).To(Equal("the-small-second"))
				Expect(servicePlansFields[1].Free).To(BeTrue())
				Expect(servicePlansFields[1].Public).To(BeFalse())
				Expect(servicePlansFields[1].Active).To(BeFalse())
			})
		})
	})

	Describe(".Update", func() {
		BeforeEach(func() {
			setupTestServer(apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/service_plans/my-service-plan-guid",
				Matcher:  testnet.RequestBodyMatcher(`{"public":true}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			}))
		})

		It("updates public on the service to whatever is passed", func() {
			servicePlan := models.ServicePlanFields{
				Name:        "my-service-plan",
				GUID:        "my-service-plan-guid",
				Description: "descriptive text",
				Free:        true,
				Public:      false,
			}

			err := repo.Update(servicePlan, "service-guid", true)
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe(".ListPlansFromManyServices", func() {
		BeforeEach(func() {
			setupTestServer(manyServiceRequest1, manyServiceRequest2)
		})

		It("returns all service plans for a list of service guids", func() {
			serviceGUIDs := []string{"service-guid1", "service-guid2"}

			servicePlansFields, err := repo.ListPlansFromManyServices(serviceGUIDs)
			Expect(err).NotTo(HaveOccurred())

			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(len(servicePlansFields)).To(Equal(2))

			Expect(servicePlansFields[0].Name).To(Equal("plan one"))
			Expect(servicePlansFields[0].GUID).To(Equal("plan1"))

			Expect(servicePlansFields[1].Name).To(Equal("plan two"))
			Expect(servicePlansFields[1].GUID).To(Equal("plan2"))
		})
	})
})

var firstPlanRequest = apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
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

var secondPlanRequest = apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
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

var firstPlanRequestWithParams = apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
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

var secondPlanRequestWithParams = apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/service_plans?q=service_guid%3AFoo&page=2",
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

var manyServiceRequest1 = apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/service_plans?q=service_guid+IN+service-guid1,service-guid2",
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `{
  "total_results": 2,
  "total_pages": 2,
  "next_url": "/v2/service_plans?q=service_guid+IN+service-guid1,service-guid2&page=2",
  "resources": [
    {
      "metadata": {
        "guid": "plan1"
      },
      "entity": {
        "name": "plan one",
        "free": true,
        "public": true,
        "active": true
      }
    }
  ]
}`,
	},
})

var manyServiceRequest2 = apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/service_plans?q=service_guid+IN+service-guid1,service-guid2&page=2",
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `{
  "total_results": 2,
  "total_pages": 1,
  "next_url": null,
  "prev_url": "/v2/service_plans?q=service_guid+IN+service-guid1,service-guid2",
  "resources": [
    {
      "metadata": {
        "guid": "plan2"
      },
      "entity": {
        "name": "plan two",
        "free": true,
        "public": true,
        "active": true
      }
    }
  ]
}`,
	},
})
