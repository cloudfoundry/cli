package api_test

import (
	. "cf/api"
	"cf/configuration"
	"cf/net"
	"encoding/base64"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mr "github.com/tjarratt/mr_t"
	"net/http"
	"net/http/httptest"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
)

var _ = Describe("AuthenticationRepository", func() {
	It("TestSuccessfullyLoggingIn", func() {
		deps := setupAuthDependencies(successfulLoginRequest)
		defer teardownAuthDependencies(deps)

		auth := NewUAAAuthenticationRepository(deps.gateway, deps.config)
		apiResponse := auth.Authenticate("foo@example.com", "bar")

		Expect(deps.handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsError()).To(BeFalse())
		Expect(deps.config.AuthorizationEndpoint()).To(Equal(deps.ts.URL))
		Expect(deps.config.AccessToken()).To(Equal("BEARER my_access_token"))
		Expect(deps.config.RefreshToken()).To(Equal("my_refresh_token"))
	})

	It("TestUnsuccessfullyLoggingIn", func() {
		deps := setupAuthDependencies(unsuccessfulLoginRequest)
		defer teardownAuthDependencies(deps)

		auth := NewUAAAuthenticationRepository(deps.gateway, deps.config)
		apiResponse := auth.Authenticate("foo@example.com", "oops wrong pass")

		Expect(deps.handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsNotSuccessful()).To(BeTrue())
		Expect(apiResponse.Message).To(Equal("Password is incorrect, please try again."))
		Expect(deps.config.AccessToken()).To(BeEmpty())
	})

	It("TestServerErrorLoggingIn", func() {
		deps := setupAuthDependencies(errorLoginRequest)
		defer teardownAuthDependencies(deps)

		auth := NewUAAAuthenticationRepository(deps.gateway, deps.config)
		apiResponse := auth.Authenticate("foo@example.com", "bar")

		Expect(deps.handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsError()).To(BeTrue())
		Expect(apiResponse.Message).To(Equal("Server error, status code: 500, error code: , message: "))
		Expect(deps.config.AccessToken()).To(BeEmpty())
	})

	It("TestLoggingInWithErrorMaskedAsSuccess", func() {
		deps := setupAuthDependencies(errorMaskedAsSuccessLoginRequest)
		defer teardownAuthDependencies(deps)

		auth := NewUAAAuthenticationRepository(deps.gateway, deps.config)
		apiResponse := auth.Authenticate("foo@example.com", "bar")

		Expect(deps.handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsError()).To(BeTrue())
		Expect(apiResponse.Message).To(Equal("Authentication Server error: I/O error: uaa.10.244.0.22.xip.io; nested exception is java.net.UnknownHostException: uaa.10.244.0.22.xip.io"))
		Expect(deps.config.AccessToken()).To(BeEmpty())
	})
})

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
		Fail(fmt.Sprintf("Failed to parse form: %s", err))
		return
	}

	Expect(request.Form.Get("username")).To(Equal("foo@example.com"), "Username did not match.")
	Expect(request.Form.Get("password")).To(Equal("bar"), "Password did not match.")
	Expect(request.Form.Get("grant_type")).To(Equal("password"), "Grant type did not match.")
	Expect(request.Form.Get("scope")).To(Equal(""), "Scope did not mathc.")
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

type authDependencies struct {
	ts      *httptest.Server
	handler *testnet.TestHandler
	config  configuration.ReadWriter
	gateway net.Gateway
}

func setupAuthDependencies(request testnet.TestRequest) (deps authDependencies) {
	deps.ts, deps.handler = testnet.NewTLSServer(GinkgoT(), []testnet.TestRequest{request})

	deps.config = testconfig.NewRepository()
	deps.config.SetAuthorizationEndpoint(deps.ts.URL)

	deps.gateway = net.NewUAAGateway()
	return
}

func teardownAuthDependencies(deps authDependencies) {
	deps.ts.Close()
}
