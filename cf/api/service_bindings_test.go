package api_test

import (
	. "github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("Testing with ginkgo", func() {
	var (
		testServer  *httptest.Server
		testHandler *testnet.TestHandler
		configRepo  configuration.ReadWriter
		repo        CloudControllerServiceBindingRepository
	)

	setupTestServer := func(reqs ...testnet.TestRequest) {
		testServer, testHandler = testnet.NewServer(reqs)
		configRepo.SetApiEndpoint(testServer.URL)
	}

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()

		gateway := net.NewCloudControllerGateway(configRepo)
		repo = NewCloudControllerServiceBindingRepository(configRepo, gateway)
	})

	AfterEach(func() {
		testServer.Close()
	})

	Describe("Create", func() {
		Context("when the service binding can be created", func() {
			BeforeEach(func() {
				setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "POST",
					Path:     "/v2/service_bindings",
					Matcher:  testnet.RequestBodyMatcher(`{"app_guid":"my-app-guid","service_instance_guid":"my-service-instance-guid","async":true}`),
					Response: testnet.TestResponse{Status: http.StatusCreated},
				}))
			})

			It("TestCreateServiceBinding", func() {
				apiErr := repo.Create("my-service-instance-guid", "my-app-guid")

				Expect(testHandler).To(testnet.HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
			})
		})

		Context("when an error occurs", func() {
			BeforeEach(func() {
				setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:  "POST",
					Path:    "/v2/service_bindings",
					Matcher: testnet.RequestBodyMatcher(`{"app_guid":"my-app-guid","service_instance_guid":"my-service-instance-guid","async":true}`),
					Response: testnet.TestResponse{
						Status: http.StatusBadRequest,
						Body:   `{"code":90003,"description":"The app space binding to service is taken: 7b959018-110a-4913-ac0a-d663e613cdea 346bf237-7eef-41a7-b892-68fb08068f09"}`,
					},
				}))
			})

			It("TestCreateServiceBindingIfError", func() {
				apiErr := repo.Create("my-service-instance-guid", "my-app-guid")

				Expect(testHandler).To(testnet.HaveAllRequestsCalled())
				Expect(apiErr).NotTo(BeNil())
				Expect(apiErr.(errors.HttpError).ErrorCode()).To(Equal("90003"))
			})
		})
	})

	Describe("Delete", func() {
		Context("when binding does exist", func() {
			var serviceInstance models.ServiceInstance

			BeforeEach(func() {
				setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "DELETE",
					Path:     "/v2/service_bindings/service-binding-2-guid",
					Response: testnet.TestResponse{Status: http.StatusOK},
				}))

				serviceInstance.Guid = "my-service-instance-guid"

				binding := models.ServiceBindingFields{}
				binding.Url = "/v2/service_bindings/service-binding-1-guid"
				binding.AppGuid = "app-1-guid"
				binding2 := models.ServiceBindingFields{}
				binding2.Url = "/v2/service_bindings/service-binding-2-guid"
				binding2.AppGuid = "app-2-guid"
				serviceInstance.ServiceBindings = []models.ServiceBindingFields{binding, binding2}
			})

			It("TestDeleteServiceBinding", func() {
				found, apiErr := repo.Delete(serviceInstance, "app-2-guid")

				Expect(testHandler).To(testnet.HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
			})
		})

		Context("when binding does not exist", func() {
			var serviceInstance models.ServiceInstance

			BeforeEach(func() {
				setupTestServer()
				serviceInstance.Guid = "my-service-instance-guid"
			})

			It("does not return an error", func() {
				found, apiErr := repo.Delete(serviceInstance, "app-2-guid")

				Expect(testHandler.CallCount).To(Equal(0))
				Expect(apiErr).NotTo(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})
	})
})
