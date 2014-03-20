package api_test

import (
	. "cf/api"
	"cf/errors"
	"cf/models"
	"cf/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
)

var _ = Describe("Testing with ginkgo", func() {
	var defaultCreateRequestBodyMatcher testnet.RequestMatcher
	var deleteBindingReq testnet.TestRequest

	BeforeEach(func() {
		defaultCreateRequestBodyMatcher = testnet.RequestBodyMatcher(`{"app_guid":"my-app-guid","service_instance_guid":"my-service-instance-guid","async":true}`)
		deleteBindingReq = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "DELETE",
			Path:     "/v2/service_bindings/service-binding-2-guid",
			Response: testnet.TestResponse{Status: http.StatusOK},
		})
	})

	It("TestCreateServiceBinding", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "POST",
			Path:     "/v2/service_bindings",
			Matcher:  defaultCreateRequestBodyMatcher,
			Response: testnet.TestResponse{Status: http.StatusCreated},
		})

		ts, handler, repo := createServiceBindingRepo([]testnet.TestRequest{req})
		defer ts.Close()

		apiErr := repo.Create("my-service-instance-guid", "my-app-guid")
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})

	It("TestCreateServiceBindingIfError", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:  "POST",
			Path:    "/v2/service_bindings",
			Matcher: defaultCreateRequestBodyMatcher,
			Response: testnet.TestResponse{
				Status: http.StatusBadRequest,
				Body:   `{"code":90003,"description":"The app space binding to service is taken: 7b959018-110a-4913-ac0a-d663e613cdea 346bf237-7eef-41a7-b892-68fb08068f09"}`,
			},
		})

		ts, handler, repo := createServiceBindingRepo([]testnet.TestRequest{req})
		defer ts.Close()

		apiErr := repo.Create("my-service-instance-guid", "my-app-guid")

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(BeNil())
		Expect(apiErr.(errors.HttpError).ErrorCode()).To(Equal("90003"))
	})

	It("TestDeleteServiceBinding", func() {
		ts, handler, repo := createServiceBindingRepo([]testnet.TestRequest{deleteBindingReq})
		defer ts.Close()

		serviceInstance := models.ServiceInstance{}
		serviceInstance.Guid = "my-service-instance-guid"

		binding := models.ServiceBindingFields{}
		binding.Url = "/v2/service_bindings/service-binding-1-guid"
		binding.AppGuid = "app-1-guid"
		binding2 := models.ServiceBindingFields{}
		binding2.Url = "/v2/service_bindings/service-binding-2-guid"
		binding2.AppGuid = "app-2-guid"
		serviceInstance.ServiceBindings = []models.ServiceBindingFields{binding, binding2}

		found, apiErr := repo.Delete(serviceInstance, "app-2-guid")

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
	})

	It("TestDeleteServiceBindingWhenBindingDoesNotExist", func() {
		ts, handler, repo := createServiceBindingRepo([]testnet.TestRequest{})
		defer ts.Close()

		serviceInstance := models.ServiceInstance{}
		serviceInstance.Guid = "my-service-instance-guid"

		found, apiErr := repo.Delete(serviceInstance, "app-2-guid")

		Expect(handler.CallCount).To(Equal(0))
		Expect(apiErr).NotTo(HaveOccurred())
		Expect(found).To(BeFalse())
	})
})

func createServiceBindingRepo(requests []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo ServiceBindingRepository) {
	ts, handler = testnet.NewServer(requests)
	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway(configRepo)
	repo = NewCloudControllerServiceBindingRepository(configRepo, gateway)
	return
}
