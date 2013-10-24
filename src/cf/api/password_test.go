package api

import (
	"cf"
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

func TestGetScore(t *testing.T) {
	testScore(t, `{"score":5,"requiredScore":5}`, "good")
	testScore(t, `{"score":10,"requiredScore":5}`, "strong")
	testScore(t, `{"score":4,"requiredScore":5}`, "weak")
}

func testScore(t *testing.T, scoreBody string, expectedScore string) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "POST",
		Path:     "/password/score",
		Matcher:  testnet.RequestBodyMatcherWithContentType("password=new-password", "application/x-www-form-urlencoded"),
		Response: testnet.TestResponse{Status: http.StatusOK, Body: scoreBody},
	})

	accessToken := "BEARER my_access_token"
	scoreServer, handler, repo := createPasswordRepo(t, req, accessToken)
	defer scoreServer.Close()

	score, apiResponse := repo.GetScore("new-password")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, score, expectedScore)
}

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

	endpointRepo := &testapi.FakeEndpointRepo{GetEndpointEndpoints: map[cf.EndpointType]string{
		cf.UaaEndpointKey: passwordServer.URL,
	}}

	config := &configuration.Configuration{
		AccessToken: accessToken,
	}
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerPasswordRepository(config, gateway, endpointRepo)
	return
}
