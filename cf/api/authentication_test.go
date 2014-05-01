package api_test

import (
	"encoding/base64"
	"fmt"
	. "github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("AuthenticationRepository", func() {
	var (
		gateway    net.Gateway
		testServer *httptest.Server
		handler    *testnet.TestHandler
		config     configuration.ReadWriter
		auth       AuthenticationRepository
	)

	BeforeEach(func() {
		config = testconfig.NewRepository()
		gateway = net.NewUAAGateway(config)
		auth = NewUAAAuthenticationRepository(gateway, config)
	})

	AfterEach(func() {
		testServer.Close()
	})

	var setupTestServer = func(request testnet.TestRequest) {
		testServer, handler = testnet.NewServer([]testnet.TestRequest{request})
		config.SetAuthenticationEndpoint(testServer.URL)
	}

	Describe("authenticating", func() {
		var err error

		JustBeforeEach(func() {
			err = auth.Authenticate(map[string]string{
				"username": "foo@example.com",
				"password": "bar",
			})
		})

		Describe("when login succeeds", func() {
			BeforeEach(func() {
				setupTestServer(successfulLoginRequest)
			})

			It("stores the access and refresh tokens in the config", func() {
				Expect(handler).To(testnet.HaveAllRequestsCalled())
				Expect(err).NotTo(HaveOccurred())
				Expect(config.AuthenticationEndpoint()).To(Equal(testServer.URL))
				Expect(config.AccessToken()).To(Equal("BEARER my_access_token"))
				Expect(config.RefreshToken()).To(Equal("my_refresh_token"))
			})
		})

		Describe("when login fails", func() {
			BeforeEach(func() {
				setupTestServer(unsuccessfulLoginRequest)
			})

			It("returns an error", func() {
				Expect(handler).To(testnet.HaveAllRequestsCalled())
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("Credentials were rejected, please try again."))
				Expect(config.AccessToken()).To(BeEmpty())
				Expect(config.RefreshToken()).To(BeEmpty())
			})
		})

		Describe("when an error occurs during login", func() {
			BeforeEach(func() {
				setupTestServer(errorLoginRequest)
			})

			It("returns a failure response", func() {
				Expect(handler).To(testnet.HaveAllRequestsCalled())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Server error, status code: 500, error code: , message: "))
				Expect(config.AccessToken()).To(BeEmpty())
			})
		})

		Describe("when the UAA server has an error but still returns a 200", func() {
			BeforeEach(func() {
				setupTestServer(errorMaskedAsSuccessLoginRequest)
			})

			It("returns an error", func() {
				Expect(handler).To(testnet.HaveAllRequestsCalled())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("I/O error: uaa.10.244.0.22.xip.io; nested exception is java.net.UnknownHostException: uaa.10.244.0.22.xip.io"))
				Expect(config.AccessToken()).To(BeEmpty())
			})
		})
	})

	Describe("getting login info", func() {
		var (
			apiErr  error
			prompts map[string]configuration.AuthPrompt
		)

		JustBeforeEach(func() {
			prompts, apiErr = auth.GetLoginPromptsAndSaveUAAServerURL()
		})

		Describe("when the login info API succeeds", func() {
			BeforeEach(func() {
				setupTestServer(loginServerLoginRequest)
			})

			It("does not return an error", func() {
				Expect(apiErr).NotTo(HaveOccurred())
			})

			It("gets the login prompts", func() {
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

			It("saves the UAA server to the config", func() {
				Expect(config.UaaEndpoint()).To(Equal("https://uaa.run.pivotal.io"))
			})
		})

		Describe("when the login info API fails", func() {
			BeforeEach(func() {
				setupTestServer(loginServerLoginFailureRequest)
			})

			It("returns a failure response when the login info API fails", func() {
				Expect(handler).To(testnet.HaveAllRequestsCalled())
				Expect(apiErr).To(HaveOccurred())
				Expect(prompts).To(BeEmpty())
			})
		})

		Context("when the response does not contain links", func() {
			BeforeEach(func() {
				setupTestServer(uaaServerLoginRequest)
			})

			It("presumes that the authorization server is the UAA", func() {
				Expect(config.UaaEndpoint()).To(Equal(config.AuthenticationEndpoint()))
			})
		})
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

var loginServerLoginRequest = testnet.TestRequest{
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
	"links":{
	    "register":"https://console.run.pivotal.io/register",
	    "passwd":"https://console.run.pivotal.io/password_resets/new",
	    "home":"https://console.run.pivotal.io",
	    "support":"https://support.cloudfoundry.com/home",
	    "login":"https://login.run.pivotal.io",
	    "uaa":"https://uaa.run.pivotal.io"
	 },
	"prompts":{
		"username": ["text","Email"],
		"pin": ["password", "PIN Number"]
	}
}`,
	},
}

var loginServerLoginFailureRequest = testnet.TestRequest{
	Method: "GET",
	Path:   "/login",
	Response: testnet.TestResponse{
		Status: http.StatusInternalServerError,
	},
}

var uaaServerLoginRequest = testnet.TestRequest{
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
