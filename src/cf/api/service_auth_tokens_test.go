package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	"testing"
)

var findAllServiceAuthTokensEndpoint, findAllStatus = testapi.CreateCheckableEndpoint(
	"GET",
	"/v2/service_auth_tokens",
	nil,
	testapi.TestResponse{Status: http.StatusOK, Body: `
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

var updateServiceAuthTokenEndpoint, updateStatus = testapi.CreateCheckableEndpoint(
	"PUT",
	"/v2/service_auth_tokens/mysql-core-guid",
	testapi.RequestBodyMatcher(`{"token":"a value"}`),
	testapi.TestResponse{Status: http.StatusCreated},
)

var deleteServiceAuthTokenEndpoint, deleteStatus = testapi.CreateCheckableEndpoint(
	"DELETE",
	"/v2/service_auth_tokens/mysql-core-guid",
	nil,
	testapi.TestResponse{Status: http.StatusOK},
)

var servicesEndpoints = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		findAllServiceAuthTokensEndpoint(writer, request)
	case "PUT":
		updateServiceAuthTokenEndpoint(writer, request)
	case "DELETE":
		deleteServiceAuthTokenEndpoint(writer, request)
	}
})

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

func TestCreate(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"POST",
		"/v2/service_auth_tokens",
		testapi.RequestBodyMatcher(`{"label":"a label","provider":"a provider","token":"a token"}`),
		testapi.TestResponse{Status: http.StatusCreated},
	)

	ts, repo := createServiceAuthTokenRepo(endpoint)
	defer ts.Close()

	apiResponse := repo.Create(cf.ServiceAuthToken{Label: "a label", Provider: "a provider", Token: "a token"})

	assert.True(t, status.Called())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestServiceAuthCreate(t *testing.T) {
	endpoint, status := testapi.CreateCheckableEndpoint(
		"POST",
		"/v2/service_auth_tokens",
		testapi.RequestBodyMatcher(`{"label":"a label","provider":"a provider","token":"a token"}`),
		testapi.TestResponse{Status: http.StatusCreated},
	)

	ts, repo := createServiceAuthTokenRepo(endpoint)
	defer ts.Close()

	apiResponse := repo.Create(cf.ServiceAuthToken{Label: "a label", Provider: "a provider", Token: "a token"})

	assert.True(t, status.Called())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestServiceAuthFindAll(t *testing.T) {
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

func TestServiceAuthFindByName(t *testing.T) {
	findAllStatus.Reset()
	ts, repo := createServiceAuthTokenRepo(findAllServiceAuthTokensEndpoint)
	defer ts.Close()

	finderToken := cf.ServiceAuthToken{
		Label:    "mysql",
		Provider: "mysql-core",
	}

	authToken, apiResponse := repo.FindByName(finderToken.FindByNameKey())
	assert.True(t, findAllStatus.Called())
	assert.True(t, apiResponse.IsSuccessful())

	assert.Equal(t, authToken.FindByNameKey(), finderToken.FindByNameKey())
	assert.Equal(t, authToken.Label, "mysql")
	assert.Equal(t, authToken.Provider, "mysql-core")
	assert.Equal(t, authToken.Guid, "mysql-core-guid")
}

func TestServiceAuthFindByNameWithoutMatch(t *testing.T) {
	findAllStatus.Reset()
	ts, repo := createServiceAuthTokenRepo(findAllServiceAuthTokensEndpoint)
	defer ts.Close()

	noMatchToken := cf.ServiceAuthToken{
		Label:    "no",
		Provider: "match",
	}

	_, apiResponse := repo.FindByName(noMatchToken.FindByNameKey())
	assert.True(t, apiResponse.IsNotFound())
	assert.True(t, findAllStatus.Called())
}

func TestServiceAuthUpdate(t *testing.T) {
	updateStatus.Reset()
	ts, repo := createServiceAuthTokenRepo(servicesEndpoints)
	defer ts.Close()

	apiResponse := repo.Update(cf.ServiceAuthToken{
		Guid:  "mysql-core-guid",
		Token: "a value",
	})

	assert.True(t, updateStatus.Called())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestServiceAuthDelete(t *testing.T) {
	deleteStatus.Reset()

	ts := httptest.NewTLSServer(servicesEndpoints)
	defer ts.Close()

	config := &configuration.Configuration{
		Target:      ts.URL,
		AccessToken: "BEARER my_access_token",
	}
	gateway := net.NewCloudControllerGateway()

	repo := NewCloudControllerServiceAuthTokenRepository(config, gateway)
	apiResponse := repo.Delete(cf.ServiceAuthToken{
		Guid: "mysql-core-guid",
	})

	assert.True(t, deleteStatus.Called())
	assert.True(t, apiResponse.IsSuccessful())
}
