package commands_test

import (
	"cf/api"
	. "cf/commands"
	"cf/configuration"
	"cf/terminal"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testterm "testhelpers/terminal"
	"testing"
)

func testSuccessfulAuthenticate(t *testing.T, args []string, inputs []string) (ui *testterm.FakeUI) {
	configRepo := testconfig.FakeConfigRepository{}
	configRepo.Delete()
	config, _ := configRepo.Get()

	ui = new(testterm.FakeUI)
	ui.Inputs = inputs
	auth := &testapi.FakeAuthenticationRepository{
		AccessToken:  "my_access_token",
		RefreshToken: "my_refresh_token",
		ConfigRepo:   configRepo,
	}
	callAuthenticate(
		args,
		ui,
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

func TestSuccessfullyAuthenticating(t *testing.T) {
	ui := testSuccessfulAuthenticate(t, []string{}, []string{"user@example.com", "password"})

	assert.Contains(t, ui.PasswordPrompts[0], "Password")
}

func TestSuccessfullyAuthenticatingWithUsernameAsArgument(t *testing.T) {
	ui := testSuccessfulAuthenticate(t, []string{"user@example.com"}, []string{"password"})

	assert.Contains(t, ui.PasswordPrompts[0], "Password")
}

func TestSuccessfullyAuthenticatingWithUsernameAndPasswordAsArguments(t *testing.T) {
	testSuccessfulAuthenticate(t, []string{"user@example.com", "password"}, []string{})
}

func TestUnsuccessfullyAuthenticating(t *testing.T) {
	configRepo := testconfig.FakeConfigRepository{}
	configRepo.Delete()
	config, _ := configRepo.Get()

	ui := new(testterm.FakeUI)
	ui.Inputs = []string{
		"foo@example.com",
		"bar",
	}

	callAuthenticate(
		[]string{},
		ui,
		configRepo,
		&testapi.FakeAuthenticationRepository{AuthError: true, ConfigRepo: configRepo},
	)

	assert.Contains(t, ui.Outputs[0], config.Target)
	assert.Equal(t, ui.Outputs[1], "Authenticating...")
	assert.Equal(t, ui.Outputs[2], "FAILED")
}

func TestUnsuccessfullyAuthenticatingWithoutInteractivity(t *testing.T) {
	configRepo := testconfig.FakeConfigRepository{}
	configRepo.Delete()
	config, _ := configRepo.Get()

	ui := new(testterm.FakeUI)

	callAuthenticate(
		[]string{
			"foo@example.com",
			"bar",
		},
		ui,
		configRepo,
		&testapi.FakeAuthenticationRepository{AuthError: true, ConfigRepo: configRepo},
	)

	assert.Contains(t, ui.Outputs[0], config.Target)
	assert.Equal(t, ui.Outputs[1], "Authenticating...")
	assert.Equal(t, ui.Outputs[2], "FAILED")
	assert.Contains(t, ui.Outputs[3], "Error authenticating")
	assert.Equal(t, len(ui.Outputs), 4)
}

func callAuthenticate(args []string, ui terminal.UI, configRepo configuration.ConfigurationRepository, auth api.AuthenticationRepository) {
	l := NewAuthenticate(ui, configRepo, auth)
	l.Run(testcmd.NewContext("authenticate", args))
}
