package api_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"

	. "github.com/cloudfoundry/cli/cf/api"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Plan Visibility Repository", func() {
	var (
		testServer  *httptest.Server
		testHandler *testnet.TestHandler
		configRepo  core_config.ReadWriter
		repo        CloudControllerServicePlanVisibilityRepository
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		gateway := net.NewCloudControllerGateway((configRepo), time.Now)
		repo = NewCloudControllerServicePlanVisibilityRepository(configRepo, gateway)
	})

	AfterEach(func() {
		testServer.Close()
	})

	setupTestServer := func(reqs ...testnet.TestRequest) {
		testServer, testHandler = testnet.NewServer(reqs)
		configRepo.SetApiEndpoint(testServer.URL)
	}

	Describe(".Create", func() {
		BeforeEach(func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "POST",
				Path:     "/v2/service_plan_visibilities",
				Matcher:  testnet.RequestBodyMatcher(`{"service_plan_guid":"service_plan_guid", "organization_guid":"org_guid"}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			}))
		})

		It("creates a service plan visibility", func() {
			err := repo.Create("service_plan_guid", "org_guid")

			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe(".List", func() {
		BeforeEach(func() {
			setupTestServer(firstPlanVisibilityRequest, secondPlanVisibilityRequest)
		})

		It("returns service plans", func() {
			servicePlansVisibilitiesFields, err := repo.List()

			Expect(err).NotTo(HaveOccurred())
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(len(servicePlansVisibilitiesFields)).To(Equal(2))
			Expect(servicePlansVisibilitiesFields[0].Guid).To(Equal("request-guid-1"))
			Expect(servicePlansVisibilitiesFields[0].ServicePlanGuid).To(Equal("service-plan-guid-1"))
			Expect(servicePlansVisibilitiesFields[0].OrganizationGuid).To(Equal("org-guid-1"))
			Expect(servicePlansVisibilitiesFields[1].Guid).To(Equal("request-guid-2"))
			Expect(servicePlansVisibilitiesFields[1].ServicePlanGuid).To(Equal("service-plan-guid-2"))
			Expect(servicePlansVisibilitiesFields[1].OrganizationGuid).To(Equal("org-guid-2"))
		})
	})

	Describe(".Delete", func() {
		It("deletes a service plan visibility", func() {
			servicePlanVisibilityGuid := "the-service-plan-visibility-guid"
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "DELETE",
				Path:   "/v2/service_plan_visibilities/" + servicePlanVisibilityGuid,
				Response: testnet.TestResponse{
					Status: http.StatusNoContent,
				},
			}))

			err := repo.Delete(servicePlanVisibilityGuid)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe(".Search", func() {
		It("finds the service plan visibilities that match the given query parameters", func() {
			setupTestServer(searchPlanVisibilityRequest)

			servicePlansVisibilitiesFields, err := repo.Search(map[string]string{"service_plan_guid": "service-plan-guid-1", "organization_guid": "org-guid-1"})
			Expect(err).ToNot(HaveOccurred())
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(len(servicePlansVisibilitiesFields)).To(Equal(1))
			Expect(servicePlansVisibilitiesFields[0].Guid).To(Equal("request-guid-1"))
			Expect(servicePlansVisibilitiesFields[0].ServicePlanGuid).To(Equal("service-plan-guid-1"))
			Expect(servicePlansVisibilitiesFields[0].OrganizationGuid).To(Equal("org-guid-1"))
		})
	})
})

var firstPlanVisibilityRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/service_plan_visibilities",
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `{
  "total_results": 2,
  "total_pages": 2,
  "next_url": "/v2/service_plan_visibilities?page=2",
  "resources": [
    {
      "metadata": {
        "guid": "request-guid-1"
      },
      "entity": {
        "service_plan_guid": "service-plan-guid-1",
        "organization_guid": "org-guid-1"
      }
    }
  ]
}`,
	},
})

var secondPlanVisibilityRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/service_plan_visibilities?page=2",
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `{
  "total_results": 2,
  "total_pages": 2,
  "resources": [
    {
      "metadata": {
        "guid": "request-guid-2"
      },
      "entity": {
        "service_plan_guid": "service-plan-guid-2",
        "organization_guid": "org-guid-2"
      }
    }
  ]
}`,
	},
})

var searchPlanVisibilityRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/service_plan_visibilities?q=service_plan_guid%3Aservice-plan-guid-1%3Borganization_guid%3Aorg-guid-1",
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `{
  "total_results": 1,
  "total_pages": 1,
  "resources": [
    {
      "metadata": {
        "guid": "request-guid-1"
      },
      "entity": {
        "service_plan_guid": "service-plan-guid-1",
        "organization_guid": "org-guid-1"
      }
    }
  ]
}`,
	},
})
