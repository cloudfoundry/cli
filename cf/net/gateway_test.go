/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/github.com/cloudfoundry/cli/cf/commands/application/delete_app_test.go
   src/github.com/cloudfoundry/cli/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package net_test

import (
	"crypto/tls"
	"fmt"
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"
)

var _ = Describe("Gateway", func() {
	var (
		ccGateway  Gateway
		uaaGateway Gateway
		config     configuration.ReadWriter
		authRepo   api.AuthenticationRepository
	)

	BeforeEach(func() {
		config = testconfig.NewRepository()
		ccGateway = NewCloudControllerGateway(config)
		uaaGateway = NewUAAGateway(config)
	})

	It("TestNewRequest", func() {
		request, apiErr := ccGateway.NewRequest("GET", "https://example.com/v2/apps", "BEARER my-access-token", nil)

		Expect(apiErr).NotTo(HaveOccurred())
		Expect(request.HttpReq.Header.Get("Authorization")).To(Equal("BEARER my-access-token"))
		Expect(request.HttpReq.Header.Get("accept")).To(Equal("application/json"))
		Expect(request.HttpReq.Header.Get("User-Agent")).To(Equal("go-cli " + cf.Version + " / " + runtime.GOOS))
	})

	Describe("making an async request", func() {
		var jobStatus string
		var apiServer *httptest.Server
		var authServer *httptest.Server

		BeforeEach(func() {
			jobStatus = "queued"

			apiServer = httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				switch request.URL.Path {
				case "/v2/foo":
					fmt.Fprintln(writer, `{ "metadata": { "url": "/v2/jobs/the-job-guid" } }`)
				case "/v2/jobs/the-job-guid":
					fmt.Fprintf(writer, `
					{
						"entity": {
							"status": "%s",
							"error_details": {
								"description": "he's dead, Jim"
							}
						}
					}`, jobStatus)
				default:
					writer.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(writer, `"Unexpected request path '%s'"`, request.URL.Path)
				}
			}))

			authServer, _ = testnet.NewTLSServer([]testnet.TestRequest{})

			config, authRepo = createAuthenticationRepository(apiServer, authServer)
			ccGateway.SetTokenRefresher(authRepo)
			ccGateway.PollingThrottle = 3 * time.Millisecond

			ccGateway.SetTrustedCerts(apiServer.TLS.Certificates)
		})

		AfterEach(func() {
			apiServer.Close()
			authServer.Close()
		})

		It("returns the last response if the job completes before the timeout", func() {
			go func() {
				time.Sleep(25 * time.Millisecond)
				jobStatus = "finished"
			}()

			request, _ := ccGateway.NewRequest("GET", config.ApiEndpoint()+"/v2/foo", config.AccessToken(), nil)
			_, apiErr := ccGateway.PerformPollingRequestForJSONResponse(request, new(struct{}), 500*time.Millisecond)
			Expect(apiErr).NotTo(HaveOccurred())
		})

		It("returns an error with the right message when the job fails", func() {
			go func() {
				time.Sleep(25 * time.Millisecond)
				jobStatus = "failed"
			}()

			request, _ := ccGateway.NewRequest("GET", config.ApiEndpoint()+"/v2/foo", config.AccessToken(), nil)
			_, apiErr := ccGateway.PerformPollingRequestForJSONResponse(request, new(struct{}), 500*time.Millisecond)
			Expect(apiErr.Error()).To(ContainSubstring("he's dead, Jim"))
		})

		It("returns an error if jobs takes longer than the timeout", func() {
			request, _ := ccGateway.NewRequest("GET", config.ApiEndpoint()+"/v2/foo", config.AccessToken(), nil)
			_, apiErr := ccGateway.PerformPollingRequestForJSONResponse(request, new(struct{}), 10*time.Millisecond)
			Expect(apiErr).To(HaveOccurred())
			Expect(apiErr.Error()).To(ContainSubstring("timed out"))
		})
	})

	Describe("when uploading a file", func() {
		var (
			err          error
			request      *Request
			apiErr       error
			apiServer    *httptest.Server
			authServer   *httptest.Server
			fileToUpload *os.File
		)

		BeforeEach(func() {
			apiServer = httptest.NewTLSServer(refreshTokenApiEndPoint(
				`{ "code": 1000, "description": "Auth token is invalid" }`,
				testnet.TestResponse{Status: http.StatusOK},
			))

			authServer = httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				fmt.Fprintln(
					writer,
					`{ "access_token": "new-access-token", "token_type": "bearer", "refresh_token": "new-refresh-token"}`)
			}))

			fileToUpload, err = ioutil.TempFile("", "test-gateway")
			strings.NewReader("expected body").WriteTo(fileToUpload)

			config, auth := createAuthenticationRepository(apiServer, authServer)
			ccGateway.SetTokenRefresher(auth)
			ccGateway.SetTrustedCerts(apiServer.TLS.Certificates)

			request, apiErr = ccGateway.NewRequest("POST", config.ApiEndpoint()+"/v2/foo", config.AccessToken(), fileToUpload)
		})

		AfterEach(func() {
			apiServer.Close()
			authServer.Close()
			fileToUpload.Close()
			os.Remove(fileToUpload.Name())
		})

		It("sets the content length to the size of the file", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(apiErr).NotTo(HaveOccurred())
			Expect(request.HttpReq.ContentLength).To(Equal(int64(13)))
		})

		Describe("when the access token expires during the upload", func() {
			It("successfully re-sends the file on the second request", func() {
				_, apiErr = ccGateway.PerformRequest(request)
				Expect(apiErr).NotTo(HaveOccurred())
			})
		})
	})

	Describe("refreshing the auth token", func() {
		var authServer *httptest.Server

		BeforeEach(func() {
			authServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, `{
				 	"access_token": "new-access-token",
				 	"token_type": "bearer",
					"refresh_token": "new-refresh-token"
				}`)
			}))

			uaaGateway.SetTrustedCerts(authServer.TLS.Certificates)
		})

		AfterEach(func() {
			authServer.Close()
		})

		It("refreshes the token when UAA requests fail", func() {
			apiServer := httptest.NewTLSServer(refreshTokenApiEndPoint(
				`{ "error": "invalid_token", "error_description": "Auth token is invalid" }`,
				testnet.TestResponse{Status: http.StatusOK},
			))
			defer apiServer.Close()
			ccGateway.SetTrustedCerts(apiServer.TLS.Certificates)

			config, auth := createAuthenticationRepository(apiServer, authServer)
			uaaGateway.SetTokenRefresher(auth)
			request, apiErr := uaaGateway.NewRequest("POST", config.ApiEndpoint()+"/v2/foo", config.AccessToken(), strings.NewReader("expected body"))
			_, apiErr = uaaGateway.PerformRequest(request)

			Expect(apiErr).NotTo(HaveOccurred())
			Expect(config.AccessToken()).To(Equal("bearer new-access-token"))
			Expect(config.RefreshToken()).To(Equal("new-refresh-token"))
		})

		It("refreshes the token when CC requests fail", func() {
			apiServer := httptest.NewTLSServer(refreshTokenApiEndPoint(
				`{ "code": 1000, "description": "Auth token is invalid" }`,
				testnet.TestResponse{Status: http.StatusOK}))
			defer apiServer.Close()
			ccGateway.SetTrustedCerts(apiServer.TLS.Certificates)

			config, auth := createAuthenticationRepository(apiServer, authServer)
			ccGateway.SetTokenRefresher(auth)
			request, apiErr := ccGateway.NewRequest("POST", config.ApiEndpoint()+"/v2/foo", config.AccessToken(), strings.NewReader("expected body"))
			_, apiErr = ccGateway.PerformRequest(request)

			Expect(apiErr).NotTo(HaveOccurred())
			Expect(config.AccessToken()).To(Equal("bearer new-access-token"))
			Expect(config.RefreshToken()).To(Equal("new-refresh-token"))
		})

		It("returns a failure response when token refresh fails after a UAA request", func() {
			apiServer := httptest.NewTLSServer(refreshTokenApiEndPoint(
				`{ "error": "invalid_token", "error_description": "Auth token is invalid" }`,
				testnet.TestResponse{Status: http.StatusBadRequest, Body: `{
					"error": "333", "error_description": "bad request"
				}`}))
			defer apiServer.Close()
			ccGateway.SetTrustedCerts(apiServer.TLS.Certificates)

			config, auth := createAuthenticationRepository(apiServer, authServer)
			uaaGateway.SetTokenRefresher(auth)
			request, apiErr := uaaGateway.NewRequest("POST", config.ApiEndpoint()+"/v2/foo", config.AccessToken(), strings.NewReader("expected body"))
			_, apiErr = uaaGateway.PerformRequest(request)

			Expect(apiErr).To(HaveOccurred())
			Expect(apiErr.(errors.HttpError).ErrorCode()).To(Equal("333"))
		})

		It("returns a failure response when token refresh fails after a CC request", func() {
			apiServer := httptest.NewTLSServer(refreshTokenApiEndPoint(
				`{ "code": 1000, "description": "Auth token is invalid" }`,
				testnet.TestResponse{Status: http.StatusBadRequest, Body: `{
					"code": 333, "description": "bad request"
				}`}))
			defer apiServer.Close()
			ccGateway.SetTrustedCerts(apiServer.TLS.Certificates)

			config, auth := createAuthenticationRepository(apiServer, authServer)
			ccGateway.SetTokenRefresher(auth)
			request, apiErr := ccGateway.NewRequest("POST", config.ApiEndpoint()+"/v2/foo", config.AccessToken(), strings.NewReader("expected body"))
			_, apiErr = ccGateway.PerformRequest(request)

			Expect(apiErr).To(HaveOccurred())
			Expect(apiErr.(errors.HttpError).ErrorCode()).To(Equal("333"))
		})
	})

	Describe("SSL certificate validation errors", func() {
		var (
			request   *Request
			apiServer *httptest.Server
		)

		BeforeEach(func() {
			apiServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				fmt.Fprintln(w, `{}`)
			}))
			request, _ = ccGateway.NewRequest("POST", apiServer.URL+"/v2/foo", "the-access-token", nil)
		})

		AfterEach(func() {
			apiServer.Close()
		})

		Context("when SSL validation is enabled", func() {
			It("returns an invalid cert error if the server's CA is unknown (e.g. cert is self-signed)", func() {
				apiServer.TLS.Certificates = []tls.Certificate{testnet.MakeSelfSignedTLSCert()}

				_, apiErr := ccGateway.PerformRequest(request)
				certErr, ok := apiErr.(*errors.InvalidSSLCert)
				Expect(ok).To(BeTrue())
				Expect(certErr.URL).To(Equal(getHost(apiServer.URL)))
				Expect(certErr.Reason).To(Equal("unknown authority"))
			})

			It("returns an invalid cert error if the server's cert doesn't match its host", func() {
				apiServer.TLS.Certificates = []tls.Certificate{testnet.MakeTLSCertWithInvalidHost()}

				_, apiErr := ccGateway.PerformRequest(request)
				certErr, ok := apiErr.(*errors.InvalidSSLCert)
				Expect(ok).To(BeTrue())
				Expect(certErr.URL).To(Equal(getHost(apiServer.URL)))
				if runtime.GOOS != "windows" {
					Expect(certErr.Reason).To(Equal("not valid for the requested host"))
				}
			})

			It("returns an invalid cert error if the server's cert has expired", func() {
				apiServer.TLS.Certificates = []tls.Certificate{testnet.MakeExpiredTLSCert()}

				_, apiErr := ccGateway.PerformRequest(request)
				certErr, ok := apiErr.(*errors.InvalidSSLCert)
				Expect(ok).To(BeTrue())
				Expect(certErr.URL).To(Equal(getHost(apiServer.URL)))
				if runtime.GOOS != "windows" {
					Expect(certErr.Reason).To(Equal(""))
				}
			})
		})

		Context("when SSL validation is disabled", func() {
			BeforeEach(func() {
				apiServer.TLS.Certificates = []tls.Certificate{testnet.MakeExpiredTLSCert()}
				config.SetSSLDisabled(true)
			})

			It("succeeds", func() {
				_, apiErr := ccGateway.PerformRequest(request)
				Expect(apiErr).NotTo(HaveOccurred())
			})
		})

	})

	Describe("collecting warnings", func() {
		var (
			apiServer  *httptest.Server
			authServer *httptest.Server
		)

		BeforeEach(func() {
			apiServer = httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				switch request.URL.Path {
				case "/v2/happy":
					fmt.Fprintln(writer, `{ "metadata": { "url": "/v2/jobs/the-job-guid" } }`)
				case "/v2/warning1":
					writer.Header().Add("X-Cf-Warnings", url.QueryEscape("Something not too awful has happened"))
					fmt.Fprintln(writer, `{ "metadata": { "url": "/v2/jobs/the-job-guid" } }`)
				case "/v2/warning2":
					writer.Header().Add("X-Cf-Warnings", url.QueryEscape("Something a little awful"))
					writer.Header().Add("X-Cf-Warnings", url.QueryEscape("Don't worry, but be careful"))
					writer.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(writer, `{ "key": "value" }`)
				}
			}))

			authServer, _ = testnet.NewTLSServer([]testnet.TestRequest{})

			config, authRepo = createAuthenticationRepository(apiServer, authServer)
			ccGateway.SetTokenRefresher(authRepo)
			ccGateway.PollingThrottle = 3 * time.Millisecond

			ccGateway.SetTrustedCerts(apiServer.TLS.Certificates)

			config, authRepo = createAuthenticationRepository(apiServer, authServer)
		})

		AfterEach(func() {
			apiServer.Close()
			authServer.Close()
		})

		It("saves all X-Cf-Warnings headers and exposes them", func() {
			request, _ := ccGateway.NewRequest("GET", config.ApiEndpoint()+"/v2/happy", config.AccessToken(), nil)
			ccGateway.PerformRequest(request)
			request, _ = ccGateway.NewRequest("GET", config.ApiEndpoint()+"/v2/warning1", config.AccessToken(), nil)
			ccGateway.PerformRequest(request)
			request, _ = ccGateway.NewRequest("GET", config.ApiEndpoint()+"/v2/warning2", config.AccessToken(), nil)
			ccGateway.PerformRequest(request)

			Expect(ccGateway.Warnings()).To(Equal(
				[]string{"Something not too awful has happened", "Something a little awful", "Don't worry, but be careful"},
			))
		})

		It("defaults warnings to an empty slice", func() {
			Expect(ccGateway.Warnings()).ToNot(BeNil())
		})
	})
})

func getHost(urlString string) string {
	url, err := url.Parse(urlString)
	if err != nil {
		panic(err)
	}
	return url.Host
}

func refreshTokenApiEndPoint(unauthorizedBody string, secondReqResp testnet.TestResponse) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var jsonResponse string

		bodyBytes, err := ioutil.ReadAll(request.Body)
		if err != nil || string(bodyBytes) != "expected body" {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		switch request.Header.Get("Authorization") {
		case "bearer initial-access-token":
			writer.WriteHeader(http.StatusUnauthorized)
			jsonResponse = unauthorizedBody
		case "bearer new-access-token":
			writer.WriteHeader(secondReqResp.Status)
			jsonResponse = secondReqResp.Body
		default:
			writer.WriteHeader(http.StatusInternalServerError)
		}

		fmt.Fprintln(writer, jsonResponse)
	}
}

func createAuthenticationRepository(apiServer *httptest.Server, authServer *httptest.Server) (configuration.ReadWriter, api.AuthenticationRepository) {
	config := testconfig.NewRepository()
	config.SetAuthenticationEndpoint(authServer.URL)
	config.SetApiEndpoint(apiServer.URL)
	config.SetAccessToken("bearer initial-access-token")
	config.SetRefreshToken("initial-refresh-token")

	authGateway := NewUAAGateway(config)
	authGateway.SetTrustedCerts(authServer.TLS.Certificates)

	authenticator := api.NewUAAAuthenticationRepository(authGateway, config)

	return config, authenticator
}
