package api_test

import (
	. "cf/api"
	"cf/configuration"
	"cf/net"
	"encoding/base64"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
)

var _ = Describe("AuthenticationRepository", func() {
	var (
		gateway net.Gateway
		ts      *httptest.Server
		handler *testnet.TestHandler
		config  configuration.ReadWriter
	)

	BeforeEach(func() {
		config = testconfig.NewRepositoryWithDefaults()
		gateway = net.NewUAAGateway(config)
	})

	AfterEach(func() {
		ts.Close()
	})

	It("logs in", func() {
		ts, handler, config = setupAuthDependencies(successfulLoginRequest)

		auth := NewUAAAuthenticationRepository(gateway, config)
		apiErr := auth.Authenticate(map[string]string{
			"username": "foo@example.com",
			"password": "bar",
		})

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
		Expect(config.AuthenticationEndpoint()).To(Equal(ts.URL))
		Expect(config.AccessToken()).To(Equal("BEARER my_access_token"))
		Expect(config.RefreshToken()).To(Equal("my_refresh_token"))
	})

	It("returns a failure response when login fails", func() {
		ts, handler, config = setupAuthDependencies(unsuccessfulLoginRequest)

		auth := NewUAAAuthenticationRepository(gateway, config)
		apiErr := auth.Authenticate(map[string]string{
			"username": "foo@example.com",
			"password": "bar",
		})

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(BeNil())
		Expect(apiErr.Error()).To(Equal("Password is incorrect, please try again."))
		Expect(config.AccessToken()).To(BeEmpty())
	})

	It("returns a failure response when an error occurs during login", func() {
		ts, handler, config = setupAuthDependencies(errorLoginRequest)

		auth := NewUAAAuthenticationRepository(gateway, config)
		apiErr := auth.Authenticate(map[string]string{
			"username": "foo@example.com",
			"password": "bar",
		})

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).To(HaveOccurred())
		Expect(apiErr.Error()).To(Equal("Server error, status code: 500, error code: , message: "))
		Expect(config.AccessToken()).To(BeEmpty())
	})

	It("returns an error response when the UAA has an error but still returns a 200", func() {
		ts, handler, config = setupAuthDependencies(errorMaskedAsSuccessLoginRequest)

		auth := NewUAAAuthenticationRepository(gateway, config)
		apiErr := auth.Authenticate(map[string]string{
			"username": "foo@example.com",
			"password": "bar",
		})

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).To(HaveOccurred())
		Expect(apiErr.Error()).To(Equal("Authentication Server error: I/O error: uaa.10.244.0.22.xip.io; nested exception is java.net.UnknownHostException: uaa.10.244.0.22.xip.io"))
		Expect(config.AccessToken()).To(BeEmpty())
	})

	It("gets the login prompts", func() {
		ts, handler, config = setupAuthDependencies(loginInfoRequest)
		auth := NewUAAAuthenticationRepository(gateway, config)

		prompts, apiErr := auth.GetLoginPrompts()
		Expect(apiErr).NotTo(HaveOccurred())
		Expect(prompts).To(Equal(map[string]configuration.AuthPrompt{
			"username": configuration.AuthPrompt{
				DisplayName: "Email",
				Type:        configuration.AuthPromptTypeText,
			},
			"pin": configuration.AuthPrompt{
				DisplayName: "PIN Number",
				Type:        configuration.AuthPromptTypePassword,
			},
		}))
	})

	It("returns a failure response when the login info API fails", func() {
		ts, handler, config = setupAuthDependencies(loginInfoFailureRequest)
		auth := NewUAAAuthenticationRepository(gateway, config)

		prompts, apiErr := auth.GetLoginPrompts()
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).To(HaveOccurred())
		Expect(prompts).To(BeEmpty())
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

var successfulLoginMatcher = func(request *http.Request) {
	err := request.ParseForm()
	if err != nil {
		Fail(fmt.Sprintf("Failed to parse form: %s", err))
		return
	}

	Expect(request.Form.Get("username")).To(Equal("foo@example.com"))
	Expect(request.Form.Get("password")).To(Equal("bar"))
	Expect(request.Form.Get("grant_type")).To(Equal("password"))
	Expect(request.Form.Get("scope")).To(Equal(""))
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
{
	"error": {
		"error": "rest_client_error",
		"error_description": "I/O error: uaa.10.244.0.22.xip.io; nested exception is java.net.UnknownHostException: uaa.10.244.0.22.xip.io"
	}
}
`},
}

var loginInfoRequest = testnet.TestRequest{
	Method: "GET",
	Path:   "/login",
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `
{
	"timestamp":"2013-12-18T11:26:53-0700",
	"app":{
		"artifact":"cloudfoundry-identity-uaa",
		"description":"User Account and Authentication Service",
		"name":"UAA",
		"version":"1.4.7"
	},
	"commit_id":"2701cc8",
	"prompts":{
		"username": ["text","Email"],
		"pin": ["password", "PIN Number"]
	}
}`,
	},
}

var loginInfoFailureRequest = testnet.TestRequest{
	Method: "GET",
	Path:   "/login",
	Response: testnet.TestResponse{
		Status: http.StatusInternalServerError,
	},
}

func setupAuthDependencies(request testnet.TestRequest) (*httptest.Server, *testnet.TestHandler, configuration.ReadWriter) {
	ts, handler := testnet.NewServer([]testnet.TestRequest{request})
	config := testconfig.NewRepository()
	config.SetAuthenticationEndpoint(ts.URL)

	return ts, handler, config
}
