package api_test

import (
	. "cf/api"
	"cf/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
)

var _ = Describe("CloudControllerPasswordRepository", func() {
	It("TestUpdatePassword", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "PUT",
			Path:     "/Users/my-user-guid/password",
			Matcher:  testnet.RequestBodyMatcher(`{"password":"new-password","oldPassword":"old-password"}`),
			Response: testnet.TestResponse{Status: http.StatusOK},
		})

		passwordUpdateServer, handler, repo := createPasswordRepo(req)
		defer passwordUpdateServer.Close()

		apiResponse := repo.UpdatePassword("old-password", "new-password")
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiResponse).NotTo(HaveOccurred())
	})
})

func createPasswordRepo(req testnet.TestRequest) (passwordServer *httptest.Server, handler *testnet.TestHandler, repo PasswordRepository) {
	passwordServer, handler = testnet.NewServer([]testnet.TestRequest{req})

	endpointRepo := &testapi.FakeEndpointRepo{}
	endpointRepo.UAAEndpointReturns.Endpoint = passwordServer.URL
	configRepo := testconfig.NewRepositoryWithDefaults()
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerPasswordRepository(configRepo, gateway, endpointRepo)
	return
}
