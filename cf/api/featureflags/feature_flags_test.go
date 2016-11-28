package featureflags_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testnet "code.cloudfoundry.org/cli/util/testhelpers/net"

	. "code.cloudfoundry.org/cli/cf/api/featureflags"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature Flags Repository", func() {
	var (
		testServer  *httptest.Server
		testHandler *testnet.TestHandler
		configRepo  coreconfig.ReadWriter
		repo        CloudControllerFeatureFlagRepository
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
		repo = NewCloudControllerFeatureFlagRepository(configRepo, gateway)
	})

	AfterEach(func() {
		testServer.Close()
	})

	setupTestServer := func(reqs ...testnet.TestRequest) {
		testServer, testHandler = testnet.NewServer(reqs)
		configRepo.SetAPIEndpoint(testServer.URL)
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

	Describe(".FindByName", func() {
		BeforeEach(func() {
			setupTestServer(featureFlagRequest)
		})

		It("returns the requested", func() {
			featureFlagModel, err := repo.FindByName("user_org_creation")

			Expect(err).NotTo(HaveOccurred())
			Expect(testHandler).To(HaveAllRequestsCalled())

			Expect(featureFlagModel.Name).To(Equal("user_org_creation"))
			Expect(featureFlagModel.Enabled).To(BeFalse())
		})
	})

	Describe(".Update", func() {
		BeforeEach(func() {
			setupTestServer(featureFlagsUpdateRequest)
		})

		It("updates the given feature flag with the specified value", func() {
			err := repo.Update("app_scaling", true)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when given a non-existent feature flag", func() {
			BeforeEach(func() {
				setupTestServer(featureFlagsUpdateErrorRequest)
			})

			It("returns an error", func() {
				err := repo.Update("i_dont_exist", true)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

var featureFlagsGetAllRequest = apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
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

var featureFlagRequest = apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/config/feature_flags/user_org_creation",
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `{
  "name": "user_org_creation",
  "enabled": false,
  "error_message": null,
  "url": "/v2/config/feature_flags/user_org_creation"
}`,
	},
})

var featureFlagsUpdateErrorRequest = apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "PUT",
	Path:   "/v2/config/feature_flags/i_dont_exist",
	Response: testnet.TestResponse{
		Status: http.StatusNotFound,
		Body: `{
         "code": 330000,
         "description": "The feature flag could not be found: i_dont_exist",
         "error_code": "CF-FeatureFlagNotFound"
         }`,
	},
})

var featureFlagsUpdateRequest = apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "PUT",
	Path:   "/v2/config/feature_flags/app_scaling",
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `{
      "name": "app_scaling",
      "enabled": true,
      "error_message": null,
      "url": "/v2/config/feature_flags/app_scaling"
    }`,
	},
})
