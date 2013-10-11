package api

import (
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
	targetServer, scoreServer, repo := createPasswordRepo(endpoint, accessToken)
	defer scoreServer.Close()
	defer targetServer.Close()

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

	targetServer, passwordUpdateServer, repo := createPasswordRepo(passwordUpdateEndpoint, accessToken)
	defer passwordUpdateServer.Close()
	defer targetServer.Close()

	apiResponse := repo.UpdatePassword("old-password", "new-password")
	assert.True(t, passwordUpdateEndpointStatus.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func createPasswordRepo(passwordEndpoint http.HandlerFunc, accessToken string) (targetServer *httptest.Server, passwordServer *httptest.Server, repo PasswordRepository) {
	passwordServer = httptest.NewTLSServer(passwordEndpoint)
	targetServer, _ = createInfoServer(passwordServer.URL)

	config := &configuration.Configuration{
		AccessToken: accessToken,
		Target:      targetServer.URL,
	}
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerPasswordRepository(config, gateway)
	return
}

func createInfoServer(tokenEndpoint string) (ts *httptest.Server, status *testapi.RequestStatus) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"GET",
		"/info",
		nil,
		testapi.TestResponse{
			Status: http.StatusOK,
			Body:   fmt.Sprintf(`{"token_endpoint": "%s"}`, tokenEndpoint),
		},
	)

	ts = httptest.NewTLSServer(endpoint)
	return
}
