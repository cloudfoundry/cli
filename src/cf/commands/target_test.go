package commands_test

import (
	"cf"
	"cf/api"
	. "cf/commands"
	"cf/configuration"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testhelpers"
	"testing"
)

func TestTargetWithoutArgument(t *testing.T) {
	configuration.Delete()

	orgRepo := &testhelpers.FakeOrgRepository{}
	spaceRepo := &testhelpers.FakeSpaceRepository{}

	fakeUI := callTarget([]string{}, orgRepo, spaceRepo)

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

	orgRepo := &testhelpers.FakeOrgRepository{}
	spaceRepo := &testhelpers.FakeSpaceRepository{}

	fakeUI := callTarget([]string{URL.Host}, orgRepo, spaceRepo)

	assert.Contains(t, fakeUI.Outputs[3], "https://"+URL.Host)
	assert.Contains(t, fakeUI.Outputs[3], "42.0.0")

	fakeUI = callTarget([]string{}, orgRepo, spaceRepo)

	assert.Contains(t, fakeUI.Outputs[1], "https://"+URL.Host)
	assert.Contains(t, fakeUI.Outputs[1], "42.0.0")

	savedConfig, err := configuration.Load()

	assert.NoError(t, err)
	assert.Equal(t, savedConfig.AuthorizationEndpoint, "https://login.example.com")
}

var notFoundEndpoint = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func TestTargetWhenEndpointReturns404(t *testing.T) {
	configuration.Delete()
	ts := httptest.NewTLSServer(http.HandlerFunc(notFoundEndpoint))
	defer ts.Close()

	URL, err := url.Parse(ts.URL)
	assert.NoError(t, err)

	orgRepo := &testhelpers.FakeOrgRepository{}
	spaceRepo := &testhelpers.FakeSpaceRepository{}

	fakeUI := callTarget([]string{URL.Host}, orgRepo, spaceRepo)

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

	orgRepo := &testhelpers.FakeOrgRepository{}
	spaceRepo := &testhelpers.FakeSpaceRepository{}

	fakeUI := callTarget([]string{URL.Host}, orgRepo, spaceRepo)

	assert.Contains(t, fakeUI.Outputs[1], "FAILED")
	assert.Contains(t, fakeUI.Outputs[2], "Invalid JSON response from server")
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

	orgRepo := &testhelpers.FakeOrgRepository{}
	spaceRepo := &testhelpers.FakeSpaceRepository{}

	ui := callTarget([]string{}, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[1], cloudController.URL)
	assert.Contains(t, ui.Outputs[2], "user:")
	assert.Contains(t, ui.Outputs[2], "user1@example.com")
	assert.Contains(t, ui.Outputs[3], "No org targeted. Use 'cf target -o' to target an org.")
}

// End with target argument

// Start test with organization option

func TestTargetOrganizationWhenUserHasAccess(t *testing.T) {
	login(t)

	orgs := []cf.Organization{
		cf.Organization{Name: "my-organization", Guid: "my-organization-guid"},
	}
	orgRepo := &testhelpers.FakeOrgRepository{
		Organizations:      orgs,
		OrganizationByName: orgs[0],
	}
	spaceRepo := &testhelpers.FakeSpaceRepository{}

	ui := callTarget([]string{"-o", "my-organization"}, orgRepo, spaceRepo)

	assert.Equal(t, orgRepo.OrganizationName, "my-organization")
	assert.Contains(t, ui.Outputs[3], "org:")
	assert.Contains(t, ui.Outputs[3], "my-organization")

	ui = callTarget([]string{}, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[3], "my-organization")
}

func TestTargetOrganizationWhenUserDoesNotHaveAccess(t *testing.T) {
	login(t)

	orgs := []cf.Organization{}

	orgRepo := &testhelpers.FakeOrgRepository{Organizations: orgs, OrganizationByNameErr: true}
	spaceRepo := &testhelpers.FakeSpaceRepository{}

	ui := callTarget([]string{}, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[3], "No org targeted.")

	ui = callTarget([]string{"-o", "my-organization"}, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[0], "FAILED")

	ui = callTarget([]string{}, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[3], "No org targeted.")
}

// End test with organization option

// Start test with space option

func TestTargetSpaceWhenNoOrganizationIsSelected(t *testing.T) {
	login(t)

	orgRepo := &testhelpers.FakeOrgRepository{}
	spaceRepo := &testhelpers.FakeSpaceRepository{}

	ui := callTarget([]string{"-s", "my-space"}, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "Organization must be set before targeting space.")

	ui = callTarget([]string{}, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[4], "No space targeted.")
}

func TestTargetSpaceWhenUserHasAccess(t *testing.T) {
	login(t)
	setOrganization(t)

	orgRepo := &testhelpers.FakeOrgRepository{}
	spaces := []cf.Space{
		cf.Space{Name: "my-space", Guid: "my-space-guid"},
	}
	spaceRepo := &testhelpers.FakeSpaceRepository{Spaces: spaces, SpaceByName: spaces[0]}

	ui := callTarget([]string{"-s", "my-space"}, orgRepo, spaceRepo)

	assert.Equal(t, spaceRepo.SpaceName, "my-space")
	assert.Contains(t, ui.Outputs[4], "space:")
	assert.Contains(t, ui.Outputs[4], "my-space")

	ui = callTarget([]string{}, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[4], "my-space")
}

func TestTargetSpaceWhenUserDoesNotHaveAccess(t *testing.T) {
	login(t)
	setOrganization(t)

	orgRepo := &testhelpers.FakeOrgRepository{}
	spaceRepo := &testhelpers.FakeSpaceRepository{SpaceByNameErr: true}
	ui := callTarget([]string{}, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[4], "No space targeted.")

	ui = callTarget([]string{"-s", "my-space"}, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "You do not have access to that space.")

	ui = callTarget([]string{}, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[4], "No space targeted.")
}

// End test with space option

func callTarget(args []string, orgRepo api.OrganizationRepository, spaceRepo api.SpaceRepository) (fakeUI *testhelpers.FakeUI) {
	fakeUI = new(testhelpers.FakeUI)
	target := NewTarget(fakeUI, orgRepo, spaceRepo)
	target.Run(testhelpers.NewContext(0, args))
	return
}

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
	config.Organization = cf.Organization{Name: "my-org", Guid: "my-org-guid"}
	err = config.Save()
	assert.NoError(t, err)
}
