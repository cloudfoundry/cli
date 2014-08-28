package feature_flag_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"

	. "github.com/cloudfoundry/cli/cf/api/feature_flag"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature Flag Repository", func() {
	var (
		testServer  *httptest.Server
		testHandler *testnet.TestHandler
		configRepo  configuration.ReadWriter
		repo        CloudControllerFeatureFlagRepository
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		gateway := net.NewCloudControllerGateway((configRepo), time.Now)
		repo = NewCloudControllerFeatureFlagRepository(configRepo, gateway)
	})

	AfterEach(func() {
		testServer.Close()
	})

	setupTestServer := func(reqs ...testnet.TestRequest) {
		testServer, testHandler = testnet.NewServer(reqs)
		configRepo.SetApiEndpoint(testServer.URL)
	}

	Describe(".List", func() {
		BeforeEach(func() {
			setupTestServer(featureFlagsGetAllRequest)
		})

		It("returns all of the feature flags", func() {
			featureFlagModels, err := repo.List()

			Expect(err).NotTo(HaveOccurred())
			Expect(testHandler).To(HaveAllRequestsCalled())
			Expect(len(featureFlagModels)).To(Equal(5))
			Expect(featureFlagModels[0].Name).To(Equal("user_org_creation"))
			Expect(featureFlagModels[0].Enabled).To(BeFalse())
			Expect(featureFlagModels[1].Name).To(Equal("private_domain_creation"))
			Expect(featureFlagModels[1].Enabled).To(BeFalse())
			Expect(featureFlagModels[2].Name).To(Equal("app_bits_upload"))
			Expect(featureFlagModels[2].Enabled).To(BeTrue())
			Expect(featureFlagModels[3].Name).To(Equal("app_scaling"))
			Expect(featureFlagModels[3].Enabled).To(BeTrue())
			Expect(featureFlagModels[4].Name).To(Equal("route_creation"))
			Expect(featureFlagModels[4].Enabled).To(BeTrue())
		})
	})
})

var featureFlagsGetAllRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/config/feature_flags",
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `[
    { 
      "name": "user_org_creation",
      "enabled": false,
      "error_message": null,
      "url": "/v2/config/feature_flags/user_org_creation"
    },
    { 
      "name": "private_domain_creation",
      "enabled": false,
      "error_message": "foobar",
      "url": "/v2/config/feature_flags/private_domain_creation"
    },
    { 
      "name": "app_bits_upload",
      "enabled": true,
      "error_message": null,
      "url": "/v2/config/feature_flags/app_bits_upload"
    },
    { 
      "name": "app_scaling",
      "enabled": true,
      "error_message": null,
      "url": "/v2/config/feature_flags/app_scaling"
    },
    { 
      "name": "route_creation",
      "enabled": true,
      "error_message": null,
      "url": "/v2/config/feature_flags/route_creation"
    }
]`,
	},
})
