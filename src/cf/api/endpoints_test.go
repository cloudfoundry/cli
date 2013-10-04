package api_test

import (
	. "cf/api"
	"cf/net"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testhelpers"
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

func TestApiWhenUrlIsValidHttpsInfoEndpoint(t *testing.T) {
	configRepo := testhelpers.FakeConfigRepository{}
	configRepo.Delete()
	configRepo.Login()

	ts, repo := createRepo(configRepo, validApiInfoEndpoint)
	defer ts.Close()

	repo.UpdateEndpoint(ts.URL)

	savedConfig := testhelpers.SavedConfiguration

	assert.Equal(t, savedConfig.AccessToken, "")
	assert.Equal(t, savedConfig.AuthorizationEndpoint, "https://login.example.com")
	assert.Equal(t, savedConfig.Target, ts.URL)
	assert.Equal(t, savedConfig.ApiVersion, "42.0.0")
}

func TestApiWhenUrlIsValidHttpInfoEndpoint(t *testing.T) {
	configRepo := testhelpers.FakeConfigRepository{}
	configRepo.Delete()
	configRepo.Login()

	ts := httptest.NewServer(http.HandlerFunc(validApiInfoEndpoint))
	defer ts.Close()

	config, _ := configRepo.Get()
	gateway := net.NewCloudControllerGateway()
	repo := NewEndpointRepository(config, gateway, configRepo)
	repo.UpdateEndpoint(ts.URL)

	savedConfig := testhelpers.SavedConfiguration

	assert.Equal(t, savedConfig.AccessToken, "")
	assert.Equal(t, savedConfig.AuthorizationEndpoint, "https://login.example.com")
	assert.Equal(t, savedConfig.Target, ts.URL)
	assert.Equal(t, savedConfig.ApiVersion, "42.0.0")
}

func TestApiWhenUrlIsMissingScheme(t *testing.T) {
	configRepo := testhelpers.FakeConfigRepository{}
	configRepo.Login()
	config, _ := configRepo.Get()
	gateway := net.NewCloudControllerGateway()
	repo := NewEndpointRepository(config, gateway, configRepo)

	apiStatus := repo.UpdateEndpoint("example.com")

	assert.True(t, apiStatus.NotSuccessful())
}

var notFoundApiEndpoint = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func TestApiWhenEndpointReturns404(t *testing.T) {
	configRepo := testhelpers.FakeConfigRepository{}
	configRepo.Login()

	ts, repo := createRepo(configRepo, notFoundApiEndpoint)
	defer ts.Close()

	apiStatus := repo.UpdateEndpoint(ts.URL)

	assert.True(t, apiStatus.NotSuccessful())
}

var invalidJsonResponseApiEndpoint = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `Foo`)
}

func TestApiWhenEndpointReturnsInvalidJson(t *testing.T) {
	configRepo := testhelpers.FakeConfigRepository{}
	configRepo.Login()

	ts, repo := createRepo(configRepo, invalidJsonResponseApiEndpoint)
	defer ts.Close()

	apiStatus := repo.UpdateEndpoint(ts.URL)

	assert.True(t, apiStatus.NotSuccessful())
}

func createRepo(configRepo testhelpers.FakeConfigRepository, endpoint func(w http.ResponseWriter, r *http.Request)) (ts *httptest.Server, repo EndpointRepository) {
	ts = httptest.NewTLSServer(http.HandlerFunc(endpoint))

	config, _ := configRepo.Get()
	gateway := net.NewCloudControllerGateway()
	repo = NewEndpointRepository(config, gateway, configRepo)
	return
}
