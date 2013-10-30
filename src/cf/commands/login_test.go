package commands_test

import (
	"cf"
	. "cf/commands"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	"strconv"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testterm "testhelpers/terminal"
	"testing"
)

type LoginTestContext struct {

	// test-specific setup
	Flags  []string
	Inputs []string
	Config configuration.Configuration

	// fakes created by callLogin
	configRepo   testconfig.FakeConfigRepository
	ui           *testterm.FakeUI
	authRepo     *testapi.FakeAuthenticationRepository
	endpointRepo *testapi.FakeEndpointRepo
	orgRepo      *testapi.FakeOrgRepository
	spaceRepo    *testapi.FakeSpaceRepository
}

// pass defaultBeforeBlock to callLogin instead of nil
func defaultBeforeBlock(*LoginTestContext) {}

func callLogin(t *testing.T, c *LoginTestContext, beforeBlock func(*LoginTestContext)) {

	// setup test fakes
	c.configRepo = testconfig.FakeConfigRepository{}
	c.ui = &testterm.FakeUI{
		Inputs: c.Inputs,
	}
	c.authRepo = &testapi.FakeAuthenticationRepository{
		AccessToken:  "my_access_token",
		RefreshToken: "my_refresh_token",
		ConfigRepo:   c.configRepo,
	}
	c.endpointRepo = &testapi.FakeEndpointRepo{}
	c.orgRepo = &testapi.FakeOrgRepository{
		FindByNameOrganization: cf.Organization{Name: "my-org", Guid: "my-org-guid"},
	}

	c.spaceRepo = &testapi.FakeSpaceRepository{
		FindByNameSpace: cf.Space{Name: "my-space", Guid: "my-space-guid"},
	}

	// initialize config
	c.configRepo.Delete()
	config, _ := c.configRepo.Get()
	config.Target = c.Config.Target
	config.Organization = c.Config.Organization
	config.Space = c.Config.Space

	// run any test-specific setup
	beforeBlock(c)

	// run login command
	l := NewLogin(c.ui, c.configRepo, c.authRepo, c.endpointRepo, c.orgRepo, c.spaceRepo)
	l.Run(testcmd.NewContext("login", c.Flags))
}

func TestSuccessfullyLoggingInWithNumericalPrompts(t *testing.T) {
	OUT_OF_RANGE_CHOICE := "3"
	c := LoginTestContext{
		Inputs: []string{"api.example.com", "user@example.com", "password", OUT_OF_RANGE_CHOICE, "2", OUT_OF_RANGE_CHOICE, "1"},
	}

	callLogin(t, &c, func(c *LoginTestContext) {
		c.orgRepo.Organizations = []cf.Organization{
			{Guid: "some-org-guid", Name: "some-org"},
			{Guid: "my-org-guid", Name: "my-org"},
		}

		c.spaceRepo.Spaces = []cf.Space{
			{Guid: "my-space-guid", Name: "my-space"},
			{Guid: "some-space-guid", Name: "some-space"},
		}
	})

	savedConfig := testconfig.SavedConfiguration

	expectedOutputs := []string{
		"Select an org:",
		"1. some-org",
		"2. my-org",
		"Select a space:",
		"1. my-space",
		"2. some-space",
	}
	testassert.SliceContains(t, c.ui.Outputs, expectedOutputs)

	assert.Equal(t, savedConfig.Target, "api.example.com")
	assert.Equal(t, savedConfig.Organization.Guid, "my-org-guid")
	assert.Equal(t, savedConfig.Space.Guid, "my-space-guid")
	assert.Equal(t, savedConfig.AccessToken, "my_access_token")
	assert.Equal(t, savedConfig.RefreshToken, "my_refresh_token")

	assert.Equal(t, c.endpointRepo.UpdateEndpointEndpoint, "api.example.com")
	assert.Equal(t, c.authRepo.Email, "user@example.com")
	assert.Equal(t, c.authRepo.Password, "password")

	assert.Equal(t, c.orgRepo.FindByNameName, "my-org")
	assert.Equal(t, c.spaceRepo.FindByNameName, "my-space")

	assert.True(t, c.ui.ShowConfigurationCalled)
}

func TestSuccessfullyLoggingInWithStringPrompts(t *testing.T) {
	c := LoginTestContext{
		Inputs: []string{"api.example.com", "user@example.com", "password", "my-org", "my-space"},
	}

	callLogin(t, &c, func(c *LoginTestContext) {
		c.orgRepo.Organizations = []cf.Organization{
			{Guid: "some-org-guid", Name: "some-org"},
			{Guid: "my-org-guid", Name: "my-org"},
		}

		c.spaceRepo.Spaces = []cf.Space{
			{Guid: "my-space-guid", Name: "my-space"},
			{Guid: "some-space-guid", Name: "some-space"},
		}
	})

	savedConfig := testconfig.SavedConfiguration

	expectedOutputs := []string{
		"Select an org:",
		"1. some-org",
		"2. my-org",
		"Select a space:",
		"1. my-space",
		"2. some-space",
	}

	testassert.SliceContains(t, c.ui.Outputs, expectedOutputs)

	assert.Equal(t, savedConfig.Target, "api.example.com")
	assert.Equal(t, savedConfig.Organization.Guid, "my-org-guid")
	assert.Equal(t, savedConfig.Space.Guid, "my-space-guid")
	assert.Equal(t, savedConfig.AccessToken, "my_access_token")
	assert.Equal(t, savedConfig.RefreshToken, "my_refresh_token")

	assert.Equal(t, c.endpointRepo.UpdateEndpointEndpoint, "api.example.com")
	assert.Equal(t, c.authRepo.Email, "user@example.com")
	assert.Equal(t, c.authRepo.Password, "password")

	assert.Equal(t, c.orgRepo.FindByNameName, "my-org")
	assert.Equal(t, c.spaceRepo.FindByNameName, "my-space")

	assert.True(t, c.ui.ShowConfigurationCalled)
}

func TestLoggingInWithTooManyOrgsDoesNotShowOrgList(t *testing.T) {
	c := LoginTestContext{
		Inputs: []string{"api.example.com", "user@example.com", "password", "my-org", "my-space"},
	}

	callLogin(t, &c, func(c *LoginTestContext) {
		for i := 0; i < 50; i++ {
			id := strconv.Itoa(i)
			c.orgRepo.Organizations = append(
				c.orgRepo.Organizations,
				cf.Organization{Guid: "my-org-guid-" + id, Name: "my-org-" + id},
			)
		}
		c.orgRepo.Organizations = append(
			c.orgRepo.Organizations,
			cf.Organization{Guid: "my-org-guid", Name: "my-org"},
		)
		c.spaceRepo.Spaces = []cf.Space{
			{Guid: "my-space-guid", Name: "my-space"},
			{Guid: "some-space-guid", Name: "some-space"},
		}
	})

	savedConfig := testconfig.SavedConfiguration

	assert.True(t, len(c.ui.Outputs) < 50)

	assert.Equal(t, savedConfig.Organization.Guid, "my-org-guid")
}

func TestSuccessfullyLoggingInWithFlags(t *testing.T) {
	c := LoginTestContext{
		Flags: []string{"-a", "api.example.com", "-u", "user@example.com", "-p", "password", "-o", "my-org", "-s", "my-space"},
	}

	callLogin(t, &c, defaultBeforeBlock)

	savedConfig := testconfig.SavedConfiguration

	assert.Equal(t, savedConfig.Target, "api.example.com")
	assert.Equal(t, savedConfig.Organization.Guid, "my-org-guid")
	assert.Equal(t, savedConfig.Space.Guid, "my-space-guid")
	assert.Equal(t, savedConfig.AccessToken, "my_access_token")
	assert.Equal(t, savedConfig.RefreshToken, "my_refresh_token")

	assert.Equal(t, c.endpointRepo.UpdateEndpointEndpoint, "api.example.com")
	assert.Equal(t, c.authRepo.Email, "user@example.com")
	assert.Equal(t, c.authRepo.Password, "password")

	assert.True(t, c.ui.ShowConfigurationCalled)
}

func TestSuccessfullyLoggingInWithEndpointSetInConfig(t *testing.T) {
	existingConfig := configuration.Configuration{
		Target: "http://api.example.com",
	}

	c := LoginTestContext{
		Flags:  []string{"-o", "my-org", "-s", "my-space"},
		Inputs: []string{"user@example.com", "password"},
		Config: existingConfig,
	}

	callLogin(t, &c, defaultBeforeBlock)

	savedConfig := testconfig.SavedConfiguration

	assert.Equal(t, savedConfig.Target, "http://api.example.com")
	assert.Equal(t, savedConfig.Organization.Guid, "my-org-guid")
	assert.Equal(t, savedConfig.Space.Guid, "my-space-guid")
	assert.Equal(t, savedConfig.AccessToken, "my_access_token")
	assert.Equal(t, savedConfig.RefreshToken, "my_refresh_token")

	assert.Equal(t, c.endpointRepo.UpdateEndpointEndpoint, "http://api.example.com")
	assert.Equal(t, c.authRepo.Email, "user@example.com")
	assert.Equal(t, c.authRepo.Password, "password")

	assert.True(t, c.ui.ShowConfigurationCalled)
}

func TestSuccessfullyLoggingInWithOrgSetInConfig(t *testing.T) {
	existingConfig := configuration.Configuration{
		Organization: cf.Organization{Name: "my-org", Guid: "my-org-guid"},
	}

	c := LoginTestContext{
		Flags:  []string{"-s", "my-space"},
		Inputs: []string{"http://api.example.com", "user@example.com", "password"},
		Config: existingConfig,
	}

	callLogin(t, &c, func(c *LoginTestContext) {
		c.orgRepo.FindByNameOrganization = cf.Organization{}
	})

	savedConfig := testconfig.SavedConfiguration

	assert.Equal(t, savedConfig.Target, "http://api.example.com")
	assert.Equal(t, savedConfig.Organization.Guid, "my-org-guid")
	assert.Equal(t, savedConfig.Space.Guid, "my-space-guid")
	assert.Equal(t, savedConfig.AccessToken, "my_access_token")
	assert.Equal(t, savedConfig.RefreshToken, "my_refresh_token")

	assert.Equal(t, c.endpointRepo.UpdateEndpointEndpoint, "http://api.example.com")
	assert.Equal(t, c.authRepo.Email, "user@example.com")
	assert.Equal(t, c.authRepo.Password, "password")

	assert.True(t, c.ui.ShowConfigurationCalled)
}

func TestSuccessfullyLoggingInWithOrgAndSpaceSetInConfig(t *testing.T) {
	existingConfig := configuration.Configuration{
		Organization: cf.Organization{Name: "my-org", Guid: "my-org-guid"},
		Space:        cf.Space{Name: "my-space", Guid: "my-space-guid"},
	}

	c := LoginTestContext{
		Inputs: []string{"http://api.example.com", "user@example.com", "password"},
		Config: existingConfig,
	}

	callLogin(t, &c, func(c *LoginTestContext) {
		c.orgRepo.FindByNameOrganization = cf.Organization{}
		c.spaceRepo.FindByNameInOrgSpace = cf.Space{}
	})

	savedConfig := testconfig.SavedConfiguration

	assert.Equal(t, savedConfig.Target, "http://api.example.com")
	assert.Equal(t, savedConfig.Organization.Guid, "my-org-guid")
	assert.Equal(t, savedConfig.Space.Guid, "my-space-guid")
	assert.Equal(t, savedConfig.AccessToken, "my_access_token")
	assert.Equal(t, savedConfig.RefreshToken, "my_refresh_token")

	assert.Equal(t, c.endpointRepo.UpdateEndpointEndpoint, "http://api.example.com")
	assert.Equal(t, c.authRepo.Email, "user@example.com")
	assert.Equal(t, c.authRepo.Password, "password")

	assert.True(t, c.ui.ShowConfigurationCalled)
}

func TestSuccessfullyLoggingInWithOnlyOneOrg(t *testing.T) {
	c := LoginTestContext{
		Flags:  []string{"-s", "my-space"},
		Inputs: []string{"http://api.example.com", "user@example.com", "password"},
	}

	callLogin(t, &c, func(c *LoginTestContext) {
		c.orgRepo.FindByNameOrganization = cf.Organization{}
		c.orgRepo.Organizations = []cf.Organization{
			{Guid: "my-org-guid", Name: "my-org"},
		}
	})

	savedConfig := testconfig.SavedConfiguration

	assert.Equal(t, savedConfig.Target, "http://api.example.com")
	assert.Equal(t, savedConfig.Organization.Guid, "my-org-guid")
	assert.Equal(t, savedConfig.Space.Guid, "my-space-guid")
	assert.Equal(t, savedConfig.AccessToken, "my_access_token")
	assert.Equal(t, savedConfig.RefreshToken, "my_refresh_token")

	assert.Equal(t, c.endpointRepo.UpdateEndpointEndpoint, "http://api.example.com")
	assert.Equal(t, c.authRepo.Email, "user@example.com")
	assert.Equal(t, c.authRepo.Password, "password")

	assert.True(t, c.ui.ShowConfigurationCalled)
}

func TestSuccessfullyLoggingInWithOnlyOneSpace(t *testing.T) {
	c := LoginTestContext{
		Flags:  []string{"-o", "my-org"},
		Inputs: []string{"http://api.example.com", "user@example.com", "password"},
	}

	callLogin(t, &c, func(c *LoginTestContext) {
		c.spaceRepo.Spaces = []cf.Space{
			{Guid: "my-space-guid", Name: "my-space"},
		}
	})

	savedConfig := testconfig.SavedConfiguration

	assert.Equal(t, savedConfig.Target, "http://api.example.com")
	assert.Equal(t, savedConfig.Organization.Guid, "my-org-guid")
	assert.Equal(t, savedConfig.Space.Guid, "my-space-guid")
	assert.Equal(t, savedConfig.AccessToken, "my_access_token")
	assert.Equal(t, savedConfig.RefreshToken, "my_refresh_token")

	assert.Equal(t, c.endpointRepo.UpdateEndpointEndpoint, "http://api.example.com")
	assert.Equal(t, c.authRepo.Email, "user@example.com")
	assert.Equal(t, c.authRepo.Password, "password")

	assert.True(t, c.ui.ShowConfigurationCalled)
}

func TestUnsuccessfullyLoggingInWithAuthError(t *testing.T) {
	c := LoginTestContext{
		Flags:  []string{"-u", "user@example.com"},
		Inputs: []string{"api.example.com", "password", "password2", "password3"},
	}

	callLogin(t, &c, func(c *LoginTestContext) {
		c.authRepo.AuthError = true
	})

	savedConfig := testconfig.SavedConfiguration

	assert.Equal(t, savedConfig.Target, "api.example.com")
	assert.Empty(t, savedConfig.Organization.Guid)
	assert.Empty(t, savedConfig.Space.Guid)
	assert.Empty(t, savedConfig.AccessToken)
	assert.Empty(t, savedConfig.RefreshToken)

	failIndex := len(c.ui.Outputs) - 2
	assert.Equal(t, c.ui.Outputs[failIndex], "FAILED")
	assert.Equal(t, len(c.ui.PasswordPrompts), 3)
}

func TestUnsuccessfullyLoggingInWithUpdateEndpointError(t *testing.T) {
	c := LoginTestContext{
		Inputs: []string{"api.example.com"},
	}
	callLogin(t, &c, func(c *LoginTestContext) {
		c.endpointRepo.UpdateEndpointError = true
	})

	savedConfig := testconfig.SavedConfiguration

	assert.Empty(t, savedConfig.Target)
	assert.Empty(t, savedConfig.Organization.Guid)
	assert.Empty(t, savedConfig.Space.Guid)
	assert.Empty(t, savedConfig.AccessToken)
	assert.Empty(t, savedConfig.RefreshToken)

	failIndex := len(c.ui.Outputs) - 2
	assert.Equal(t, c.ui.Outputs[failIndex], "FAILED")
}

func TestUnsuccessfullyLoggingInWithOrgFindByNameErr(t *testing.T) {
	c := LoginTestContext{
		Flags:  []string{"-u", "user@example.com", "-o", "my-org", "-s", "my-space"},
		Inputs: []string{"api.example.com", "user@example.com", "password"},
	}

	callLogin(t, &c, func(c *LoginTestContext) {
		c.orgRepo.FindByNameErr = true
	})

	savedConfig := testconfig.SavedConfiguration

	assert.Equal(t, savedConfig.Target, "api.example.com")
	assert.Empty(t, savedConfig.Organization.Guid)
	assert.Empty(t, savedConfig.Space.Guid)
	assert.Equal(t, savedConfig.AccessToken, "my_access_token")
	assert.Equal(t, savedConfig.RefreshToken, "my_refresh_token")

	failIndex := len(c.ui.Outputs) - 2
	assert.Equal(t, c.ui.Outputs[failIndex], "FAILED")
}

func TestUnsuccessfullyLoggingInWithSpaceFindByNameErr(t *testing.T) {
	c := LoginTestContext{
		Flags:  []string{"-u", "user@example.com", "-o", "my-org", "-s", "my-space"},
		Inputs: []string{"api.example.com", "user@example.com", "password"},
	}

	callLogin(t, &c, func(c *LoginTestContext) {
		c.spaceRepo.FindByNameErr = true
	})

	savedConfig := testconfig.SavedConfiguration

	assert.Equal(t, savedConfig.Target, "api.example.com")
	assert.Equal(t, savedConfig.Organization.Guid, "my-org-guid")
	assert.Empty(t, savedConfig.Space.Guid)
	assert.Equal(t, savedConfig.AccessToken, "my_access_token")
	assert.Equal(t, savedConfig.RefreshToken, "my_refresh_token")

	failIndex := len(c.ui.Outputs) - 2
	assert.Equal(t, c.ui.Outputs[failIndex], "FAILED")
}
