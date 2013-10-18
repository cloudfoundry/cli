package commands_test

import (
	"cf/api"
	. "cf/commands"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func testSuccessfulAuthenticate(t *testing.T, args []string) (ui *testterm.FakeUI) {
	configRepo := testconfig.FakeConfigRepository{}
	configRepo.Delete()
	config, _ := configRepo.Get()

	auth := &testapi.FakeAuthenticationRepository{
		AccessToken:  "my_access_token",
		RefreshToken: "my_refresh_token",
		ConfigRepo:   configRepo,
	}
	ui = callAuthenticate(
		args,
		configRepo,
		auth,
	)

	savedConfig := testconfig.SavedConfiguration

	assert.Contains(t, ui.Outputs[0], config.Target)
	assert.Contains(t, ui.Outputs[2], "OK")

	assert.Equal(t, savedConfig.AccessToken, "my_access_token")
	assert.Equal(t, savedConfig.RefreshToken, "my_refresh_token")
	assert.Equal(t, auth.Email, "user@example.com")
	assert.Equal(t, auth.Password, "password")

	return
}

func TestAuthenticateFailsWithUsage(t *testing.T) {
	configRepo := testconfig.FakeConfigRepository{}
	configRepo.Delete()

	auth := &testapi.FakeAuthenticationRepository{
		AccessToken:  "my_access_token",
		RefreshToken: "my_refresh_token",
		ConfigRepo:   configRepo,
	}

	ui := callAuthenticate([]string{}, configRepo, auth)
	assert.True(t, ui.FailedWithUsage)

	ui = callAuthenticate([]string{"my-username"}, configRepo, auth)
	assert.True(t, ui.FailedWithUsage)

	ui = callAuthenticate([]string{"my-username", "my-password"}, configRepo, auth)
	assert.False(t, ui.FailedWithUsage)

}

func TestSuccessfullyAuthenticatingWithUsernameAndPasswordAsArguments(t *testing.T) {
	testSuccessfulAuthenticate(t, []string{"user@example.com", "password"})
}

func TestUnsuccessfullyAuthenticatingWithoutInteractivity(t *testing.T) {
	configRepo := testconfig.FakeConfigRepository{}
	configRepo.Delete()
	config, _ := configRepo.Get()

	ui := callAuthenticate(
		[]string{
			"foo@example.com",
			"bar",
		},
		configRepo,
		&testapi.FakeAuthenticationRepository{AuthError: true, ConfigRepo: configRepo},
	)

	assert.Contains(t, ui.Outputs[0], config.Target)
	assert.Equal(t, ui.Outputs[1], "Authenticating...")
	assert.Equal(t, ui.Outputs[2], "FAILED")
	assert.Contains(t, ui.Outputs[3], "Error authenticating")
	assert.Equal(t, len(ui.Outputs), 4)
}

func callAuthenticate(args []string, configRepo configuration.ConfigurationRepository, auth api.AuthenticationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("auth", args)
	cmd := NewAuthenticate(ui, configRepo, auth)
	testcmd.RunCommand(cmd, ctxt, &testreq.FakeReqFactory{})
	return
}
