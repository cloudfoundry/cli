package net_test

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/errors"
	. "cf/net"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
	"time"
)

var _ = Describe("Gateway", func() {
	var ccGateway Gateway
	var uaaGateway Gateway
	var config configuration.ReadWriter
	var authRepo api.AuthenticationRepository

	BeforeEach(func() {
		ccGateway = NewCloudControllerGateway()
		uaaGateway = NewUAAGateway()
	})

	It("TestNewRequest", func() {
		request, apiResponse := ccGateway.NewRequest("GET", "https://example.com/v2/apps", "BEARER my-access-token", nil)

		Expect(apiResponse).NotTo(HaveOccurred())
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
					fmt.Fprintf(writer, `{ "entity": { "status": "%s" } }`, jobStatus)
				default:
					writer.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(writer, `"Unexpected request path '%s'"`, request.URL.Path)
				}
			}))

			authServer, _ = testnet.NewTLSServer([]testnet.TestRequest{})

			config, authRepo = createAuthenticationRepository(apiServer, authServer)
			ccGateway.SetTokenRefresher(authRepo)
			ccGateway.PollingThrottle = 3 * time.Millisecond

			ccGateway.AddTrustedCerts(apiServer.TLS.Certificates)
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
			_, apiResponse := ccGateway.PerformPollingRequestForJSONResponse(request, new(struct{}), 500*time.Millisecond)
			Expect(apiResponse).NotTo(HaveOccurred())
		})

		It("returns an error if jobs takes longer than the timeout", func() {
			request, _ := ccGateway.NewRequest("GET", config.ApiEndpoint()+"/v2/foo", config.AccessToken(), nil)
			_, apiResponse := ccGateway.PerformPollingRequestForJSONResponse(request, new(struct{}), 10*time.Millisecond)
			Expect(apiResponse).To(HaveOccurred())
			Expect(apiResponse.Error()).To(ContainSubstring("timed out"))
		})
	})

	Describe("when uploading a file", func() {
		var err error
		var request *Request
		var apiResponse errors.Error
		var apiServer *httptest.Server
		var authServer *httptest.Server
		var fileToUpload *os.File

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
			ccGateway.AddTrustedCerts(apiServer.TLS.Certificates)

			request, apiResponse = ccGateway.NewRequest("POST", config.ApiEndpoint()+"/v2/foo", config.AccessToken(), fileToUpload)
		})

		AfterEach(func() {
			apiServer.Close()
			authServer.Close()
			fileToUpload.Close()
			os.Remove(fileToUpload.Name())
		})

		It("sets the content length to the size of the file", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(apiResponse).NotTo(HaveOccurred())
			Expect(request.HttpReq.ContentLength).To(Equal(int64(13)))
		})

		Describe("when the access token expires during the upload", func() {
			It("successfully re-sends the file on the second request", func() {
				apiResponse = ccGateway.PerformRequest(request)
				Expect(apiResponse.Error()).To(BeEmpty())
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

			uaaGateway.AddTrustedCerts(authServer.TLS.Certificates)
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
			ccGateway.AddTrustedCerts(apiServer.TLS.Certificates)

			config, auth := createAuthenticationRepository(apiServer, authServer)
			uaaGateway.SetTokenRefresher(auth)
			request, apiResponse := uaaGateway.NewRequest("POST", config.ApiEndpoint()+"/v2/foo", config.AccessToken(), strings.NewReader("expected body"))
			apiResponse = uaaGateway.PerformRequest(request)

			Expect(apiResponse).NotTo(HaveOccurred())
			Expect(config.AccessToken()).To(Equal("bearer new-access-token"))
			Expect(config.RefreshToken()).To(Equal("new-refresh-token"))
		})

		It("refreshes the token when CC requests fail", func() {
			apiServer := httptest.NewTLSServer(refreshTokenApiEndPoint(
				`{ "code": 1000, "description": "Auth token is invalid" }`,
				testnet.TestResponse{Status: http.StatusOK}))
			defer apiServer.Close()
			ccGateway.AddTrustedCerts(apiServer.TLS.Certificates)

			config, auth := createAuthenticationRepository(apiServer, authServer)
			ccGateway.SetTokenRefresher(auth)
			request, apiResponse := ccGateway.NewRequest("POST", config.ApiEndpoint()+"/v2/foo", config.AccessToken(), strings.NewReader("expected body"))
			apiResponse = ccGateway.PerformRequest(request)

			Expect(apiResponse).NotTo(HaveOccurred())
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
			ccGateway.AddTrustedCerts(apiServer.TLS.Certificates)

			config, auth := createAuthenticationRepository(apiServer, authServer)
			uaaGateway.SetTokenRefresher(auth)
			request, apiResponse := uaaGateway.NewRequest("POST", config.ApiEndpoint()+"/v2/foo", config.AccessToken(), strings.NewReader("expected body"))
			apiResponse = uaaGateway.PerformRequest(request)

			Expect(apiResponse).To(HaveOccurred())
			Expect(apiResponse.ErrorCode()).To(Equal("333"))
		})

		It("returns a failure response when token refresh fails after a CC request", func() {
			apiServer := httptest.NewTLSServer(refreshTokenApiEndPoint(
				`{ "code": 1000, "description": "Auth token is invalid" }`,
				testnet.TestResponse{Status: http.StatusBadRequest, Body: `{
					"code": 333, "description": "bad request"
				}`}))
			defer apiServer.Close()
			ccGateway.AddTrustedCerts(apiServer.TLS.Certificates)

			config, auth := createAuthenticationRepository(apiServer, authServer)
			ccGateway.SetTokenRefresher(auth)
			request, apiResponse := ccGateway.NewRequest("POST", config.ApiEndpoint()+"/v2/foo", config.AccessToken(), strings.NewReader("expected body"))
			apiResponse = ccGateway.PerformRequest(request)

			Expect(apiResponse).To(HaveOccurred())
			Expect(apiResponse.ErrorCode()).To(Equal("333"))
		})
	})

	It("validates the server's SSL certificates", func() {
		apiServer := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			fmt.Fprintln(writer, `{}`)
		}))
		defer apiServer.Close()

		request, apiResponse := ccGateway.NewRequest("POST", apiServer.URL+"/v2/foo", "the-access-token", nil)
		apiResponse = ccGateway.PerformRequest(request)

		Expect(apiResponse).To(HaveOccurred())
		Expect(apiResponse.Error()).To(ContainSubstring("certificate"))
	})
})

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
	config.SetAuthorizationEndpoint(authServer.URL)
	config.SetApiEndpoint(apiServer.URL)
	config.SetAccessToken("bearer initial-access-token")
	config.SetRefreshToken("initial-refresh-token")

	authGateway := NewUAAGateway()
	authGateway.AddTrustedCerts(authServer.TLS.Certificates)

	authenticator := api.NewUAAAuthenticationRepository(authGateway, config)

	return config, authenticator
}
