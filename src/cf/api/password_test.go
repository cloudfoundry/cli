package api_test

import (
	. "cf/api"
	"cf/configuration"
	"cf/net"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
)

func createPasswordRepo(t mr.TestingT, req testnet.TestRequest, accessToken string) (passwordServer *httptest.Server, handler *testnet.TestHandler, repo PasswordRepository) {
	passwordServer, handler = testnet.NewTLSServer(t, []testnet.TestRequest{req})

	endpointRepo := &testapi.FakeEndpointRepo{}
	endpointRepo.UAAEndpointReturns.Endpoint = passwordServer.URL

	config := &configuration.Configuration{
		AccessToken: accessToken,
	}
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerPasswordRepository(config, gateway, endpointRepo)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestUpdatePassword", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/Users/my-user-guid/password",
				Matcher:  testnet.RequestBodyMatcher(`{"password":"new-password","oldPassword":"old-password"}`),
				Response: testnet.TestResponse{Status: http.StatusOK},
			})

			accessToken, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{UserGuid: "my-user-guid"})
			assert.NoError(mr.T(), err)

			passwordUpdateServer, handler, repo := createPasswordRepo(mr.T(), req, accessToken)
			defer passwordUpdateServer.Close()

			apiResponse := repo.UpdatePassword("old-password", "new-password")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
		})
	})
}
