package api

import (
	"cf"
	"cf/net"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	testconfig "testhelpers/configuration"
	"testing"
)

var validApiInfoEndpoint = func(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/v2/info" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	infoResponse := `
{
  "name": "vcap",
  "build": "2222",
  "support": "http://support.cloudfoundry.com",
  "version": 2,
  "description": "Cloud Foundry sponsored by Pivotal",
  "authorization_endpoint": "https://login.example.com",
  "api_version": "42.0.0"
} `
	fmt.Fprintln(w, infoResponse)
}

func TestUpdateEndpointWhenUrlIsValidHttpsInfoEndpoint(t *testing.T) {
	configRepo := testconfig.FakeConfigRepository{}
	configRepo.Delete()
	configRepo.Login()

	ts, repo := createEndpointRepo(configRepo, validApiInfoEndpoint)
	defer ts.Close()

	repo.UpdateEndpoint(ts.URL)

	savedConfig := testconfig.SavedConfiguration

	assert.Equal(t, savedConfig.AccessToken, "")
	assert.Equal(t, savedConfig.AuthorizationEndpoint, "https://login.example.com")
	assert.Equal(t, savedConfig.Target, ts.URL)
	assert.Equal(t, savedConfig.ApiVersion, "42.0.0")
}

func TestUpdateEndpointWhenUrlIsValidHttpInfoEndpoint(t *testing.T) {
	configRepo := testconfig.FakeConfigRepository{}
	configRepo.Delete()
	configRepo.Login()

	ts, repo := createEndpointRepo(configRepo, validApiInfoEndpoint)
	defer ts.Close()

	repo.UpdateEndpoint(ts.URL)

	savedConfig := testconfig.SavedConfiguration

	assert.Equal(t, savedConfig.AccessToken, "")
	assert.Equal(t, savedConfig.AuthorizationEndpoint, "https://login.example.com")
	assert.Equal(t, savedConfig.Target, ts.URL)
	assert.Equal(t, savedConfig.ApiVersion, "42.0.0")
}

func TestUpdateEndpointWhenUrlIsMissingScheme(t *testing.T) {
	configRepo := testconfig.FakeConfigRepository{}
	configRepo.Login()
	_, repo := createEndpointRepo(configRepo, nil)

	apiResponse := repo.UpdateEndpoint("example.com")

	assert.True(t, apiResponse.IsNotSuccessful())
}

var notFoundApiEndpoint = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func TestUpdateEndpointWhenEndpointReturns404(t *testing.T) {
	configRepo := testconfig.FakeConfigRepository{}
	configRepo.Login()

	ts, repo := createEndpointRepo(configRepo, notFoundApiEndpoint)
	defer ts.Close()

	apiResponse := repo.UpdateEndpoint(ts.URL)

	assert.True(t, apiResponse.IsNotSuccessful())
}

var invalidJsonResponseApiEndpoint = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `Foo`)
}

func TestUpdateEndpointWhenEndpointReturnsInvalidJson(t *testing.T) {
	configRepo := testconfig.FakeConfigRepository{}
	configRepo.Login()

	ts, repo := createEndpointRepo(configRepo, invalidJsonResponseApiEndpoint)
	defer ts.Close()

	apiResponse := repo.UpdateEndpoint(ts.URL)

	assert.True(t, apiResponse.IsNotSuccessful())
}

func TestGetEndpointForCloudController(t *testing.T) {
	configRepo := testconfig.FakeConfigRepository{}
	configRepo.Delete()
	configRepo.Login()

	ts, repo := createEndpointRepo(configRepo, validApiInfoEndpoint)
	defer ts.Close()

	endpoint, apiResponse := repo.GetEndpoint(cf.CloudControllerEndpointKey)

	config := testconfig.TestConfigurationSingleton

	assert.True(t, apiResponse.IsSuccessful())
	assert.NotEqual(t, endpoint, "")
	assert.Equal(t, endpoint, config.Target)
}

func createEndpointRepo(configRepo testconfig.FakeConfigRepository, endpoint func(w http.ResponseWriter, r *http.Request)) (ts *httptest.Server, repo EndpointRepository) {
	if endpoint != nil {
		ts = httptest.NewTLSServer(http.HandlerFunc(endpoint))
	}

	config, _ := configRepo.Get()
	gateway := net.NewCloudControllerGateway()
	repo = NewEndpointRepository(config, gateway, configRepo)
	return
}
