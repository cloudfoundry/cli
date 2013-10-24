package net_test

import (
	"cf"
	"cf/api"
	"cf/configuration"
	. "cf/net"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
	"testing"
)

func TestNewRequest(t *testing.T) {
	// arbitrarily picking cloud controller to test
	gateway := NewCloudControllerGateway()

	request, apiResponse := gateway.NewRequest("GET", "https://example.com/v2/apps", "BEARER my-access-token", nil)

	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, request.HttpReq.Header.Get("Authorization"), "BEARER my-access-token")
	assert.Equal(t, request.HttpReq.Header.Get("accept"), "application/json")
	assert.Equal(t, request.HttpReq.Header.Get("User-Agent"), "go-cli "+cf.Version+" / "+runtime.GOOS)
}

func TestNewRequestWithAFileBody(t *testing.T) {
	// arbitrarily picking cloud controller to test
	gateway := NewCloudControllerGateway()

	body, err := os.Open("../../fixtures/hello_world.txt")
	assert.NoError(t, err)
	request, apiResponse := gateway.NewRequest("GET", "https://example.com/v2/apps", "BEARER my-access-token", body)

	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, request.HttpReq.ContentLength, 12) // 12 is the size of the file
}

func TestRefreshingTheTokenWithUAARequest(t *testing.T) {
	gateway := NewUAAGateway()
	endpoint := refreshTokenApiEndPoint(
		`{ "error": "invalid_token", "error_description": "Auth token is invalid" }`,
		testnet.TestResponse{Status: http.StatusOK},
	)

	testRefreshTokenWithSuccess(t, gateway, endpoint)
}

func TestRefreshingTheTokenWithUAARequestAndReturningError(t *testing.T) {
	gateway := NewUAAGateway()
	endpoint := refreshTokenApiEndPoint(
		`{ "error": "invalid_token", "error_description": "Auth token is invalid" }`,
		testnet.TestResponse{Status: http.StatusBadRequest, Body: `{
			"error": "333", "error_description": "bad request"
		}`},
	)

	testRefreshTokenWithError(t, gateway, endpoint)
}

func TestRefreshingTheTokenWithCloudControllerRequest(t *testing.T) {
	gateway := NewCloudControllerGateway()
	endpoint := refreshTokenApiEndPoint(
		`{ "code": 1000, "description": "Auth token is invalid" }`,
		testnet.TestResponse{Status: http.StatusOK},
	)

	testRefreshTokenWithSuccess(t, gateway, endpoint)
}

func TestRefreshingTheTokenWithCloudControllerRequestAndReturningError(t *testing.T) {
	gateway := NewCloudControllerGateway()
	endpoint := refreshTokenApiEndPoint(
		`{ "code": 1000, "description": "Auth token is invalid" }`,
		testnet.TestResponse{Status: http.StatusBadRequest, Body: `{
			"code": 333, "description": "bad request"
		}`},
	)

	testRefreshTokenWithError(t, gateway, endpoint)
}

func testRefreshTokenWithSuccess(t *testing.T, gateway Gateway, endpoint http.HandlerFunc) {
	apiResponse := testRefreshToken(t, gateway, endpoint)
	assert.True(t, apiResponse.IsSuccessful())

	savedConfig := testconfig.SavedConfiguration
	assert.Equal(t, savedConfig.AccessToken, "bearer new-access-token")
	assert.Equal(t, savedConfig.RefreshToken, "new-refresh-token")
}

func testRefreshTokenWithError(t *testing.T, gateway Gateway, endpoint http.HandlerFunc) {
	apiResponse := testRefreshToken(t, gateway, endpoint)
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

func testRefreshToken(t *testing.T, gateway Gateway, endpoint http.HandlerFunc) (apiResponse ApiResponse) {
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

	request, apiResponse := gateway.NewRequest("POST", config.Target+"/v2/foo", config.AccessToken, strings.NewReader("expected body"))
	assert.False(t, apiResponse.IsNotSuccessful())

	apiResponse = gateway.PerformRequest(request)
	return
}

func createAuthenticationRepository(t *testing.T, apiServer *httptest.Server, authServer *httptest.Server) (*configuration.Configuration, api.AuthenticationRepository) {
	configRepo := testconfig.FakeConfigRepository{}
	configRepo.Delete()
	config, err := configRepo.Get()
	assert.NoError(t, err)

	config.AuthorizationEndpoint = authServer.URL
	config.Target = apiServer.URL
	config.AccessToken = "bearer initial-access-token"
	config.RefreshToken = "initial-refresh-token"

	authGateway := NewUAAGateway()
	authenticator := api.NewUAAAuthenticationRepository(authGateway, configRepo)

	return config, authenticator
}
