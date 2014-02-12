package net_test

import (
	"cf"
	"cf/api"
	"cf/configuration"
	. "cf/net"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
)

func testRefreshTokenWithSuccess(t mr.TestingT, gateway Gateway, endpoint http.HandlerFunc) {
	config, apiResponse := testRefreshToken(t, gateway, endpoint)
	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, config.AccessToken(), "bearer new-access-token")
	assert.Equal(t, config.RefreshToken(), "new-refresh-token")
}

func testRefreshTokenWithError(t mr.TestingT, gateway Gateway, endpoint http.HandlerFunc) {
	_, apiResponse := testRefreshToken(t, gateway, endpoint)
	assert.False(t, apiResponse.IsSuccessful())
	assert.Equal(t, apiResponse.ErrorCode, "333")
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

func testRefreshToken(t mr.TestingT, gateway Gateway, endpoint http.HandlerFunc) (config configuration.Reader, apiResponse ApiResponse) {
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

	config, auth := createAuthenticationRepository(t, apiServer, authServer)
	gateway.SetTokenRefresher(auth)

	request, apiResponse := gateway.NewRequest("POST", config.ApiEndpoint()+"/v2/foo", config.AccessToken(), strings.NewReader("expected body"))
	assert.False(t, apiResponse.IsNotSuccessful())

	apiResponse = gateway.PerformRequest(request)
	return
}

func createAuthenticationRepository(t mr.TestingT, apiServer *httptest.Server, authServer *httptest.Server) (configuration.ReadWriter, api.AuthenticationRepository) {
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

	It("TestNewRequest", func() {
		gateway := NewCloudControllerGateway()

		request, apiResponse := gateway.NewRequest("GET", "https://example.com/v2/apps", "BEARER my-access-token", nil)

		assert.True(mr.T(), apiResponse.IsSuccessful())
		Expect(request.HttpReq.Header.Get("Authorization")).To(Equal("BEARER my-access-token"))
		assert.Equal(mr.T(), request.HttpReq.Header.Get("accept"), "application/json")
		assert.Equal(mr.T(), request.HttpReq.Header.Get("User-Agent"), "go-cli "+cf.Version+" / "+runtime.GOOS)
	})

	It("TestNewRequestWithAFileBody", func() {
		gateway := NewCloudControllerGateway()

		body, err := os.Open("../../fixtures/hello_world.txt")
		assert.NoError(mr.T(), err)
		request, apiResponse := gateway.NewRequest("GET", "https://example.com/v2/apps", "BEARER my-access-token", body)

		assert.True(mr.T(), apiResponse.IsSuccessful())
		assert.Equal(mr.T(), request.HttpReq.ContentLength, 12)
	})

	It("TestRefreshingTheTokenWithUAARequest", func() {
		gateway := NewUAAGateway()
		endpoint := refreshTokenApiEndPoint(
			`{ "error": "invalid_token", "error_description": "Auth token is invalid" }`,
			testnet.TestResponse{Status: http.StatusOK},
		)

		testRefreshTokenWithSuccess(mr.T(), gateway, endpoint)
	})

	It("TestRefreshingTheTokenWithUAARequestAndReturningError", func() {
		gateway := NewUAAGateway()
		endpoint := refreshTokenApiEndPoint(
			`{ "error": "invalid_token", "error_description": "Auth token is invalid" }`,
			testnet.TestResponse{Status: http.StatusBadRequest, Body: `{
			"error": "333", "error_description": "bad request"
		}`},
		)

		testRefreshTokenWithError(mr.T(), gateway, endpoint)
	})

	It("TestRefreshingTheTokenWithCloudControllerRequest", func() {
		gateway := NewCloudControllerGateway()
		endpoint := refreshTokenApiEndPoint(
			`{ "code": 1000, "description": "Auth token is invalid" }`,
			testnet.TestResponse{Status: http.StatusOK},
		)

		testRefreshTokenWithSuccess(mr.T(), gateway, endpoint)
	})
	It("TestRefreshingTheTokenWithCloudControllerRequestAndReturningError", func() {

		gateway := NewCloudControllerGateway()
		endpoint := refreshTokenApiEndPoint(
			`{ "code": 1000, "description": "Auth token is invalid" }`,
			testnet.TestResponse{Status: http.StatusBadRequest, Body: `{
			"code": 333, "description": "bad request"
		}`},
		)

		testRefreshTokenWithError(mr.T(), gateway, endpoint)
	})
})
