package api_test

import (
	"cf"
	. "cf/api"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
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

var findAllStatus = &testhelpers.RequestStatus{}

var findAllServiceAuthTokensEndpoint = testhelpers.CreateEndpoint(
	"GET",
	"/v2/service_auth_tokens",
	testhelpers.EndpointCalledMatcher(findAllStatus),
	testhelpers.TestResponse{Status: http.StatusOK, Body: `
{
  "resources": [
    {
      "metadata": {
        "guid": "mysql-core-guid"
      },
      "entity": {
        "label": "mysql",
        "provider": "mysql-core",
        "token": "mysql-token-guid"
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
}`},
)

func TestFindAll(t *testing.T) {
	findAllStatus.Reset()
	ts, repo := createServiceAuthTokenRepo(findAllServiceAuthTokensEndpoint)
	defer ts.Close()

	authTokens, apiResponse := repo.FindAll()
	assert.True(t, findAllStatus.Called())
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

var updateServiceAuthTokenEndpoint, updateStatus = testhelpers.CreateCheckableEndpoint(
	"PUT",
	"/v2/service_auth_tokens/mysql-core-guid",
	testhelpers.RequestBodyMatcher(`{"token":"a value"}`),
	testhelpers.TestResponse{Status: http.StatusCreated},
)

var servicesEndpoints = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	if strings.Contains(request.Method, "PUT") {
		updateServiceAuthTokenEndpoint(writer, request)
	} else {
		findAllServiceAuthTokensEndpoint(writer, request)
	}
})

func TestServiceAuthUpdate(t *testing.T) {
	updateStatus.Reset()
	findAllStatus.Reset()

	ts := httptest.NewTLSServer(servicesEndpoints)
	defer ts.Close()

	config := &configuration.Configuration{
		Target:      ts.URL,
		AccessToken: "BEARER my_access_token",
	}
	gateway := net.NewCloudControllerGateway()

	repo := NewCloudControllerServiceAuthTokenRepository(config, gateway)
	apiResponse := repo.Update(cf.ServiceAuthToken{
		Label:    "mysql",
		Provider: "mysql-core",
		Token:    "a value",
	})

	assert.True(t, findAllStatus.Called())
	assert.True(t, updateStatus.Called())
	assert.True(t, apiResponse.IsSuccessful())
}
