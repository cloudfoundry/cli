package authentication_test

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"

	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testnet "code.cloudfoundry.org/cli/util/testhelpers/net"

	. "code.cloudfoundry.org/cli/cf/api/authentication"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("AuthenticationRepository", func() {
	Describe("legacy tests", func() {
		var (
			gateway     net.Gateway
			testServer  *httptest.Server
			handler     *testnet.TestHandler
			config      coreconfig.ReadWriter
			auth        Repository
			dumper      net.RequestDumper
			fakePrinter *tracefakes.FakePrinter
		)

		BeforeEach(func() {
			config = testconfig.NewRepository()
			fakePrinter = new(tracefakes.FakePrinter)
			gateway = net.NewUAAGateway(config, new(terminalfakes.FakeUI), fakePrinter, "")
			dumper = net.NewRequestDumper(fakePrinter)
			auth = NewUAARepository(gateway, config, dumper)
		})

		AfterEach(func() {
			testServer.Close()
		})

		var setupTestServer = func(request testnet.TestRequest) {
			testServer, handler = testnet.NewServer([]testnet.TestRequest{request})
			config.SetAuthenticationEndpoint(testServer.URL)
			config.SetUAAOAuthClient("cf")
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
					Expect(handler).To(HaveAllRequestsCalled())
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
					Expect(handler).To(HaveAllRequestsCalled())
					Expect(err).NotTo(BeNil())
					Expect(err.Error()).To(Equal("Credentials were rejected, please try again."))
					Expect(config.AccessToken()).To(BeEmpty())
					Expect(config.RefreshToken()).To(BeEmpty())
				})
			})

			Context("when the authentication server returns status code 500", func() {
				BeforeEach(func() {
					setupTestServer(errorLoginRequest)
				})

				It("returns a failure response", func() {
					Expect(handler).To(HaveAllRequestsCalled())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("The targeted API endpoint could not be reached."))
					Expect(config.AccessToken()).To(BeEmpty())
				})
			})

			Context("when the authentication server returns status code 502", func() {
				var request testnet.TestRequest

				BeforeEach(func() {
					request = testnet.TestRequest{
						Method: "POST",
						Path:   "/oauth/token",
						Response: testnet.TestResponse{
							Status: http.StatusBadGateway,
						},
					}
					setupTestServer(request)
				})

				It("returns a failure response", func() {
					Expect(handler).To(HaveAllRequestsCalled())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("The targeted API endpoint could not be reached."))
					Expect(config.AccessToken()).To(BeEmpty())
				})
			})

			Describe("when the UAA server has an error but still returns a 200", func() {
				BeforeEach(func() {
					setupTestServer(errorMaskedAsSuccessLoginRequest)
				})

				It("returns an error", func() {
					Expect(handler).To(HaveAllRequestsCalled())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("I/O error: uaa.10.244.0.22.xip.io; nested exception is java.net.UnknownHostException: uaa.10.244.0.22.xip.io"))
					Expect(config.AccessToken()).To(BeEmpty())
				})
			})
		})

		Describe("getting login info", func() {
			var (
				apiErr  error
				prompts map[string]coreconfig.AuthPrompt
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
					Expect(prompts).To(Equal(map[string]coreconfig.AuthPrompt{
						"username": {
							DisplayName: "Email",
							Type:        coreconfig.AuthPromptTypeText,
						},
						"pin": {
							DisplayName: "PIN Number",
							Type:        coreconfig.AuthPromptTypePassword,
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
					Expect(handler).To(HaveAllRequestsCalled())
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

		Describe("refreshing the auth token", func() {
			var apiErr error

			JustBeforeEach(func() {
				_, apiErr = auth.RefreshAuthToken()
			})

			Context("when the refresh token has expired", func() {
				BeforeEach(func() {
					setupTestServer(refreshTokenExpiredRequestError)
				})
				It("the returns the reauthentication error message", func() {
					Expect(apiErr.Error()).To(Equal("Authentication has expired.  Please log back in to re-authenticate.\n\nTIP: Use `cf login -a <endpoint> -u <user> -o <org> -s <space>` to log back in and re-authenticate."))
				})
			})
			Context("when there is a UAA error", func() {
				BeforeEach(func() {
					setupTestServer(errorLoginRequest)
				})

				It("returns the API error", func() {
					Expect(apiErr).NotTo(BeNil())
				})
			})
		})
	})

	Describe("Authorize", func() {
		var (
			uaaServer   *ghttp.Server
			gateway     net.Gateway
			config      coreconfig.ReadWriter
			authRepo    Repository
			dumper      net.RequestDumper
			fakePrinter *tracefakes.FakePrinter
		)

		BeforeEach(func() {
			uaaServer = ghttp.NewServer()
			config = testconfig.NewRepository()
			config.SetUaaEndpoint(uaaServer.URL())
			config.SetSSHOAuthClient("ssh-oauth-client")

			fakePrinter = new(tracefakes.FakePrinter)
			gateway = net.NewUAAGateway(config, new(terminalfakes.FakeUI), fakePrinter, "")
			dumper = net.NewRequestDumper(fakePrinter)
			authRepo = NewUAARepository(gateway, config, dumper)

			uaaServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyHeader(http.Header{"authorization": []string{"auth-token"}}),
					ghttp.VerifyRequest("GET", "/oauth/authorize",
						"response_type=code&grant_type=authorization_code&client_id=ssh-oauth-client",
					),
					ghttp.RespondWith(http.StatusFound, ``, http.Header{
						"Location": []string{"https://www.cloudfoundry.example.com?code=F45jH"},
					}),
				),
			)
		})

		AfterEach(func() {
			uaaServer.Close()
		})

		It("requests the one time code", func() {
			_, err := authRepo.Authorize("auth-token")
			Expect(err).NotTo(HaveOccurred())
			Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("returns the one time code", func() {
			code, err := authRepo.Authorize("auth-token")
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal("F45jH"))
		})

		Context("when the authentication endpoint is malformed", func() {
			BeforeEach(func() {
				config.SetUaaEndpoint(":not-well-formed")
			})

			It("returns an error", func() {
				_, err := authRepo.Authorize("auth-token")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the authorization server does not return a redirect", func() {
			BeforeEach(func() {
				uaaServer.SetHandler(0, ghttp.RespondWith(http.StatusOK, ``))
			})

			It("returns an error", func() {
				_, err := authRepo.Authorize("auth-token")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Authorization server did not redirect with one time code"))
			})
		})

		Context("when the authorization server does not return a redirect", func() {
			BeforeEach(func() {
				config.SetUaaEndpoint("https://127.0.0.1:1")
			})

			It("returns an error", func() {
				_, err := authRepo.Authorize("auth-token")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Error requesting one time code from server"))
			})
		})

		Context("when the authorization server returns multiple codes", func() {
			BeforeEach(func() {
				uaaServer.SetHandler(0, ghttp.RespondWith(http.StatusFound, ``, http.Header{
					"Location": []string{"https://www.cloudfoundry.example.com?code=F45jH&code=LLLLL"},
				}))
			})

			It("returns an error", func() {
				_, err := authRepo.Authorize("auth-token")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Unable to acquire one time code from authorization response"))
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
var refreshTokenExpiredRequestError = testnet.TestRequest{
	Method: "POST",
	Path:   "/oauth/token",
	Response: testnet.TestResponse{
		Status: http.StatusUnauthorized,
		Body: `
{
	"error": "invalid_token",
	"error_description": "Invalid auth token: Invalid refresh token (expired): eyJhbGckjsdfdf"
}
`},
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
