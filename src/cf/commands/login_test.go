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

func testSuccessfulLogin(t *testing.T, args []string, inputs []string) (ui *testterm.FakeUI) {
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
	callLogin(
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

func TestSuccessfullyLoggingIn(t *testing.T) {
	ui := testSuccessfulLogin(t, []string{}, []string{"user@example.com", "password"})

	assert.Contains(t, ui.PasswordPrompts[0], "Password")
}

func TestSuccessfullyLoggingInWithUsernameAsArgument(t *testing.T) {
	ui := testSuccessfulLogin(t, []string{"user@example.com"}, []string{"password"})

	assert.Contains(t, ui.PasswordPrompts[0], "Password")
}

func TestSuccessfullyLoggingInWithUsernameAndPasswordAsArguments(t *testing.T) {
	testSuccessfulLogin(t, []string{"user@example.com", "password"}, []string{})
}

func TestUnsuccessfullyLoggingIn(t *testing.T) {
	configRepo := testconfig.FakeConfigRepository{}
	configRepo.Delete()
	config, _ := configRepo.Get()

	ui := new(testterm.FakeUI)
	ui.Inputs = []string{
		"foo@example.com",
		"bar",
		"bar",
		"bar",
		"bar",
	}

	callLogin(
		[]string{},
		ui,
		configRepo,
		&testapi.FakeAuthenticationRepository{AuthError: true, ConfigRepo: configRepo},
	)

	assert.Contains(t, ui.Outputs[0], config.Target)
	assert.Equal(t, ui.Outputs[1], "Authenticating...")
	assert.Equal(t, ui.Outputs[2], "FAILED")
	assert.Equal(t, ui.Outputs[4], "Authenticating...")
	assert.Equal(t, ui.Outputs[5], "FAILED")
	assert.Equal(t, ui.Outputs[7], "Authenticating...")
	assert.Equal(t, ui.Outputs[8], "FAILED")
}

func TestUnsuccessfullyLoggingInWithoutInteractivity(t *testing.T) {
	configRepo := testconfig.FakeConfigRepository{}
	configRepo.Delete()
	config, _ := configRepo.Get()

	ui := new(testterm.FakeUI)

	callLogin(
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

func callLogin(args []string, ui terminal.UI, configRepo configuration.ConfigurationRepository, auth api.AuthenticationRepository) {
	l := NewLogin(ui, configRepo, auth)
	l.Run(testcmd.NewContext("login", args))
}
