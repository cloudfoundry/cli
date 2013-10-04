package api_test

import (
	. "cf/api"
	"cf/net"
	"encoding/base64"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testhelpers"
	"testing"
)

var successfulLoginEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	contentTypeMatches := request.Header.Get("content-type") == "application/x-www-form-urlencoded"
	acceptHeaderMatches := request.Header.Get("accept") == "application/json"
	methodMatches := request.Method == "POST"
	pathMatches := request.URL.Path == "/oauth/token"
	encodedAuth := base64.StdEncoding.EncodeToString([]byte("cf:"))
	basicAuthMatches := request.Header.Get("authorization") == "Basic "+encodedAuth

	err := request.ParseForm()

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	usernameMatches := request.Form.Get("username") == "foo@example.com"
	passwordMatches := request.Form.Get("password") == "bar"
	grantTypeMatches := request.Form.Get("grant_type") == "password"
	scopeMatches := request.Form.Get("scope") == ""
	bodyMatches := usernameMatches && passwordMatches && grantTypeMatches && scopeMatches

	if !(contentTypeMatches && acceptHeaderMatches && methodMatches && pathMatches && bodyMatches && basicAuthMatches) {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonResponse := `
{
  "access_token": "my_access_token",
  "token_type": "BEARER",
  "refresh_token": "my_refresh_token",
  "scope": "openid",
  "expires_in": 98765
} `
	fmt.Fprintln(writer, jsonResponse)
}

func TestSuccessfullyLoggingIn(t *testing.T) {
	ts, auth := setupAuthWithEndpoint(t, successfulLoginEndpoint)
	defer ts.Close()

	apiStatus := auth.Authenticate("foo@example.com", "bar")
	savedConfig := testhelpers.SavedConfiguration

	assert.False(t, apiStatus.IsError())
	assert.Equal(t, savedConfig.AuthorizationEndpoint, ts.URL)
	assert.Equal(t, savedConfig.AccessToken, "BEARER my_access_token")
	assert.Equal(t, savedConfig.RefreshToken, "my_refresh_token")
}

var unsuccessfulLoginEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusUnauthorized)
}

func TestUnsuccessfullyLoggingIn(t *testing.T) {
	ts, auth := setupAuthWithEndpoint(t, unsuccessfulLoginEndpoint)
	defer ts.Close()

	apiStatus := auth.Authenticate("foo@example.com", "oops wrong pass")
	savedConfig := testhelpers.SavedConfiguration

	assert.True(t, apiStatus.NotSuccessful())
	assert.Equal(t, apiStatus.Message, "Password is incorrect, please try again.")
	assert.Empty(t, savedConfig.AccessToken)
}

var errorLoginEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusInternalServerError)
}

func TestServerErrorLoggingIn(t *testing.T) {
	ts, auth := setupAuthWithEndpoint(t, errorLoginEndpoint)
	defer ts.Close()

	apiStatus := auth.Authenticate("foo@example.com", "bar")
	savedConfig := testhelpers.SavedConfiguration

	assert.True(t, apiStatus.IsError())
	assert.Equal(t, apiStatus.Message, "Server error, status code: 500, error code: , message: ")
	assert.Empty(t, savedConfig.AccessToken)
}

var errorMaskedAsSuccessEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	jsonResponse := `
{"error":{"error":"rest_client_error","error_description":"I/O error: uaa.10.244.0.22.xip.io; nested exception is java.net.UnknownHostException: uaa.10.244.0.22.xip.io"}}
`

	writer.WriteHeader(http.StatusOK)
	fmt.Fprintln(writer, jsonResponse)
}

func TestLoggingInWithErrorMaskedAsSuccess(t *testing.T) {
	ts, auth := setupAuthWithEndpoint(t, errorMaskedAsSuccessEndpoint)
	defer ts.Close()

	apiStatus := auth.Authenticate("foo@example.com", "bar")
	savedConfig := testhelpers.SavedConfiguration

	assert.True(t, apiStatus.IsError())
	assert.Equal(t, apiStatus.Message, "Authentication Server error: I/O error: uaa.10.244.0.22.xip.io; nested exception is java.net.UnknownHostException: uaa.10.244.0.22.xip.io")
	assert.Empty(t, savedConfig.AccessToken)
}

func setupAuthWithEndpoint(t *testing.T, handler func(http.ResponseWriter, *http.Request)) (ts *httptest.Server, auth UAAAuthenticationRepository) {
	ts = httptest.NewTLSServer(http.HandlerFunc(handler))

	configRepo := testhelpers.FakeConfigRepository{}
	configRepo.Delete()
	config, err := configRepo.Get()
	assert.NoError(t, err)
	config.AuthorizationEndpoint = ts.URL
	config.AccessToken = ""

	gateway := net.NewUAAAuthGateway()

	auth = NewUAAAuthenticationRepository(gateway, configRepo)
	return
}
