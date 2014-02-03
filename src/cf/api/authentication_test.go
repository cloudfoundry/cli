package api_test

import (
	. "cf/api"
	"cf/net"
	"encoding/base64"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"net/http"
	"net/http/httptest"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
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

var successfulLoginMatcher = func(t mr.TestingT, request *http.Request) {
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

var unsuccessfulLoginRequest = testnet.TestRequest{
	Method: "POST",
	Path:   "/oauth/token",
	Response: testnet.TestResponse{
		Status: http.StatusUnauthorized,
	},
}

var errorLoginRequest = testnet.TestRequest{
	Method: "POST",
	Path:   "/oauth/token",
	Response: testnet.TestResponse{
		Status: http.StatusInternalServerError,
	},
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

func setupAuthWithEndpoint(t mr.TestingT, request testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, auth UAAAuthenticationRepository) {
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
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSuccessfullyLoggingIn", func() {
			ts, handler, auth := setupAuthWithEndpoint(mr.T(), successfulLoginRequest)
			defer ts.Close()

			apiResponse := auth.Authenticate("foo@example.com", "bar")
			savedConfig := testconfig.SavedConfiguration

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsError())
			assert.Equal(mr.T(), savedConfig.AuthorizationEndpoint, ts.URL)
			assert.Equal(mr.T(), savedConfig.AccessToken, "BEARER my_access_token")
			assert.Equal(mr.T(), savedConfig.RefreshToken, "my_refresh_token")
		})
		It("TestUnsuccessfullyLoggingIn", func() {

			ts, handler, auth := setupAuthWithEndpoint(mr.T(), unsuccessfulLoginRequest)
			defer ts.Close()

			apiResponse := auth.Authenticate("foo@example.com", "oops wrong pass")
			savedConfig := testconfig.SavedConfiguration

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsNotSuccessful())
			assert.Equal(mr.T(), apiResponse.Message, "Password is incorrect, please try again.")
			assert.Empty(mr.T(), savedConfig.AccessToken)
		})
		It("TestServerErrorLoggingIn", func() {

			ts, handler, auth := setupAuthWithEndpoint(mr.T(), errorLoginRequest)
			defer ts.Close()

			apiResponse := auth.Authenticate("foo@example.com", "bar")
			savedConfig := testconfig.SavedConfiguration

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsError())
			assert.Equal(mr.T(), apiResponse.Message, "Server error, status code: 500, error code: , message: ")
			assert.Empty(mr.T(), savedConfig.AccessToken)
		})
		It("TestLoggingInWithErrorMaskedAsSuccess", func() {

			ts, handler, auth := setupAuthWithEndpoint(mr.T(), errorMaskedAsSuccessLoginRequest)
			defer ts.Close()

			apiResponse := auth.Authenticate("foo@example.com", "bar")
			savedConfig := testconfig.SavedConfiguration

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsError())
			assert.Equal(mr.T(), apiResponse.Message, "Authentication Server error: I/O error: uaa.10.244.0.22.xip.io; nested exception is java.net.UnknownHostException: uaa.10.244.0.22.xip.io")
			assert.Empty(mr.T(), savedConfig.AccessToken)
		})
	})
}
