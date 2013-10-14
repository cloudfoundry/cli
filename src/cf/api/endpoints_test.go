package api

import (
	"cf"
	"cf/configuration"
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

	ts, repo := createEndpointRepoForUpdate(configRepo, validApiInfoEndpoint)
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

	ts, repo := createEndpointRepoForUpdate(configRepo, validApiInfoEndpoint)
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
	_, repo := createEndpointRepoForUpdate(configRepo, nil)

	apiResponse := repo.UpdateEndpoint("example.com")

	assert.True(t, apiResponse.IsNotSuccessful())
}

var notFoundApiEndpoint = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func TestUpdateEndpointWhenEndpointReturns404(t *testing.T) {
	configRepo := testconfig.FakeConfigRepository{}
	configRepo.Login()

	ts, repo := createEndpointRepoForUpdate(configRepo, notFoundApiEndpoint)
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

	ts, repo := createEndpointRepoForUpdate(configRepo, invalidJsonResponseApiEndpoint)
	defer ts.Close()

	apiResponse := repo.UpdateEndpoint(ts.URL)

	assert.True(t, apiResponse.IsNotSuccessful())
}

func createEndpointRepoForUpdate(configRepo testconfig.FakeConfigRepository, endpoint func(w http.ResponseWriter, r *http.Request)) (ts *httptest.Server, repo EndpointRepository) {
	if endpoint != nil {
		ts = httptest.NewTLSServer(http.HandlerFunc(endpoint))
	}

	config, _ := configRepo.Get()
	gateway := net.NewCloudControllerGateway()
	repo = NewEndpointRepository(config, gateway, configRepo)
	return
}

// Tests for GetEndpoint

func TestGetEndpointForCloudController(t *testing.T) {
	configRepo := testconfig.FakeConfigRepository{}
	config := &configuration.Configuration{
		Target: "http://api.example.com",
	}

	repo := NewEndpointRepository(config, net.NewCloudControllerGateway(), configRepo)

	endpoint, apiResponse := repo.GetEndpoint(cf.CloudControllerEndpointKey)

	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, endpoint, "http://api.example.com")
}

func TestGetEndpointForLoggregatorSecure(t *testing.T) {
	config := &configuration.Configuration{
		AuthorizationEndpoint: "https://foo.run.pivotal.io",
	}

	repo := createEndpointRepoForGet(config)

	endpoint, apiResponse := repo.GetEndpoint(cf.LoggregatorEndpointKey)

	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, endpoint, "wss://loggregator.run.pivotal.io:4443")
}

func TestGetEndpointForLoggregatorInsecure(t *testing.T) {
	//This is current behavior, which will probably need to be changed to properly support unsecure websocket connections (SH)
	config := &configuration.Configuration{
		AuthorizationEndpoint: "http://bar.run.pivotal.io",
	}

	repo := createEndpointRepoForGet(config)

	endpoint, apiResponse := repo.GetEndpoint(cf.LoggregatorEndpointKey)

	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, endpoint, "ws://loggregator.run.pivotal.io:4443")
}

func createEndpointRepoForGet(config *configuration.Configuration) (repo EndpointRepository) {
	configRepo := testconfig.FakeConfigRepository{}
	repo = NewEndpointRepository(config, net.NewCloudControllerGateway(), configRepo)
	return
}
