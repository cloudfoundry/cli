package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"encoding/base64"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	"testing"
)

func TestGetScore(t *testing.T) {
	testScore(t, `{"score":5,"requiredScore":5}`, "good")
	testScore(t, `{"score":10,"requiredScore":5}`, "strong")
	testScore(t, `{"score":4,"requiredScore":5}`, "weak")
}

func testScore(t *testing.T, scoreBody string, expectedScore string) {
	passwordScoreResponse := testapi.TestResponse{Status: http.StatusOK, Body: scoreBody}

	endpoint, status := testapi.CreateCheckableEndpoint(
		"POST",
		"/password/score",
		testapi.RequestBodyMatcherWithContentType("password=new-password", "application/x-www-form-urlencoded"),
		passwordScoreResponse,
	)

	accessToken := "BEARER my_access_token"
	scoreServer, repo := createPasswordRepo(endpoint, accessToken)
	defer scoreServer.Close()

	score, apiResponse := repo.GetScore("new-password")
	assert.True(t, status.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, score, expectedScore)
}

func TestUpdatePassword(t *testing.T) {
	passwordUpdateResponse := testapi.TestResponse{Status: http.StatusOK}

	passwordUpdateEndpoint, passwordUpdateEndpointStatus := testapi.CreateCheckableEndpoint(
		"PUT",
		"/Users/my-user-guid/password",
		testapi.RequestBodyMatcher(`{"password":"new-password","oldPassword":"old-password"}`),
		passwordUpdateResponse,
	)

	tokenInfo := `{"user_id":"my-user-guid"}`
	encodedTokenInfo := base64.StdEncoding.EncodeToString([]byte(tokenInfo))
	accessToken := fmt.Sprintf("BEARER my_access_token.%s.baz", encodedTokenInfo)

	passwordUpdateServer, repo := createPasswordRepo(passwordUpdateEndpoint, accessToken)
	defer passwordUpdateServer.Close()

	apiResponse := repo.UpdatePassword("old-password", "new-password")
	assert.True(t, passwordUpdateEndpointStatus.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func createPasswordRepo(passwordEndpoint http.HandlerFunc, accessToken string) (passwordServer *httptest.Server, repo PasswordRepository) {
	passwordServer = httptest.NewTLSServer(passwordEndpoint)
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
