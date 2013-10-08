package api_test

import (
	"cf"
	. "cf/api"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testhelpers"
	"testing"
)

var createServiceAuthTokenWithProviderEndpoint = testhelpers.CreateEndpoint(
	"POST",
	"/v2/service_auth_tokens",
	testhelpers.RequestBodyMatcher(`{"label":"a label","provider":"a provider","token":"a value"}`),
	testhelpers.TestResponse{Status: http.StatusCreated},
)

func TestCreateWithProvider(t *testing.T) {
	authToken := cf.ServiceAuthToken{Label: "a label", Provider: "a provider", Value: "a value"}
	testCreate(t, createServiceAuthTokenWithProviderEndpoint, authToken)
}

var createServiceAuthTokenWithoutProviderEndpoint = testhelpers.CreateEndpoint(
	"POST",
	"/v2/service_auth_tokens",
	testhelpers.RequestBodyMatcher(`{"label":"a label","token":"a value"}`),
	testhelpers.TestResponse{Status: http.StatusCreated},
)

func TestCreateWithoutProvider(t *testing.T) {
	authToken := cf.ServiceAuthToken{Label: "a label", Provider: "", Value: "a value"}
	testCreate(t, createServiceAuthTokenWithoutProviderEndpoint, authToken)
}

func testCreate(t *testing.T, endpoint http.HandlerFunc, authToken cf.ServiceAuthToken) {
	ts := httptest.NewTLSServer(endpoint)
	defer ts.Close()

	config := &configuration.Configuration{
		Target:      ts.URL,
		AccessToken: "BEARER my_access_token",
	}
	gateway := net.NewCloudControllerGateway()

	repo := NewCloudControllerServiceAuthTokenRepository(config, gateway)
	apiResponse := repo.Create(authToken)

	assert.True(t, apiResponse.IsSuccessful())
}
