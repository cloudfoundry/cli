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
	"testhelpers"
	"testing"
)

func TestTargetWithoutArgument(t *testing.T) {
	orgRepo := &testhelpers.FakeOrgRepository{}
	spaceRepo := &testhelpers.FakeSpaceRepository{}
	configRepo := &testhelpers.FakeConfigRepository{}
	config := configRepo.Login()
	config.Target = "https://api.run.pivotal.io"
	fakeUI := callTarget([]string{}, config, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, fakeUI.Outputs[0], "https://api.run.pivotal.io")
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

func TestTargetWhenUrlIsValidHttpsInfoEndpoint(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(validInfoEndpoint))
	defer ts.Close()

	orgRepo := &testhelpers.FakeOrgRepository{}
	spaceRepo := &testhelpers.FakeSpaceRepository{}
	configRepo := &testhelpers.FakeConfigRepository{}
	configRepo.Delete()
	config := configRepo.Login()
	fakeUI := callTarget([]string{ts.URL}, config, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, fakeUI.Outputs[2], ts.URL)
	assert.Contains(t, fakeUI.Outputs[2], "42.0.0")

	savedConfig := testhelpers.SavedConfiguration

	assert.Equal(t, savedConfig.AccessToken, "")
	assert.Equal(t, savedConfig.AuthorizationEndpoint, "https://login.example.com")
	assert.Equal(t, savedConfig.Target, ts.URL)
	assert.Equal(t, savedConfig.ApiVersion, "42.0.0")
}

func TestTargetWhenUrlIsValidHttpInfoEndpoint(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(validInfoEndpoint))
	defer ts.Close()

	orgRepo := &testhelpers.FakeOrgRepository{}
	spaceRepo := &testhelpers.FakeSpaceRepository{}
	configRepo := &testhelpers.FakeConfigRepository{}
	configRepo.Delete()
	config := configRepo.Login()
	fakeUI := callTarget([]string{ts.URL}, config, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, fakeUI.Outputs[2], "Warning: Insecure http API Endpoint detected. Secure https API Endpoints are recommended.")
	assert.Contains(t, fakeUI.Outputs[3], ts.URL)
	assert.Contains(t, fakeUI.Outputs[3], "42.0.0")

	savedConfig := testhelpers.SavedConfiguration

	assert.Equal(t, savedConfig.AccessToken, "")
	assert.Equal(t, savedConfig.AuthorizationEndpoint, "https://login.example.com")
	assert.Equal(t, savedConfig.Target, ts.URL)
	assert.Equal(t, savedConfig.ApiVersion, "42.0.0")
}

func TestTargetWhenUrlIsMissingScheme(t *testing.T) {
	orgRepo := &testhelpers.FakeOrgRepository{}
	spaceRepo := &testhelpers.FakeSpaceRepository{}
	configRepo := &testhelpers.FakeConfigRepository{}
	config := configRepo.Login()
	fakeUI := callTarget([]string{"example.com"}, config, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Setting target")
	assert.Contains(t, fakeUI.Outputs[1], "FAILED")
	assert.Contains(t, fakeUI.Outputs[2], "API Endpoints should start with https:// or http://")
}

var notFoundEndpoint = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func TestTargetWhenEndpointReturns404(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(notFoundEndpoint))
	defer ts.Close()

	configRepo := &testhelpers.FakeConfigRepository{}
	config := configRepo.Login()
	orgRepo := &testhelpers.FakeOrgRepository{}
	spaceRepo := &testhelpers.FakeSpaceRepository{}
	fakeUI := callTarget([]string{ts.URL}, config, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, fakeUI.Outputs[0], ts.URL)
	assert.Contains(t, fakeUI.Outputs[1], "FAILED")
	assert.Contains(t, fakeUI.Outputs[2], "Server error, status code: 404")
}

var invalidJsonResponseEndpoint = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `Foo`)
}

func TestTargetWhenEndpointReturnsInvalidJson(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(invalidJsonResponseEndpoint))
	defer ts.Close()

	configRepo := &testhelpers.FakeConfigRepository{}
	config := configRepo.Login()
	orgRepo := &testhelpers.FakeOrgRepository{}
	spaceRepo := &testhelpers.FakeSpaceRepository{}
	fakeUI := callTarget([]string{ts.URL}, config, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, fakeUI.Outputs[1], "FAILED")
	assert.Contains(t, fakeUI.Outputs[2], "Invalid JSON response from server")
}

var orgInfoEndpoint = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `Foo`)
}

func TestTargetWithLoggedInUserShowsOrgInfo(t *testing.T) {
	cloudController := httptest.NewTLSServer(http.HandlerFunc(orgInfoEndpoint))
	defer cloudController.Close()

	configRepo := &testhelpers.FakeConfigRepository{}
	config := configRepo.Login()
	config.Target = cloudController.URL
	// Token contents
	/*
		"{\"jti\":\"c41899e5-de15-494d-aab4-8fcec517e005\",\"sub\":\"772dda3f-669f-4276-b2bd-90486abe1f6f\",\"scope\":[\"cloud_controller.read\",\"cloud_controller.write\",\"openid\",\"password.write\"],\"client_id\":\"cf\",\"cid\":\"cf\",\"grant_type\":\"password\",\"user_id\":\"772dda3f-669f-4276-b2bd-90486abe1f6f\",\"user_name\":\"user1@example.com\",\"email\":\"user1@example.com\",\"iat\":1377028356,\"exp\":1377035556,\"iss\":\"https://uaa.arborglen.cf-app.com/oauth/token\",\"aud\":[\"openid\",\"cloud_controller\",\"password\"]}"
	*/
	config.AccessToken = `BEARER eyJhbGciOiJSUzI1NiJ9.eyJqdGkiOiJjNDE4OTllNS1kZTE1LTQ5NGQtYWFiNC04ZmNlYzUxN2UwMDUiLCJzdWIiOiI3NzJkZGEzZi02NjlmLTQyNzYtYjJiZC05MDQ4NmFiZTFmNmYiLCJzY29wZSI6WyJjbG91ZF9jb250cm9sbGVyLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwib3BlbmlkIiwicGFzc3dvcmQud3JpdGUiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImdyYW50X3R5cGUiOiJwYXNzd29yZCIsInVzZXJfaWQiOiI3NzJkZGEzZi02NjlmLTQyNzYtYjJiZC05MDQ4NmFiZTFmNmYiLCJ1c2VyX25hbWUiOiJ1c2VyMUBleGFtcGxlLmNvbSIsImVtYWlsIjoidXNlcjFAZXhhbXBsZS5jb20iLCJpYXQiOjEzNzcwMjgzNTYsImV4cCI6MTM3NzAzNTU1NiwiaXNzIjoiaHR0cHM6Ly91YWEuYXJib3JnbGVuLmNmLWFwcC5jb20vb2F1dGgvdG9rZW4iLCJhdWQiOlsib3BlbmlkIiwiY2xvdWRfY29udHJvbGxlciIsInBhc3N3b3JkIl19.kjFJHi0Qir9kfqi2eyhHy6kdewhicAFu8hrPR1a5AxFvxGB45slKEjuP0_72cM_vEYICgZn3PcUUkHU9wghJO9wjZ6kiIKK1h5f2K9g-Iprv9BbTOWUODu1HoLIvg2TtGsINxcRYy_8LW1RtvQc1b4dBPoopaEH4no-BIzp0E5E`

	orgRepo := &testhelpers.FakeOrgRepository{}
	spaceRepo := &testhelpers.FakeSpaceRepository{}

	ui := callTarget([]string{}, config, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[0], cloudController.URL)
	assert.Contains(t, ui.Outputs[1], "user:")
	assert.Contains(t, ui.Outputs[1], "user1@example.com")
	assert.Contains(t, ui.Outputs[2], "No org targeted. Use 'cf target -o' to target an org.")
}

// End with target argument

// Start test with organization option

func TestTargetOrganizationWhenUserHasAccess(t *testing.T) {
	configRepo := &testhelpers.FakeConfigRepository{}
	config := configRepo.Login()
	orgs := []cf.Organization{
		cf.Organization{Name: "my-organization", Guid: "my-organization-guid"},
	}
	orgRepo := &testhelpers.FakeOrgRepository{
		Organizations:      orgs,
		OrganizationByName: orgs[0],
	}
	spaceRepo := &testhelpers.FakeSpaceRepository{}

	ui := callTarget([]string{"-o", "my-organization"}, config, configRepo, orgRepo, spaceRepo)

	assert.Equal(t, orgRepo.OrganizationName, "my-organization")
	assert.Contains(t, ui.Outputs[2], "org:")
	assert.Contains(t, ui.Outputs[2], "my-organization")

	ui = callTarget([]string{}, config, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[2], "my-organization")
}

func TestTargetOrganizationWhenUserDoesNotHaveAccess(t *testing.T) {
	configRepo := &testhelpers.FakeConfigRepository{}
	configRepo.Delete()
	config := configRepo.Login()
	orgs := []cf.Organization{}
	orgRepo := &testhelpers.FakeOrgRepository{Organizations: orgs, OrganizationByNameErr: true}
	spaceRepo := &testhelpers.FakeSpaceRepository{}

	ui := callTarget([]string{}, config, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[2], "No org targeted.")

	ui = callTarget([]string{"-o", "my-organization"}, config, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[0], "FAILED")

	ui = callTarget([]string{}, config, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[2], "No org targeted.")
}

// End test with organization option

// Start test with space option

func TestTargetSpaceWhenNoOrganizationIsSelected(t *testing.T) {
	configRepo := &testhelpers.FakeConfigRepository{}
	configRepo.Delete()
	config := configRepo.Login()
	orgRepo := &testhelpers.FakeOrgRepository{}
	spaceRepo := &testhelpers.FakeSpaceRepository{}

	ui := callTarget([]string{"-s", "my-space"}, config, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "Organization must be set before targeting space.")

	ui = callTarget([]string{}, config, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[2], "No org targeted.")
}

func TestTargetSpaceWhenUserHasAccess(t *testing.T) {
	configRepo := &testhelpers.FakeConfigRepository{}
	configRepo.Delete()
	config := configRepo.Login()
	config.Organization = cf.Organization{Name: "my-org", Guid: "my-org-guid"}

	orgRepo := &testhelpers.FakeOrgRepository{}
	spaces := []cf.Space{
		cf.Space{Name: "my-space", Guid: "my-space-guid"},
	}
	spaceRepo := &testhelpers.FakeSpaceRepository{Spaces: spaces, SpaceByName: spaces[0]}

	ui := callTarget([]string{"-s", "my-space"}, config, configRepo, orgRepo, spaceRepo)

	assert.Equal(t, spaceRepo.SpaceName, "my-space")
	assert.Contains(t, ui.Outputs[3], "space:")
	assert.Contains(t, ui.Outputs[3], "my-space")

	ui = callTarget([]string{}, config, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[3], "my-space")
}

func TestTargetSpaceWhenUserDoesNotHaveAccess(t *testing.T) {
	configRepo := &testhelpers.FakeConfigRepository{}
	configRepo.Delete()
	config := configRepo.Login()
	config.Organization = cf.Organization{Name: "my-org", Guid: "my-org-guid"}
	orgRepo := &testhelpers.FakeOrgRepository{}
	spaceRepo := &testhelpers.FakeSpaceRepository{SpaceByNameErr: true}

	ui := callTarget([]string{"-s", "my-space"}, config, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "You do not have access to that space.")

	ui = callTarget([]string{}, config, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[3], "No space targeted.")
}

// End test with space option

func callTarget(args []string, config *configuration.Configuration, configRepo configuration.ConfigurationRepository, orgRepo api.OrganizationRepository, spaceRepo api.SpaceRepository) (fakeUI *testhelpers.FakeUI) {
	fakeUI = new(testhelpers.FakeUI)
	target := NewTarget(fakeUI, config, configRepo, orgRepo, spaceRepo)
	target.Run(testhelpers.NewContext("target", args))
	return
}
