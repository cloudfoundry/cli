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

	passwordScoreEndpoint := testhelpers.CreateEndpoint(
		"POST",
		"/password/score",
		func(req *http.Request) bool {
			bodyMatcher := testhelpers.RequestBodyMatcher("password=new-password")
			contentTypeMatches := req.Header.Get("Content-Type") == "application/x-www-form-urlencoded"

			return contentTypeMatches && bodyMatcher(req)
		},
		passwordScoreResponse,
	)

	scoreServer := httptest.NewTLSServer(http.HandlerFunc(passwordScoreEndpoint))
	defer scoreServer.Close()

	targetServer := createInfoServer(scoreServer.URL)
	defer targetServer.Close()

	config := configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      targetServer.URL,
	}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerPasswordRepository(config, gateway)

	score, apiStatus := repo.GetScore("new-password")
	assert.False(t, apiStatus.IsError())
	assert.Equal(t, score, expectedScore)
}

func TestUpdatePassword(t *testing.T) {
	var passwordWasUpdated bool

	passwordUpdateResponse := testhelpers.TestResponse{Status: http.StatusOK}

	passwordUpdateEndpoint := testhelpers.CreateEndpoint(
		"PUT",
		"/Users/my-user-guid/password",
		func(req *http.Request) bool {
			passwordWasUpdated = true

			bodyMatcher := testhelpers.RequestBodyMatcher(`{"password":"new-password","oldPassword":"old-password"}`)
			contentTypeMatches := req.Header.Get("Content-Type") == "application/json"

			return contentTypeMatches && bodyMatcher(req)
		},
		passwordUpdateResponse,
	)

	passwordUpdateServer := httptest.NewTLSServer(http.HandlerFunc(passwordUpdateEndpoint))
	defer passwordUpdateServer.Close()

	targetServer := createInfoServer(passwordUpdateServer.URL)
	defer targetServer.Close()

	tokenInfo := `{"user_id":"my-user-guid"}`
	encodedTokenInfo := base64.StdEncoding.EncodeToString([]byte(tokenInfo))

	config := configuration.Configuration{
		AccessToken: fmt.Sprintf("BEARER my_access_token.%s.baz", encodedTokenInfo),
		Target:      targetServer.URL,
	}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerPasswordRepository(config, gateway)

	apiStatus := repo.UpdatePassword("old-password", "new-password")
	assert.False(t, apiStatus.IsError())
	assert.True(t, passwordWasUpdated)
}

func createInfoServer(tokenEndpoint string) *httptest.Server {
	targetInfoResponse := testhelpers.TestResponse{
		Status: http.StatusOK,
		Body:   fmt.Sprintf(`{"token_endpoint": "%s"}`, tokenEndpoint),
	}
	targetInfoEndpoint := testhelpers.CreateEndpoint("GET", "/info", nil, targetInfoResponse)

	return httptest.NewTLSServer(http.HandlerFunc(targetInfoEndpoint))
}
