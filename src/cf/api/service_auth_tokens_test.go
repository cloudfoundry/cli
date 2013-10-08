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
	ts, repo := createServiceAuthTokenRepo(createServiceAuthTokenEndpoint)
	defer ts.Close()

	apiResponse := repo.Create(cf.ServiceAuthToken{Label: "a label", Provider: "a provider", Token: "a token"})

	assert.True(t, apiResponse.IsSuccessful())
}

func TestFindAll(t *testing.T) {
	reqStatus := &testhelpers.RequestStatus{}

	findAllServiceAuthTokensEndpoint := testhelpers.CreateEndpoint(
		"GET",
		"/v2/service_auth_tokens",
		testhelpers.EndpointCalledMatcher(reqStatus),
		testhelpers.TestResponse{Status: http.StatusOK, Body: `
{
  "resources": [
    {
      "metadata": {
        "guid": "mysql-core-guid"
      },
      "entity": {
        "label": "mysql",
        "provider": "mysql-core"
      }
    },
    {
      "metadata": {
        "guid": "postgres-core-guid"
      },
      "entity": {
        "label": "postgres",
        "provider": "postgres-core"
      }
    }
  ]
}`})

	ts, repo := createServiceAuthTokenRepo(findAllServiceAuthTokensEndpoint)
	defer ts.Close()

	authTokens, apiResponse := repo.FindAll()
	assert.True(t, reqStatus.Called)
	assert.True(t, apiResponse.IsSuccessful())

	assert.Equal(t, len(authTokens), 2)

	assert.Equal(t, authTokens[0].Label, "mysql")
	assert.Equal(t, authTokens[0].Provider, "mysql-core")
	assert.Equal(t, authTokens[0].Guid, "mysql-core-guid")

	assert.Equal(t, authTokens[1].Label, "postgres")
	assert.Equal(t, authTokens[1].Provider, "postgres-core")
	assert.Equal(t, authTokens[1].Guid, "postgres-core-guid")
}

func createServiceAuthTokenRepo(endpoint http.HandlerFunc) (ts *httptest.Server, repo ServiceAuthTokenRepository) {
	ts = httptest.NewTLSServer(endpoint)

	config := &configuration.Configuration{
		Target:      ts.URL,
		AccessToken: "BEARER my_access_token",
	}
	gateway := net.NewCloudControllerGateway()

	repo = NewCloudControllerServiceAuthTokenRepository(config, gateway)
	return
}

var updateServiceAuthTokenEndpoint = testhelpers.CreateEndpoint(
	"PUT",
	"/v2/service_auth_tokens/my-auth-token-guid",
	testhelpers.RequestBodyMatcher(`{"token":"a value"}`),
	testhelpers.TestResponse{Status: http.StatusCreated},
)

func TestUpdate(t *testing.T) {
	ts := httptest.NewTLSServer(updateServiceAuthTokenEndpoint)
	defer ts.Close()

	config := &configuration.Configuration{
		Target:      ts.URL,
		AccessToken: "BEARER my_access_token",
	}
	gateway := net.NewCloudControllerGateway()

	repo := NewCloudControllerServiceAuthTokenRepository(config, gateway)
	apiResponse := repo.Update(cf.ServiceAuthToken{
		Guid: "my-auth-token-guid",
		Label: "a label",
		Provider: "a provider",
		Value: "a value",
	})

	assert.True(t, apiResponse.IsSuccessful())
}
