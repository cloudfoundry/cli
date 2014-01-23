package api_test

import (
	. "cf/api"
	"cf/net"
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
	"testing"
)

var authHeaders = http.Header{
	"accept":        {"application/json"},
	"content-type":  {"application/x-www-form-urlencoded"},
	"authorization": {"Basic " + base64.StdEncoding.EncodeToString([]byte("cf:"))},
}

var successfulLoginRequest = testnet.TestRequest{
	Method:  "POST",
	Path:    "/oauth/token",
	Header:  authHeaders,
	Matcher: successfulLoginMatcher,
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `
{
  "access_token": "my_access_token",
  "token_type": "BEARER",
  "refresh_token": "my_refresh_token",
  "scope": "openid",
  "expires_in": 98765
} `},
}

var successfulLoginMatcher = func(t *testing.T, request *http.Request) {
	err := request.ParseForm()
	if err != nil {
		assert.Fail(t, "Failed to parse form: %s", err)
		return
	}

	assert.Equal(t, request.Form.Get("username"), "foo@example.com", "Username did not match.")
	assert.Equal(t, request.Form.Get("password"), "bar", "Password did not match.")
	assert.Equal(t, request.Form.Get("grant_type"), "password", "Grant type did not match.")
	assert.Equal(t, request.Form.Get("scope"), "", "Scope did not mathc.")
}

func TestSuccessfullyLoggingIn(t *testing.T) {
	ts, handler, auth := setupAuthWithEndpoint(t, successfulLoginRequest)
	defer ts.Close()

	apiResponse := auth.Authenticate("foo@example.com", "bar")
	savedConfig := testconfig.SavedConfiguration

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsError())
	assert.Equal(t, savedConfig.AuthorizationEndpoint, ts.URL)
	assert.Equal(t, savedConfig.AccessToken, "BEARER my_access_token")
	assert.Equal(t, savedConfig.RefreshToken, "my_refresh_token")
}

var unsuccessfulLoginRequest = testnet.TestRequest{
	Method: "POST",
	Path:   "/oauth/token",
	Response: testnet.TestResponse{
		Status: http.StatusUnauthorized,
	},
}

func TestUnsuccessfullyLoggingIn(t *testing.T) {
	ts, handler, auth := setupAuthWithEndpoint(t, unsuccessfulLoginRequest)
	defer ts.Close()

	apiResponse := auth.Authenticate("foo@example.com", "oops wrong pass")
	savedConfig := testconfig.SavedConfiguration

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, apiResponse.Message, "Password is incorrect, please try again.")
	assert.Empty(t, savedConfig.AccessToken)
}

var errorLoginRequest = testnet.TestRequest{
	Method: "POST",
	Path:   "/oauth/token",
	Response: testnet.TestResponse{
		Status: http.StatusInternalServerError,
	},
}

func TestServerErrorLoggingIn(t *testing.T) {
	ts, handler, auth := setupAuthWithEndpoint(t, errorLoginRequest)
	defer ts.Close()

	apiResponse := auth.Authenticate("foo@example.com", "bar")
	savedConfig := testconfig.SavedConfiguration

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsError())
	assert.Equal(t, apiResponse.Message, "Server error, status code: 500, error code: , message: ")
	assert.Empty(t, savedConfig.AccessToken)
}

var errorMaskedAsSuccessLoginRequest = testnet.TestRequest{
	Method: "POST",
	Path:   "/oauth/token",
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `
{"error":{"error":"rest_client_error","error_description":"I/O error: uaa.10.244.0.22.xip.io; nested exception is java.net.UnknownHostException: uaa.10.244.0.22.xip.io"}}
`},
}

func TestLoggingInWithErrorMaskedAsSuccess(t *testing.T) {
	ts, handler, auth := setupAuthWithEndpoint(t, errorMaskedAsSuccessLoginRequest)
	defer ts.Close()

	apiResponse := auth.Authenticate("foo@example.com", "bar")
	savedConfig := testconfig.SavedConfiguration

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsError())
	assert.Equal(t, apiResponse.Message, "Authentication Server error: I/O error: uaa.10.244.0.22.xip.io; nested exception is java.net.UnknownHostException: uaa.10.244.0.22.xip.io")
	assert.Empty(t, savedConfig.AccessToken)
}

func setupAuthWithEndpoint(t *testing.T, request testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, auth UAAAuthenticationRepository) {
	ts, handler = testnet.NewTLSServer(t, []testnet.TestRequest{request})

	configRepo := testconfig.FakeConfigRepository{}
	configRepo.Delete()
	config, err := configRepo.Get()
	assert.NoError(t, err)
	config.AuthorizationEndpoint = ts.URL
	config.AccessToken = ""

	gateway := net.NewUAAGateway()

	auth = NewUAAAuthenticationRepository(gateway, configRepo)
	return
}
