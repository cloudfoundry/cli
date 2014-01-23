package api_test

import (
	. "cf/api"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
	"testing"
)

func TestUpdatePassword(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     "/Users/my-user-guid/password",
		Matcher:  testnet.RequestBodyMatcher(`{"password":"new-password","oldPassword":"old-password"}`),
		Response: testnet.TestResponse{Status: http.StatusOK},
	})

	accessToken, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{UserGuid: "my-user-guid"})
	assert.NoError(t, err)

	passwordUpdateServer, handler, repo := createPasswordRepo(t, req, accessToken)
	defer passwordUpdateServer.Close()

	apiResponse := repo.UpdatePassword("old-password", "new-password")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func createPasswordRepo(t *testing.T, req testnet.TestRequest, accessToken string) (passwordServer *httptest.Server, handler *testnet.TestHandler, repo PasswordRepository) {
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
