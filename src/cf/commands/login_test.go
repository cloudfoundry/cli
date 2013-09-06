package commands_test

import (
	"cf"
	"cf/api"
	. "cf/commands"
	"cf/configuration"
	term "cf/terminal"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestSuccessfullyLoggingIn(t *testing.T) {
	configRepo := testhelpers.FakeConfigRepository{}
	configRepo.Delete()
	config, _ := configRepo.Get()

	ui := new(testhelpers.FakeUI)
	ui.Inputs = []string{"foo@example.com", "bar"}
	auth := &testhelpers.FakeAuthenticator{
		AccessToken: "my_access_token",
		ConfigRepo:  configRepo,
	}
	callLogin(
		[]string{},
		ui,
		config,
		configRepo,
		&testhelpers.FakeOrgRepository{},
		&testhelpers.FakeSpaceRepository{},
		auth,
	)

	savedConfig := testhelpers.SavedConfiguration

	assert.Contains(t, ui.Outputs[0], config.Target)
	assert.Contains(t, ui.Outputs[2], "OK")
	assert.Contains(t, ui.Prompts[0], "Username")
	assert.Contains(t, ui.Prompts[1], "Password")

	assert.Equal(t, savedConfig.AccessToken, "my_access_token")
	assert.Equal(t, auth.Email, "foo@example.com")
	assert.Equal(t, auth.Password, "bar")
}

func TestSuccessfullyLoggingInWithUsernameAsArgument(t *testing.T) {
	configRepo := testhelpers.FakeConfigRepository{}
	configRepo.Delete()
	config, _ := configRepo.Get()

	ui := new(testhelpers.FakeUI)
	ui.Inputs = []string{"bar"}
	auth := &testhelpers.FakeAuthenticator{
		AccessToken: "my_access_token",
		ConfigRepo:  configRepo,
	}
	callLogin(
		[]string{"foo@example.com"},
		ui,
		config,
		configRepo,
		&testhelpers.FakeOrgRepository{},
		&testhelpers.FakeSpaceRepository{},
		auth,
	)

	savedConfig := testhelpers.SavedConfiguration

	assert.Contains(t, ui.Outputs[0], config.Target)
	assert.Contains(t, ui.Outputs[2], "OK")
	assert.Contains(t, ui.Prompts[0], "Password")

	assert.Equal(t, savedConfig.AccessToken, "my_access_token")
	assert.Equal(t, auth.Email, "foo@example.com")
	assert.Equal(t, auth.Password, "bar")
}

func TestLoggingInWithMultipleOrgsAndSpaces(t *testing.T) {
	configRepo := testhelpers.FakeConfigRepository{}
	configRepo.Delete()
	config, _ := configRepo.Get()

	ui := new(testhelpers.FakeUI)
	ui.Inputs = []string{"foo@example.com", "bar", "2", "1"}

	orgs := []cf.Organization{
		cf.Organization{"FirstOrg", "org-1-guid"},
		cf.Organization{"SecondOrg", "org-2-guid"},
	}
	spaces := []cf.Space{
		cf.Space{Name: "FirstSpace", Guid: "space-1-guid"},
		cf.Space{Name: "SecondSpace", Guid: "space-2-guid"},
	}

	callLogin(
		[]string{},
		ui,
		config,
		configRepo,
		&testhelpers.FakeOrgRepository{Organizations: orgs},
		&testhelpers.FakeSpaceRepository{Spaces: spaces},
		&testhelpers.FakeAuthenticator{ConfigRepo: configRepo},
	)

	assert.Contains(t, ui.Outputs[0], config.Target)

	assert.Contains(t, ui.Prompts[0], "Username")
	assert.Contains(t, ui.Prompts[1], "Password")
	assert.Contains(t, ui.Outputs[2], "OK")

	assert.Contains(t, ui.Outputs[3], "FirstOrg")
	assert.Contains(t, ui.Outputs[4], "SecondOrg")

	assert.Contains(t, ui.Prompts[2], "Organization")
	assert.Contains(t, ui.Outputs[5], "SecondOrg")
	assert.Contains(t, ui.Outputs[7], "FirstSpace")
	assert.Contains(t, ui.Outputs[8], "SecondSpace")

	assert.Contains(t, ui.Prompts[3], "Space")
	assert.Contains(t, ui.Outputs[9], "FirstSpace")

	savedConfig := testhelpers.SavedConfiguration
	assert.Equal(t, orgs[1], savedConfig.Organization)
	assert.Equal(t, spaces[0], savedConfig.Space)
}

func TestWhenUserPicksInvalidOrgNumberAndSpaceNumber(t *testing.T) {
	configRepo := testhelpers.FakeConfigRepository{}
	configRepo.Delete()
	config, _ := configRepo.Get()

	orgs := []cf.Organization{
		cf.Organization{"Org1", "org-1-guid"},
		cf.Organization{"Org2", "org-2-guid"},
	}

	spaces := []cf.Space{
		cf.Space{Name: "FirstSpace", Guid: "space-1-guid"},
		cf.Space{Name: "SecondSpace", Guid: "space-2-guid"},
	}

	ui := new(testhelpers.FakeUI)
	ui.Inputs = []string{"foo@example.com", "bar", "3", "2", "3", "1"}

	callLogin(
		[]string{},
		ui,
		config,
		configRepo,
		&testhelpers.FakeOrgRepository{Organizations: orgs},
		&testhelpers.FakeSpaceRepository{Spaces: spaces},
		&testhelpers.FakeAuthenticator{ConfigRepo: configRepo},
	)

	assert.Contains(t, ui.Prompts[2], "Organization")
	assert.Contains(t, ui.Outputs[5], "FAILED")

	assert.Contains(t, ui.Prompts[3], "Organization")
	assert.Contains(t, ui.Outputs[9], "Targeting org")

	assert.Contains(t, ui.Prompts[4], "Space")
	assert.Contains(t, ui.Outputs[13], "FAILED")

	assert.Contains(t, ui.Prompts[5], "Space")
	assert.Contains(t, ui.Outputs[17], "Targeting space")

	savedConfig := testhelpers.SavedConfiguration
	assert.Equal(t, orgs[1], savedConfig.Organization)
	assert.Equal(t, spaces[0], savedConfig.Space)
}

func TestLoggingInWitOneOrgAndOneSpace(t *testing.T) {
	configRepo := testhelpers.FakeConfigRepository{}
	configRepo.Delete()
	config, _ := configRepo.Get()

	ui := new(testhelpers.FakeUI)
	ui.Inputs = []string{"foo@example.com", "bar"}

	orgs := []cf.Organization{
		cf.Organization{"FirstOrg", "org-1-guid"},
	}
	spaces := []cf.Space{
		cf.Space{Name: "FirstSpace", Guid: "space-1-guid"},
	}

	callLogin(
		[]string{},
		ui,
		config,
		configRepo,
		&testhelpers.FakeOrgRepository{Organizations: orgs},
		&testhelpers.FakeSpaceRepository{Spaces: spaces},
		&testhelpers.FakeAuthenticator{ConfigRepo: configRepo},
	)

	assert.Contains(t, ui.Outputs[0], config.Target)

	assert.Contains(t, ui.Prompts[0], "Username")
	assert.Contains(t, ui.Prompts[1], "Password")
	assert.Contains(t, ui.Outputs[2], "OK")
	assert.Contains(t, ui.Outputs[3], "FirstOrg")
	assert.Contains(t, ui.Outputs[4], "OK")

	assert.Contains(t, ui.Outputs[5], "API endpoint:")
	assert.Contains(t, ui.Outputs[7], "FirstOrg")
	assert.Contains(t, ui.Outputs[8], "FirstSpace")

	savedConfig := testhelpers.SavedConfiguration
	assert.Equal(t, orgs[0], savedConfig.Organization)
	assert.Equal(t, spaces[0], savedConfig.Space)
}

func TestLoggingInWithoutOrg(t *testing.T) {
	configRepo := testhelpers.FakeConfigRepository{}
	configRepo.Delete()
	config, _ := configRepo.Get()

	ui := new(testhelpers.FakeUI)
	ui.Inputs = []string{"foo@example.com", "bar"}
	orgs := []cf.Organization{}
	spaces := []cf.Space{}

	callLogin(
		[]string{},
		ui,
		config,
		configRepo,
		&testhelpers.FakeOrgRepository{Organizations: orgs},
		&testhelpers.FakeSpaceRepository{Spaces: spaces},
		&testhelpers.FakeAuthenticator{ConfigRepo: configRepo},
	)

	assert.Contains(t, ui.Outputs[0], config.Target)

	assert.Contains(t, ui.Prompts[0], "Username")
	assert.Contains(t, ui.Prompts[1], "Password")
	assert.Contains(t, ui.Outputs[2], "OK")
	assert.Contains(t, ui.Outputs[3], "No orgs found.")

	savedConfig := testhelpers.SavedConfiguration
	assert.Equal(t, cf.Organization{}, savedConfig.Organization)
	assert.Equal(t, cf.Space{}, savedConfig.Space)
}

func TestLoggingInWithOneOrgAndNoSpace(t *testing.T) {
	configRepo := testhelpers.FakeConfigRepository{}
	configRepo.Delete()
	config, _ := configRepo.Get()

	ui := new(testhelpers.FakeUI)
	ui.Inputs = []string{"foo@example.com", "bar"}
	orgs := []cf.Organization{
		cf.Organization{"FirstOrg", "org-1-guid"},
	}
	spaces := []cf.Space{}

	callLogin(
		[]string{},
		ui,
		config,
		configRepo,
		&testhelpers.FakeOrgRepository{Organizations: orgs},
		&testhelpers.FakeSpaceRepository{Spaces: spaces},
		&testhelpers.FakeAuthenticator{ConfigRepo: configRepo},
	)

	assert.Contains(t, ui.Outputs[0], config.Target)

	assert.Contains(t, ui.Prompts[0], "Username")
	assert.Contains(t, ui.Prompts[1], "Password")
	assert.Contains(t, ui.Outputs[2], "OK")

	assert.Contains(t, ui.Outputs[5], "API endpoint:")
	assert.Contains(t, ui.Outputs[7], "FirstOrg")
	assert.Contains(t, ui.Outputs[8], "No spaces found")

	savedConfig := testhelpers.SavedConfiguration
	assert.Equal(t, orgs[0], savedConfig.Organization)
	assert.Equal(t, cf.Space{}, savedConfig.Space)
}

func TestUnsuccessfullyLoggingIn(t *testing.T) {
	configRepo := testhelpers.FakeConfigRepository{}
	configRepo.Delete()
	config, _ := configRepo.Get()

	ui := new(testhelpers.FakeUI)
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
		config,
		configRepo,
		&testhelpers.FakeOrgRepository{},
		&testhelpers.FakeSpaceRepository{},
		&testhelpers.FakeAuthenticator{AuthError: true, ConfigRepo: configRepo},
	)

	assert.Contains(t, ui.Outputs[0], config.Target)
	assert.Equal(t, ui.Outputs[1], "Authenticating...")
	assert.Equal(t, ui.Outputs[2], "FAILED")
	assert.Equal(t, ui.Outputs[5], "Authenticating...")
	assert.Equal(t, ui.Outputs[6], "FAILED")
	assert.Equal(t, ui.Outputs[9], "Authenticating...")
	assert.Equal(t, ui.Outputs[10], "FAILED")
}

func callLogin(args []string, ui term.UI, config *configuration.Configuration, configRepo configuration.ConfigurationRepository, orgRepo api.OrganizationRepository, spaceRepo api.SpaceRepository, auth api.Authenticator) {
	l := NewLogin(ui, config, configRepo, orgRepo, spaceRepo, auth)
	l.Run(testhelpers.NewContext("login", args))
}
