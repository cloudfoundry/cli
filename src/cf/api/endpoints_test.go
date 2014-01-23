package api_test

import (
	"cf"
	. "cf/api"
	"cf/configuration"
	"cf/net"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
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
  "logging_endpoint": "wss://loggregator.foo.example.org:4443",
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
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	org.Guid = "my-org-guid"

	space := cf.SpaceFields{}
	space.Name = "my-space"
	space.Guid = "my-space-guid"

	config, _ := configRepo.Get()
	config.OrganizationFields = org
	config.SpaceFields = space

	repo.UpdateEndpoint(ts.URL)

	savedConfig := testconfig.SavedConfiguration

	assert.Equal(t, savedConfig.AccessToken, "")
	assert.Equal(t, savedConfig.AuthorizationEndpoint, "https://login.example.com")
	assert.Equal(t, savedConfig.LoggregatorEndPoint, "wss://loggregator.foo.example.org:4443")
	assert.Equal(t, savedConfig.Target, ts.URL)
	assert.Equal(t, savedConfig.ApiVersion, "42.0.0")
	assert.False(t, savedConfig.HasOrganization())
	assert.False(t, savedConfig.HasSpace())
}

func TestUpdateEndpointWhenUrlIsAlreadyTargeted(t *testing.T) {
	configRepo := testconfig.FakeConfigRepository{}
	configRepo.Delete()
	configRepo.Login()

	ts, repo := createEndpointRepoForUpdate(configRepo, validApiInfoEndpoint)
	defer ts.Close()

	org := cf.OrganizationFields{}
	org.Name = "my-org"
	org.Guid = "my-org-guid"

	space := cf.SpaceFields{}
	space.Name = "my-space"
	space.Guid = "my-space-guid"

	config, _ := configRepo.Get()
	config.Target = ts.URL
	config.AccessToken = "some access token"
	config.RefreshToken = "some refresh token"
	config.OrganizationFields = org
	config.SpaceFields = space

	repo.UpdateEndpoint(ts.URL)

	assert.Equal(t, config.OrganizationFields, org)
	assert.Equal(t, config.SpaceFields, space)
	assert.Equal(t, config.AccessToken, "some access token")
	assert.Equal(t, config.RefreshToken, "some refresh token")
}

func TestUpdateEndpointWhenUrlIsMissingSchemeAndHttpsEndpointExists(t *testing.T) {
	configRepo := testconfig.FakeConfigRepository{}
	configRepo.Delete()
	configRepo.Login()

	ts, repo := createEndpointRepoForUpdate(configRepo, validApiInfoEndpoint)
	defer ts.Close()

	schemelessURL := strings.Replace(ts.URL, "https://", "", 1)
	endpoint, apiResponse := repo.UpdateEndpoint(schemelessURL)
	assert.Equal(t, "https://"+schemelessURL, endpoint)

	assert.True(t, apiResponse.IsSuccessful())

	savedConfig := testconfig.SavedConfiguration

	assert.Equal(t, savedConfig.AccessToken, "")
	assert.Equal(t, savedConfig.AuthorizationEndpoint, "https://login.example.com")
	assert.Equal(t, savedConfig.Target, ts.URL)
	assert.Equal(t, savedConfig.ApiVersion, "42.0.0")
}

func TestUpdateEndpointWhenUrlIsMissingSchemeAndHttpEndpointExists(t *testing.T) {
	configRepo := testconfig.FakeConfigRepository{}
	configRepo.Delete()
	configRepo.Login()

	ts, repo := createInsecureEndpointRepoForUpdate(configRepo, validApiInfoEndpoint)
	defer ts.Close()

	schemelessURL := strings.Replace(ts.URL, "http://", "", 1)

	endpoint, apiResponse := repo.UpdateEndpoint(schemelessURL)
	assert.Equal(t, "http://"+schemelessURL, endpoint)

	assert.True(t, apiResponse.IsSuccessful())

	savedConfig := testconfig.SavedConfiguration

	assert.Equal(t, savedConfig.AccessToken, "")
	assert.Equal(t, savedConfig.AuthorizationEndpoint, "https://login.example.com")
	assert.Equal(t, savedConfig.Target, ts.URL)
	assert.Equal(t, savedConfig.ApiVersion, "42.0.0")
}

func TestUpdateEndpointWhenEndpointReturns404(t *testing.T) {
	configRepo := testconfig.FakeConfigRepository{}
	configRepo.Login()

	ts, repo := createEndpointRepoForUpdate(configRepo, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	defer ts.Close()

	_, apiResponse := repo.UpdateEndpoint(ts.URL)

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

	_, apiResponse := repo.UpdateEndpoint(ts.URL)

	assert.True(t, apiResponse.IsNotSuccessful())
}

func createEndpointRepoForUpdate(configRepo testconfig.FakeConfigRepository, endpoint func(w http.ResponseWriter, r *http.Request)) (ts *httptest.Server, repo EndpointRepository) {
	if endpoint != nil {
		ts = httptest.NewTLSServer(http.HandlerFunc(endpoint))
	}
	return ts, makeRepo(configRepo)
}

func createInsecureEndpointRepoForUpdate(configRepo testconfig.FakeConfigRepository, endpoint func(w http.ResponseWriter, r *http.Request)) (ts *httptest.Server, repo EndpointRepository) {
	if endpoint != nil {
		ts = httptest.NewServer(http.HandlerFunc(endpoint))
	}
	return ts, makeRepo(configRepo)
}

func makeRepo(configRepo testconfig.FakeConfigRepository) (repo EndpointRepository) {
	config, _ := configRepo.Get()
	gateway := net.NewCloudControllerGateway()
	return NewEndpointRepository(config, gateway, configRepo)
}

func TestGetCloudControllerEndpoint(t *testing.T) {
	configRepo := testconfig.FakeConfigRepository{}
	config := &configuration.Configuration{
		Target: "http://api.example.com",
	}

	repo := NewEndpointRepository(config, net.NewCloudControllerGateway(), configRepo)

	endpoint, apiResponse := repo.GetCloudControllerEndpoint()

	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, endpoint, "http://api.example.com")
}

func TestGetLoggregatorEndpoint(t *testing.T) {
	config := &configuration.Configuration{
		LoggregatorEndPoint: "wss://loggregator.example.com:4443",
	}

	repo := createEndpointRepoForGet(config)

	endpoint, apiResponse := repo.GetLoggregatorEndpoint()

	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, endpoint, "wss://loggregator.example.com:4443")
}

func TestGetUAAEndpoint(t *testing.T) {
	config := &configuration.Configuration{
		AuthorizationEndpoint: "https://login.example.com",
	}

	repo := createEndpointRepoForGet(config)

	endpoint, apiResponse := repo.GetUAAEndpoint()

	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, endpoint, "https://uaa.example.com")
}

func TestEndpointsReturnAnErrorWhenMissing(t *testing.T) {
	config := &configuration.Configuration{}
	repo := createEndpointRepoForGet(config)

	_, response := repo.GetLoggregatorEndpoint()
	assert.True(t, response.IsNotSuccessful())

	_, response = repo.GetCloudControllerEndpoint()
	assert.True(t, response.IsNotSuccessful())

	_, response = repo.GetUAAEndpoint()
	assert.True(t, response.IsNotSuccessful())
}

func createEndpointRepoForGet(config *configuration.Configuration) (repo EndpointRepository) {
	configRepo := testconfig.FakeConfigRepository{}
	repo = NewEndpointRepository(config, net.NewCloudControllerGateway(), configRepo)
	return
}
