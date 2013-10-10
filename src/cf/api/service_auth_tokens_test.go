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

func TestCreate(t *testing.T) {
	endpoint, status := testhelpers.CreateCheckableEndpoint(
		"POST",
		"/v2/service_auth_tokens",
		testhelpers.RequestBodyMatcher(`{"label":"a label","provider":"a provider","token":"a token"}`),
		testhelpers.TestResponse{Status: http.StatusCreated},
	)

	ts, repo := createServiceAuthTokenRepo(endpoint)
	defer ts.Close()

	apiResponse := repo.Create(cf.ServiceAuthToken{Label: "a label", Provider: "a provider", Token: "a token"})

	assert.True(t, status.Called())
	assert.True(t, apiResponse.IsSuccessful())
}

var findAllServiceAuthTokensEndpoint, findAllStatus = testhelpers.CreateCheckableEndpoint(
	"GET",
	"/v2/service_auth_tokens",
	nil,
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

var updateServiceAuthTokenEndpoint, updateStatus = testhelpers.CreateCheckableEndpoint(
	"PUT",
	"/v2/service_auth_tokens/mysql-core-guid",
	testhelpers.RequestBodyMatcher(`{"token":"a value"}`),
	testhelpers.TestResponse{Status: http.StatusCreated},
)

func TestServiceAuthUpdate(t *testing.T) {
	servicesEndpoints := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.Contains(request.Method, "PUT") {
			updateServiceAuthTokenEndpoint(writer, request)
		} else {
			findAllServiceAuthTokensEndpoint(writer, request)
		}
	})
	updateStatus.Reset()
	findAllStatus.Reset()

	ts, repo := createServiceAuthTokenRepo(servicesEndpoints)
	defer ts.Close()

	apiResponse := repo.Update(cf.ServiceAuthToken{
		Label:    "mysql",
		Provider: "mysql-core",
		Token:    "a value",
	})

	assert.True(t, findAllStatus.Called())
	assert.True(t, updateStatus.Called())
	assert.True(t, apiResponse.IsSuccessful())
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
