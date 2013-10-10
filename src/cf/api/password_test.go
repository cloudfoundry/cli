package api_test

import (
	. "cf/api"
	"cf/configuration"
	"cf/net"
	"encoding/base64"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testhelpers"
	"testing"
)

func TestGetScore(t *testing.T) {
	testScore(t, `{"score":5,"requiredScore":5}`, "good")
	testScore(t, `{"score":10,"requiredScore":5}`, "strong")
	testScore(t, `{"score":4,"requiredScore":5}`, "weak")
}

func testScore(t *testing.T, scoreBody string, expectedScore string) {
	passwordScoreResponse := testhelpers.TestResponse{Status: http.StatusOK, Body: scoreBody}

	endpoint, status := testhelpers.CreateCheckableEndpoint(
		"POST",
		"/password/score",
		func(req *http.Request) bool {
			bodyMatcher := testhelpers.RequestBodyMatcher("password=new-password")
			contentTypeMatches := req.Header.Get("Content-Type") == "application/x-www-form-urlencoded"

			return contentTypeMatches && bodyMatcher(req)
		},
		passwordScoreResponse,
	)

	scoreServer := httptest.NewTLSServer(endpoint)
	defer scoreServer.Close()

	targetServer, targetEndpointStatus := createInfoServer(scoreServer.URL)
	defer targetServer.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      targetServer.URL,
	}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerPasswordRepository(config, gateway)

	score, apiResponse := repo.GetScore("new-password")
	assert.True(t, targetEndpointStatus.Called())
	assert.True(t, status.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, score, expectedScore)
}

func TestUpdatePassword(t *testing.T) {
	passwordUpdateResponse := testhelpers.TestResponse{Status: http.StatusOK}

	passwordUpdateEndpoint, passwordUpdateEndpointStatus := testhelpers.CreateCheckableEndpoint(
		"PUT",
		"/Users/my-user-guid/password",
		func(req *http.Request) bool {
			bodyMatcher := testhelpers.RequestBodyMatcher(`{"password":"new-password","oldPassword":"old-password"}`)
			contentTypeMatches := req.Header.Get("Content-Type") == "application/json"

			return contentTypeMatches && bodyMatcher(req)
		},
		passwordUpdateResponse,
	)

	passwordUpdateServer := httptest.NewTLSServer(passwordUpdateEndpoint)
	defer passwordUpdateServer.Close()

	targetServer, targetEndpointStatus := createInfoServer(passwordUpdateServer.URL)
	defer targetServer.Close()

	tokenInfo := `{"user_id":"my-user-guid"}`
	encodedTokenInfo := base64.StdEncoding.EncodeToString([]byte(tokenInfo))

	config := &configuration.Configuration{
		AccessToken: fmt.Sprintf("BEARER my_access_token.%s.baz", encodedTokenInfo),
		Target:      targetServer.URL,
	}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerPasswordRepository(config, gateway)

	apiResponse := repo.UpdatePassword("old-password", "new-password")
	assert.True(t, targetEndpointStatus.Called())
	assert.True(t, passwordUpdateEndpointStatus.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func createInfoServer(tokenEndpoint string) (ts *httptest.Server, status *testhelpers.RequestStatus) {
	endpoint, status := testhelpers.CreateCheckableEndpoint(
		"GET",
		"/info",
		nil,
		testhelpers.TestResponse{
			Status: http.StatusOK,
			Body:   fmt.Sprintf(`{"token_endpoint": "%s"}`, tokenEndpoint),
		},
	)

	ts = httptest.NewTLSServer(endpoint)
	return
}
