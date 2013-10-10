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
	"runtime"
	"strings"
	testconfig "testhelpers/configuration"
	"testing"
)

func TestNewRequest(t *testing.T) {
	// arbitrarily picking cloud controller to test
	gateway := NewCloudControllerGateway()

	request, apiResponse := gateway.NewRequest("GET", "https://example.com/v2/apps", "BEARER my-access-token", nil)

	assert.False(t, apiResponse.IsNotSuccessful())
	assert.Equal(t, request.Header.Get("Authorization"), "BEARER my-access-token")
	assert.Equal(t, request.Header.Get("accept"), "application/json")
	assert.Equal(t, request.Header.Get("User-Agent"), "go-cli "+cf.Version+" / "+runtime.GOOS)
}

var refreshTokenApiEndPoint = func(unauthorizedBody string) http.HandlerFunc {
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
			writer.WriteHeader(http.StatusOK)
		default:
			writer.WriteHeader(http.StatusInternalServerError)
		}

		fmt.Fprintln(writer, jsonResponse)
	}
}

var refreshTokenAuthEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	jsonResponse := `
	{
	  "access_token": "new-access-token",
	  "token_type": "bearer",
	  "refresh_token": "new-refresh-token"
	}`
	fmt.Fprintln(writer, jsonResponse)
}

var refreshTokenUAAApiEndpoint = refreshTokenApiEndPoint(
	`{ "error": "invalid_token", "error_description": "Auth token is invalid" }`,
)

func TestRefreshingTheTokenWithUAARequest(t *testing.T) {
	uaaServer := httptest.NewTLSServer(http.HandlerFunc(refreshTokenUAAApiEndpoint))
	defer uaaServer.Close()

	authServer := httptest.NewTLSServer(http.HandlerFunc(refreshTokenAuthEndpoint))
	defer authServer.Close()

	configRepo, auth := createAuthenticationRepository(t, uaaServer, authServer)

	gateway := NewUAAGateway()
	gateway.SetTokenRefresher(auth)

	testRefreshToken(t, configRepo, gateway)
}

var refreshTokenCloudControllerApiEndpoint = refreshTokenApiEndPoint(`{ "code": 1000, "description": "Auth token is invalid" }`)

func TestRefreshingTheTokenWithCloudControllerRequest(t *testing.T) {
	ccServer := httptest.NewTLSServer(http.HandlerFunc(refreshTokenCloudControllerApiEndpoint))
	defer ccServer.Close()

	authServer := httptest.NewTLSServer(http.HandlerFunc(refreshTokenAuthEndpoint))
	defer authServer.Close()

	configRepo, auth := createAuthenticationRepository(t, ccServer, authServer)

	gateway := NewCloudControllerGateway()
	gateway.SetTokenRefresher(auth)

	testRefreshToken(t, configRepo, gateway)
}

func createAuthenticationRepository(t *testing.T, apiServer *httptest.Server, authServer *httptest.Server) (configuration.ConfigurationRepository, api.AuthenticationRepository) {
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

	return configRepo, authenticator
}

func testRefreshToken(t *testing.T, configRepo configuration.ConfigurationRepository, gateway Gateway) {
	config, err := configRepo.Get()
	assert.NoError(t, err)

	request, apiResponse := gateway.NewRequest("POST", config.Target+"/v2/foo", config.AccessToken, strings.NewReader("expected body"))
	assert.False(t, apiResponse.IsNotSuccessful())

	apiResponse = gateway.PerformRequest(request)
	assert.False(t, apiResponse.IsNotSuccessful())

	savedConfig := testconfig.SavedConfiguration
	assert.Equal(t, savedConfig.AccessToken, "bearer new-access-token")
	assert.Equal(t, savedConfig.RefreshToken, "new-refresh-token")
}
