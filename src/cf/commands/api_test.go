package commands_test

import (
	. "cf/commands"
	"cf/configuration"
	"cf/net"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testhelpers"
	"testing"
)

func TestApiWithoutArgument(t *testing.T) {
	configRepo := &testhelpers.FakeConfigRepository{}
	config, err := configRepo.Get()
	assert.NoError(t, err)
	config.Target = "https://api.run.pivotal.io"
	config.ApiVersion = "2.0"

	ui := callApi([]string{}, configRepo)

	assert.Equal(t, len(ui.Outputs), 1)
	assert.Contains(t, ui.Outputs[0], "https://api.run.pivotal.io")
	assert.Contains(t, ui.Outputs[0], "2.0")
}

// Targeting an API Endpoint

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
	ts := httptest.NewTLSServer(http.HandlerFunc(validApiInfoEndpoint))
	defer ts.Close()

	configRepo := &testhelpers.FakeConfigRepository{}
	configRepo.Delete()
	configRepo.Login()
	ui := callApi([]string{ts.URL}, configRepo)

	assert.Contains(t, ui.Outputs[2], ts.URL)
	assert.Contains(t, ui.Outputs[2], "42.0.0")

	savedConfig := testhelpers.SavedConfiguration

	assert.Equal(t, savedConfig.AccessToken, "")
	assert.Equal(t, savedConfig.AuthorizationEndpoint, "https://login.example.com")
	assert.Equal(t, savedConfig.Target, ts.URL)
	assert.Equal(t, savedConfig.ApiVersion, "42.0.0")
}

func TestApiWhenUrlIsValidHttpInfoEndpoint(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(validApiInfoEndpoint))
	defer ts.Close()

	configRepo := &testhelpers.FakeConfigRepository{}
	configRepo.Delete()
	configRepo.Login()
	ui := callApi([]string{ts.URL}, configRepo)

	assert.Contains(t, ui.Outputs[2], "Warning: Insecure http API Endpoint detected. Secure https API Endpoints are recommended.")
	assert.Contains(t, ui.Outputs[3], ts.URL)
	assert.Contains(t, ui.Outputs[3], "42.0.0")
	assert.Contains(t, ui.Outputs[4], "Not logged in.")

	savedConfig := testhelpers.SavedConfiguration

	assert.Equal(t, savedConfig.AccessToken, "")
	assert.Equal(t, savedConfig.AuthorizationEndpoint, "https://login.example.com")
	assert.Equal(t, savedConfig.Target, ts.URL)
	assert.Equal(t, savedConfig.ApiVersion, "42.0.0")
}

func TestApiWhenUrlIsMissingScheme(t *testing.T) {
	configRepo := &testhelpers.FakeConfigRepository{}
	configRepo.Login()
	ui := callApi([]string{"example.com"}, configRepo)

	assert.Contains(t, ui.Outputs[0], "Setting api endpoint")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "API Endpoints should start with https:// or http://")
}

var notFoundApiEndpoint = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func TestApiWhenEndpointReturns404(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(notFoundApiEndpoint))
	defer ts.Close()

	configRepo := &testhelpers.FakeConfigRepository{}
	configRepo.Login()
	ui := callApi([]string{ts.URL}, configRepo)

	assert.Contains(t, ui.Outputs[0], ts.URL)
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "Server error, status code: 404")
}

var invalidJsonResponseApiEndpoint = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `Foo`)
}

func TestApiWhenEndpointReturnsInvalidJson(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(invalidJsonResponseApiEndpoint))
	defer ts.Close()

	configRepo := &testhelpers.FakeConfigRepository{}
	configRepo.Login()
	ui := callApi([]string{ts.URL}, configRepo)

	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "Invalid JSON response from server")
}

func callApi(args []string, configRepo configuration.ConfigurationRepository) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	gateway := net.NewCloudControllerGateway(nil)
	cmd := NewApi(ui, gateway, configRepo)
	ctxt := testhelpers.NewContext("api", args)
	reqFactory := &testhelpers.FakeReqFactory{}
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
