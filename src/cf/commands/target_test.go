package commands

import (
	"cf/configuration"
	"flag"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testhelpers"
	"testing"
)

func newContext(args []string) *cli.Context {
	flagSet := new(flag.FlagSet)
	flagSet.Parse(args)
	globalSet := new(flag.FlagSet)

	return cli.NewContext(cli.NewApp(), flagSet, globalSet)
}

func TestTargetDefaults(t *testing.T) {
	configuration.Delete()
	context := newContext([]string{})
	fakeUI := new(testhelpers.FakeUI)

	Target(context, fakeUI)

	assert.Contains(t, fakeUI.Outputs[0], "https://api.run.pivotal.io")
}

var validInfoEndpoint = func(w http.ResponseWriter, r *http.Request) {
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

func TestTargetWhenUrlIsValidInfoEndpoint(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(validInfoEndpoint))
	defer ts.Close()

	URL, err := url.Parse(ts.URL)
	assert.NoError(t, err)

	context := newContext([]string{URL.Host})
	fakeUI := new(testhelpers.FakeUI)
	Target(context, fakeUI)

	assert.Contains(t, fakeUI.Outputs[2], "https://"+URL.Host)
	assert.Contains(t, fakeUI.Outputs[2], "42.0.0")

	context = newContext([]string{})
	fakeUI = new(testhelpers.FakeUI)
	Target(context, fakeUI)

	assert.Contains(t, fakeUI.Outputs[0], "https://"+URL.Host)
	assert.Contains(t, fakeUI.Outputs[0], "42.0.0")

	savedConfig, err := configuration.Load()

	assert.NoError(t, err)
	assert.Equal(t, savedConfig.AuthorizationEndpoint, "https://login.example.com")
}

var notFoundEndpoint = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	return
}

func TestTargetWhenEndpointReturns404(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(notFoundEndpoint))
	defer ts.Close()

	URL, err := url.Parse(ts.URL)
	assert.NoError(t, err)

	context := newContext([]string{URL.Host})
	fakeUI := new(testhelpers.FakeUI)
	Target(context, fakeUI)

	assert.Contains(t, fakeUI.Outputs[0], "https://"+URL.Host)
	assert.Contains(t, fakeUI.Outputs[1], "FAILED")
	assert.Contains(t, fakeUI.Outputs[2], "Server error, status code: 404")
}

var invalidJsonResponseEndpoint = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `Foo`)
}

func TestTargetWhenEndpointReturnsInvalidJson(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(invalidJsonResponseEndpoint))
	defer ts.Close()

	URL, err := url.Parse(ts.URL)
	assert.NoError(t, err)

	context := newContext([]string{URL.Host})
	fakeUI := new(testhelpers.FakeUI)
	Target(context, fakeUI)

	assert.Contains(t, fakeUI.Outputs[1], "FAILED")
	assert.Contains(t, fakeUI.Outputs[2], "Invalid JSON response from server")
}

var orgInfoEndpoint = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `Foo`)
}

func TestTargetWithLoggedInUserShowsOrgInfo(t *testing.T) {
	cloudController := httptest.NewTLSServer(http.HandlerFunc(orgInfoEndpoint))
	defer cloudController.Close()

	config, err := configuration.Load()
	assert.NoError(t, err)
	config.Target = cloudController.URL
	// Token contents
	/*
		"{\"jti\":\"c41899e5-de15-494d-aab4-8fcec517e005\",\"sub\":\"772dda3f-669f-4276-b2bd-90486abe1f6f\",\"scope\":[\"cloud_controller.read\",\"cloud_controller.write\",\"openid\",\"password.write\"],\"client_id\":\"cf\",\"cid\":\"cf\",\"grant_type\":\"password\",\"user_id\":\"772dda3f-669f-4276-b2bd-90486abe1f6f\",\"user_name\":\"user1@example.com\",\"email\":\"user1@example.com\",\"iat\":1377028356,\"exp\":1377035556,\"iss\":\"https://uaa.arborglen.cf-app.com/oauth/token\",\"aud\":[\"openid\",\"cloud_controller\",\"password\"]}"
	*/
	config.AccessToken = `BEARER eyJhbGciOiJSUzI1NiJ9.eyJqdGkiOiJjNDE4OTllNS1kZTE1LTQ5NGQtYWFiNC04ZmNlYzUxN2UwMDUiLCJzdWIiOiI3NzJkZGEzZi02NjlmLTQyNzYtYjJiZC05MDQ4NmFiZTFmNmYiLCJzY29wZSI6WyJjbG91ZF9jb250cm9sbGVyLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwib3BlbmlkIiwicGFzc3dvcmQud3JpdGUiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImdyYW50X3R5cGUiOiJwYXNzd29yZCIsInVzZXJfaWQiOiI3NzJkZGEzZi02NjlmLTQyNzYtYjJiZC05MDQ4NmFiZTFmNmYiLCJ1c2VyX25hbWUiOiJ1c2VyMUBleGFtcGxlLmNvbSIsImVtYWlsIjoidXNlcjFAZXhhbXBsZS5jb20iLCJpYXQiOjEzNzcwMjgzNTYsImV4cCI6MTM3NzAzNTU1NiwiaXNzIjoiaHR0cHM6Ly91YWEuYXJib3JnbGVuLmNmLWFwcC5jb20vb2F1dGgvdG9rZW4iLCJhdWQiOlsib3BlbmlkIiwiY2xvdWRfY29udHJvbGxlciIsInBhc3N3b3JkIl19.kjFJHi0Qir9kfqi2eyhHy6kdewhicAFu8hrPR1a5AxFvxGB45slKEjuP0_72cM_vEYICgZn3PcUUkHU9wghJO9wjZ6kiIKK1h5f2K9g-Iprv9BbTOWUODu1HoLIvg2TtGsINxcRYy_8LW1RtvQc1b4dBPoopaEH4no-BIzp0E5E`
	err = config.Save()
	assert.NoError(t, err)

	ui := new(testhelpers.FakeUI)

	context := newContext([]string{})
	Target(context, ui)

	assert.Contains(t, ui.Outputs[0], cloudController.URL)
	assert.Contains(t, ui.Outputs[1], "user:")
	assert.Contains(t, ui.Outputs[1], "user1@example.com")
	assert.Contains(t, ui.Outputs[2], "No org targeted. Use 'cf target -o' to target an org.")
}
