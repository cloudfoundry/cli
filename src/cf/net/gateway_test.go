package net_test

import (
	"cf"
	"cf/api"
	"cf/configuration"
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
)

func testRefreshTokenWithSuccess(gateway Gateway, endpoint http.HandlerFunc) {
	config, apiResponse := testRefreshToken(gateway, endpoint)
	Expect(apiResponse.IsSuccessful()).To(BeTrue())
	Expect(config.AccessToken()).To(Equal("bearer new-access-token"))
	Expect(config.RefreshToken()).To(Equal("new-refresh-token"))
}

func testRefreshTokenWithError(gateway Gateway, endpoint http.HandlerFunc) {
	_, apiResponse := testRefreshToken(gateway, endpoint)
	Expect(apiResponse.IsSuccessful()).To(BeFalse())
	Expect(apiResponse.ErrorCode).To(Equal("333"))
}

var refreshTokenApiEndPoint = func(unauthorizedBody string, secondReqResp testnet.TestResponse) http.HandlerFunc {
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

func testRefreshToken(gateway Gateway, endpoint http.HandlerFunc) (config configuration.Reader, apiResponse ApiResponse) {
	authEndpoint := func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintln(
			writer,
			`{ "access_token": "new-access-token", "token_type": "bearer", "refresh_token": "new-refresh-token"}`,
		)
	}

	apiServer := httptest.NewTLSServer(endpoint)
	defer apiServer.Close()

	authServer := httptest.NewTLSServer(http.HandlerFunc(authEndpoint))
	defer authServer.Close()

	config, auth := createAuthenticationRepository(apiServer, authServer)
	gateway.SetTokenRefresher(auth)

	request, apiResponse := gateway.NewRequest("POST", config.ApiEndpoint()+"/v2/foo", config.AccessToken(), strings.NewReader("expected body"))
	Expect(apiResponse.IsNotSuccessful()).To(BeFalse())

	apiResponse = gateway.PerformRequest(request)
	return
}

func createAuthenticationRepository(apiServer *httptest.Server, authServer *httptest.Server) (configuration.ReadWriter, api.AuthenticationRepository) {
	config := testconfig.NewRepository()
	config.SetAuthorizationEndpoint(authServer.URL)
	config.SetApiEndpoint(apiServer.URL)
	config.SetAccessToken("bearer initial-access-token")
	config.SetRefreshToken("initial-refresh-token")

	authGateway := NewUAAGateway()
	authenticator := api.NewUAAAuthenticationRepository(authGateway, config)

	return config, authenticator
}

var _ = Describe("Testing with ginkgo", func() {
	var ccGateway Gateway
	var uaaGateway Gateway

	BeforeEach(func() {
		ccGateway = NewCloudControllerGateway()
		uaaGateway = NewUAAGateway()
	})

	It("TestNewRequest", func() {
		request, apiResponse := ccGateway.NewRequest("GET", "https://example.com/v2/apps", "BEARER my-access-token", nil)

		Expect(apiResponse.IsSuccessful()).To(BeTrue())
		Expect(request.HttpReq.Header.Get("Authorization")).To(Equal("BEARER my-access-token"))
		Expect(request.HttpReq.Header.Get("accept")).To(Equal("application/json"))
		Expect(request.HttpReq.Header.Get("User-Agent")).To(Equal("go-cli " + cf.Version + " / " + runtime.GOOS))
	})

	Describe("when uploading a file", func() {
		var err error
		var request *Request
		var apiResponse ApiResponse
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
			Expect(apiResponse.IsSuccessful()).To(BeTrue())
			Expect(request.HttpReq.ContentLength).To(Equal(int64(13)))
		})

		Describe("when the access token expires during the upload", func() {
			It("successfully re-sends the file on the second request", func() {
				apiResponse = ccGateway.PerformRequest(request)
				Expect(apiResponse.Message).To(BeEmpty())
			})
		})
	})

	It("TestRefreshingTheTokenWithUAARequest", func() {
		endpoint := refreshTokenApiEndPoint(
			`{ "error": "invalid_token", "error_description": "Auth token is invalid" }`,
			testnet.TestResponse{Status: http.StatusOK},
		)

		testRefreshTokenWithSuccess(uaaGateway, endpoint)
	})

	It("TestRefreshingTheTokenWithUAARequestAndReturningError", func() {
		endpoint := refreshTokenApiEndPoint(
			`{ "error": "invalid_token", "error_description": "Auth token is invalid" }`,
			testnet.TestResponse{Status: http.StatusBadRequest, Body: `{
			"error": "333", "error_description": "bad request"
		}`},
		)

		testRefreshTokenWithError(uaaGateway, endpoint)
	})

	It("TestRefreshingTheTokenWithCloudControllerRequest", func() {
		endpoint := refreshTokenApiEndPoint(
			`{ "code": 1000, "description": "Auth token is invalid" }`,
			testnet.TestResponse{Status: http.StatusOK},
		)

		testRefreshTokenWithSuccess(ccGateway, endpoint)
	})

	It("TestRefreshingTheTokenWithCloudControllerRequestAndReturningError", func() {
		endpoint := refreshTokenApiEndPoint(
			`{ "code": 1000, "description": "Auth token is invalid" }`,
			testnet.TestResponse{Status: http.StatusBadRequest, Body: `{
			"code": 333, "description": "bad request"
		}`},
		)

		testRefreshTokenWithError(ccGateway, endpoint)
	})
})
