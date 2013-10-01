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
	ts := httptest.NewTLSServer(http.HandlerFunc(successfulLoginEndpoint))
	defer ts.Close()

	configRepo := testhelpers.FakeConfigRepository{}
	configRepo.Delete()
	config, err := configRepo.Get()
	assert.NoError(t, err)
	config.AuthorizationEndpoint = ts.URL
	config.AccessToken = ""
	configRepo.Save(config)
	newConfig, err := configRepo.Get()
	assert.Equal(t, newConfig.AuthorizationEndpoint, ts.URL)

	gateway := net.NewUAAAuthGateway()
	auth := NewUAAAuthenticator(gateway, configRepo)
	apiStatus := auth.Authenticate("foo@example.com", "bar")
	assert.False(t, apiStatus.IsError())
	savedConfig, err := configRepo.Get()
	assert.NoError(t, err)
	assert.Equal(t, savedConfig.AuthorizationEndpoint, ts.URL)
	assert.Equal(t, savedConfig.AccessToken, "BEARER my_access_token")
	assert.Equal(t, savedConfig.RefreshToken, "my_refresh_token")
}

var unsuccessfulLoginEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusUnauthorized)
}

func TestUnsuccessfullyLoggingIn(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(unsuccessfulLoginEndpoint))
	defer ts.Close()

	configRepo := testhelpers.FakeConfigRepository{}
	configRepo.Delete()
	config, err := configRepo.Get()
	assert.NoError(t, err)
	config.AuthorizationEndpoint = ts.URL
	config.AccessToken = ""
	gateway := net.NewUAAAuthGateway()

	auth := NewUAAAuthenticator(gateway, configRepo)
	apiStatus := auth.Authenticate("foo@example.com", "oops wrong pass")
	assert.True(t, apiStatus.IsError())
	assert.Equal(t, apiStatus.Message, "Password is incorrect, please try again.")
	savedConfig, _ := configRepo.Get()
	assert.Empty(t, savedConfig.AccessToken)
}

var errorLoginEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusInternalServerError)
}

func TestServerErrorLoggingIn(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(errorLoginEndpoint))
	defer ts.Close()

	configRepo := testhelpers.FakeConfigRepository{}
	configRepo.Delete()
	config, err := configRepo.Get()
	assert.NoError(t, err)
	config.AuthorizationEndpoint = ts.URL
	config.AccessToken = ""
	configRepo.Save(config)

	gateway := net.NewUAAAuthGateway()
	auth := NewUAAAuthenticator(gateway, configRepo)
	apiStatus := auth.Authenticate("foo@example.com", "bar")
	assert.True(t, apiStatus.IsError())
	assert.Equal(t, apiStatus.Message, "Server error, status code: 500, error code: , message: ")
	savedConfig, _ := configRepo.Get()
	assert.Empty(t, savedConfig.AccessToken)
}
