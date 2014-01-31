package api_test

import (
	"cf"
	. "cf/api"
	"cf/configuration"
	"cf/net"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testnet "testhelpers/net"
)

var defaultCreateRequestBodyMatcher = testnet.RequestBodyMatcher(`{"app_guid":"my-app-guid","service_instance_guid":"my-service-instance-guid","async":true}`)

var deleteBindingReq = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method:   "DELETE",
	Path:     "/v2/service_bindings/service-binding-2-guid",
	Response: testnet.TestResponse{Status: http.StatusOK},
})

func createServiceBindingRepo(t mr.TestingT, requests []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo ServiceBindingRepository) {
	ts, handler = testnet.NewTLSServer(t, requests)
	space := cf.SpaceFields{}
	space.Guid = "my-space-guid"
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		SpaceFields: space,
		Target:      ts.URL,
	}

	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerServiceBindingRepository(config, gateway)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestCreateServiceBinding", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "POST",
				Path:     "/v2/service_bindings",
				Matcher:  defaultCreateRequestBodyMatcher,
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			ts, handler, repo := createServiceBindingRepo(mr.T(), []testnet.TestRequest{req})
			defer ts.Close()

			apiResponse := repo.Create("my-service-instance-guid", "my-app-guid")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
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

			ts, handler, repo := createServiceBindingRepo(mr.T(), []testnet.TestRequest{req})
			defer ts.Close()

			apiResponse := repo.Create("my-service-instance-guid", "my-app-guid")

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsNotSuccessful())
			assert.Equal(mr.T(), apiResponse.ErrorCode, "90003")
		})
		It("TestDeleteServiceBinding", func() {

			ts, handler, repo := createServiceBindingRepo(mr.T(), []testnet.TestRequest{deleteBindingReq})
			defer ts.Close()

			serviceInstance := cf.ServiceInstance{}
			serviceInstance.Guid = "my-service-instance-guid"

			binding := cf.ServiceBindingFields{}
			binding.Url = "/v2/service_bindings/service-binding-1-guid"
			binding.AppGuid = "app-1-guid"
			binding2 := cf.ServiceBindingFields{}
			binding2.Url = "/v2/service_bindings/service-binding-2-guid"
			binding2.AppGuid = "app-2-guid"
			serviceInstance.ServiceBindings = []cf.ServiceBindingFields{binding, binding2}

			found, apiResponse := repo.Delete(serviceInstance, "app-2-guid")

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
			assert.True(mr.T(), found)
		})
		It("TestDeleteServiceBindingWhenBindingDoesNotExist", func() {

			ts, handler, repo := createServiceBindingRepo(mr.T(), []testnet.TestRequest{})
			defer ts.Close()

			serviceInstance := cf.ServiceInstance{}
			serviceInstance.Guid = "my-service-instance-guid"

			found, apiResponse := repo.Delete(serviceInstance, "app-2-guid")

			assert.Equal(mr.T(), handler.CallCount, 0)
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
			assert.False(mr.T(), found)
		})
	})
}
