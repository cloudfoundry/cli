package commands_test

import (
	"cf/app"
	. "cf/commands"
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

type FakeAuthorizer struct {
	hasAccess bool
}

func (a FakeAuthorizer) CanAccessOrg(userGuid string, orgName string) bool {
	return a.hasAccess
}

func (a FakeAuthorizer) CanAccessSpace(userGuid string, spaceName string) bool {
	return a.hasAccess
}

func newContext(args []string) *cli.Context {
	app := app.New()
	targetCommand := app.Commands[0]

	flagSet := new(flag.FlagSet)
	for i, _ := range targetCommand.Flags {
		targetCommand.Flags[i].Apply(flagSet)
	}

	flagSet.Parse(args)

	globalSet := new(flag.FlagSet)

	return cli.NewContext(cli.NewApp(), flagSet, globalSet)
}

func TestTargetWithoutArgument(t *testing.T) {
	configuration.Delete()
	context := newContext([]string{})
	fakeUI := new(testhelpers.FakeUI)

	Target(context, fakeUI, FakeAuthorizer{true})

	assert.Contains(t, fakeUI.Outputs[1], "https://api.run.pivotal.io")
}

// With target argument
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
	configuration.Delete()
	ts := httptest.NewTLSServer(http.HandlerFunc(validInfoEndpoint))
	defer ts.Close()

	URL, err := url.Parse(ts.URL)
	assert.NoError(t, err)

	context := newContext([]string{URL.Host})
	fakeUI := new(testhelpers.FakeUI)
	Target(context, fakeUI, FakeAuthorizer{true})

	assert.Contains(t, fakeUI.Outputs[3], "https://"+URL.Host)
	assert.Contains(t, fakeUI.Outputs[3], "42.0.0")

	context = newContext([]string{})
	fakeUI = new(testhelpers.FakeUI)
	Target(context, fakeUI, FakeAuthorizer{true})

	assert.Contains(t, fakeUI.Outputs[1], "https://"+URL.Host)
	assert.Contains(t, fakeUI.Outputs[1], "42.0.0")

	savedConfig, err := configuration.Load()

	assert.NoError(t, err)
	assert.Equal(t, savedConfig.AuthorizationEndpoint, "https://login.example.com")
}

var notFoundEndpoint = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	return
}

func TestTargetWhenEndpointReturns404(t *testing.T) {
	configuration.Delete()
	ts := httptest.NewTLSServer(http.HandlerFunc(notFoundEndpoint))
	defer ts.Close()

	URL, err := url.Parse(ts.URL)
	assert.NoError(t, err)

	context := newContext([]string{URL.Host})
	fakeUI := new(testhelpers.FakeUI)
	Target(context, fakeUI, FakeAuthorizer{true})

	assert.Contains(t, fakeUI.Outputs[0], "https://"+URL.Host)
	assert.Contains(t, fakeUI.Outputs[1], "FAILED")
	assert.Contains(t, fakeUI.Outputs[2], "Server error, status code: 404")
}

var invalidJsonResponseEndpoint = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `Foo`)
}

func TestTargetWhenEndpointReturnsInvalidJson(t *testing.T) {
	configuration.Delete()
	ts := httptest.NewTLSServer(http.HandlerFunc(invalidJsonResponseEndpoint))
	defer ts.Close()

	URL, err := url.Parse(ts.URL)
	assert.NoError(t, err)

	context := newContext([]string{URL.Host})
	fakeUI := new(testhelpers.FakeUI)
	Target(context, fakeUI, FakeAuthorizer{true})

	assert.Contains(t, fakeUI.Outputs[1], "FAILED")
	assert.Contains(t, fakeUI.Outputs[2], "Invalid JSON response from server")
}

func TestTargetWithUnreachableEndpoint(t *testing.T) {
	configuration.Delete()
	URL, err := url.Parse("https://foo")
	assert.NoError(t, err)

	context := newContext([]string{URL.Host})
	fakeUI := new(testhelpers.FakeUI)
	Target(context, fakeUI, FakeAuthorizer{true})

	assert.Equal(t, 3, len(fakeUI.Outputs))
	assert.Contains(t, fakeUI.Outputs[1], "FAILED")
}

var orgInfoEndpoint = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `Foo`)
}

func TestTargetWithLoggedInUserShowsOrgInfo(t *testing.T) {
	configuration.Delete()
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
	Target(context, ui, FakeAuthorizer{true})

	assert.Contains(t, ui.Outputs[1], cloudController.URL)
	assert.Contains(t, ui.Outputs[2], "user:")
	assert.Contains(t, ui.Outputs[2], "user1@example.com")
	assert.Contains(t, ui.Outputs[3], "No org targeted. Use 'cf target -o' to target an org.")
}

// End with target argument

// Start test with organization option

func TestTargetOrganizationWhenUserHasAccess(t *testing.T) {
	login(t)

	ui := new(testhelpers.FakeUI)
	context := newContext([]string{"-o", "my-organization"})
	authorizer := FakeAuthorizer{true}

	Target(context, ui, authorizer)
	assert.Contains(t, ui.Outputs[3], "org:")
	assert.Contains(t, ui.Outputs[3], "my-organization")

	ui = new(testhelpers.FakeUI)
	context = newContext([]string{})

	Target(context, ui, authorizer)
	assert.Contains(t, ui.Outputs[3], "my-organization")
}

func TestTargetOrganizationWhenUserDoesNotHaveAccess(t *testing.T) {
	login(t)

	authorizer := FakeAuthorizer{false}

	ui := new(testhelpers.FakeUI)
	context := newContext([]string{})
	Target(context, ui, authorizer)

	assert.Contains(t, ui.Outputs[3], "No org targeted.")

	ui = new(testhelpers.FakeUI)
	context = newContext([]string{"-o", "my-organization"})
	Target(context, ui, authorizer)

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "You do not have access to that org.")

	ui = new(testhelpers.FakeUI)
	context = newContext([]string{})
	Target(context, ui, authorizer)

	assert.Contains(t, ui.Outputs[3], "No org targeted.")
}

// End test with organization option

// Start test with space option

func TestTargetSpaceWhenNoOrganizationIsSelected(t *testing.T) {
	login(t)

	ui := new(testhelpers.FakeUI)
	context := newContext([]string{"-s", "my-space"})
	authorizer := FakeAuthorizer{true}

	Target(context, ui, authorizer)
	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "Organization must be set before targeting space.")

	ui = new(testhelpers.FakeUI)
	context = newContext([]string{})

	Target(context, ui, authorizer)
	assert.Contains(t, ui.Outputs[4], "No space targeted.")
}

func TestTargetSpaceWhenUserHasAccess(t *testing.T) {
	login(t)
	setOrganization(t)

	ui := new(testhelpers.FakeUI)
	context := newContext([]string{"-s", "my-space"})
	authorizer := FakeAuthorizer{true}

	Target(context, ui, authorizer)
	assert.Contains(t, ui.Outputs[4], "space:")
	assert.Contains(t, ui.Outputs[4], "my-space")

	ui = new(testhelpers.FakeUI)
	context = newContext([]string{})

	Target(context, ui, authorizer)
	assert.Contains(t, ui.Outputs[4], "my-space")
}

func TestTargetSpaceWhenUserDoesNotHaveAccess(t *testing.T) {
	login(t)
	setOrganization(t)

	authorizer := FakeAuthorizer{false}

	ui := new(testhelpers.FakeUI)
	context := newContext([]string{})
	Target(context, ui, authorizer)

	assert.Contains(t, ui.Outputs[4], "No space targeted.")

	ui = new(testhelpers.FakeUI)
	context = newContext([]string{"-s", "my-space"})
	Target(context, ui, authorizer)

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "You do not have access to that space.")

	ui = new(testhelpers.FakeUI)
	context = newContext([]string{})
	Target(context, ui, authorizer)

	assert.Contains(t, ui.Outputs[4], "No space targeted.")
}

// End test with space option

func login(t *testing.T) {
	configuration.Delete()
	config, err := configuration.Load()
	assert.NoError(t, err)

	config.AccessToken = `BEARER eyJhbGciOiJSUzI1NiJ9.eyJqdGkiOiJjNDE4OTllNS1kZTE1LTQ5NGQtYWFiNC04ZmNlYzUxN2UwMDUiLCJzdWIiOiI3NzJkZGEzZi02NjlmLTQyNzYtYjJiZC05MDQ4NmFiZTFmNmYiLCJzY29wZSI6WyJjbG91ZF9jb250cm9sbGVyLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwib3BlbmlkIiwicGFzc3dvcmQud3JpdGUiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImdyYW50X3R5cGUiOiJwYXNzd29yZCIsInVzZXJfaWQiOiI3NzJkZGEzZi02NjlmLTQyNzYtYjJiZC05MDQ4NmFiZTFmNmYiLCJ1c2VyX25hbWUiOiJ1c2VyMUBleGFtcGxlLmNvbSIsImVtYWlsIjoidXNlcjFAZXhhbXBsZS5jb20iLCJpYXQiOjEzNzcwMjgzNTYsImV4cCI6MTM3NzAzNTU1NiwiaXNzIjoiaHR0cHM6Ly91YWEuYXJib3JnbGVuLmNmLWFwcC5jb20vb2F1dGgvdG9rZW4iLCJhdWQiOlsib3BlbmlkIiwiY2xvdWRfY29udHJvbGxlciIsInBhc3N3b3JkIl19.kjFJHi0Qir9kfqi2eyhHy6kdewhicAFu8hrPR1a5AxFvxGB45slKEjuP0_72cM_vEYICgZn3PcUUkHU9wghJO9wjZ6kiIKK1h5f2K9g-Iprv9BbTOWUODu1HoLIvg2TtGsINxcRYy_8LW1RtvQc1b4dBPoopaEH4no-BIzp0E5E`
	err = config.Save()
	assert.NoError(t, err)

	return
}

func setOrganization(t *testing.T) {
	config, err := configuration.Load()
	assert.NoError(t, err)
	config.Organization = "my-org"
	err = config.Save()
	assert.NoError(t, err)
}
