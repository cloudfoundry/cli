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

var createServiceAuthTokenEndpoint = testhelpers.CreateEndpoint(
	"POST",
	"/v2/service_auth_tokens",
	testhelpers.RequestBodyMatcher(`{"label":"a label","provider":"a provider","token":"a token"}`),
	testhelpers.TestResponse{Status: http.StatusCreated},
)

func TestCreate(t *testing.T) {
	ts := httptest.NewTLSServer(createServiceAuthTokenEndpoint)
	defer ts.Close()

	config := &configuration.Configuration{
		Target:      ts.URL,
		AccessToken: "BEARER my_access_token",
	}
	gateway := net.NewCloudControllerGateway()

	repo := NewCloudControllerServiceAuthTokenRepository(config, gateway)
	apiResponse := repo.Create(cf.ServiceAuthToken{Label: "a label", Provider: "a provider", Token: "a token"})

	assert.True(t, apiResponse.IsSuccessful())
}
