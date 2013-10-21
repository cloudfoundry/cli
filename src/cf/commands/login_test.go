package commands_test

import (
	"cf"
	. "cf/commands"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testterm "testhelpers/terminal"
	"testing"
)

type LoginTestContext struct {
	Flags               []string
	Inputs              []string
	Config              configuration.Configuration
	AuthError           bool
	UpdateEndpointError bool
	OrgFindByNameErr    bool
	SpaceFindByNameErr  bool
}

func callLogin(t *testing.T, context LoginTestContext) (ui *testterm.FakeUI, authRepo *testapi.FakeAuthenticationRepository, endpointRepo *testapi.FakeEndpointRepo) {
	configRepo := testconfig.FakeConfigRepository{}
	configRepo.Delete()

	ui = &testterm.FakeUI{
		Inputs: context.Inputs,
	}
	authRepo = &testapi.FakeAuthenticationRepository{
		AccessToken:  "my_access_token",
		RefreshToken: "my_refresh_token",
		ConfigRepo:   configRepo,
		AuthError:    context.AuthError,
	}
	endpointRepo = &testapi.FakeEndpointRepo{
		UpdateEndpointError: context.UpdateEndpointError,
	}
	orgRepo := &testapi.FakeOrgRepository{
		FindByNameOrganization: cf.Organization{Name: "my-org", Guid: "my-org-guid"},
		FindByNameErr:          context.OrgFindByNameErr,
	}
	spaceRepo := &testapi.FakeSpaceRepository{
		FindByNameSpace: cf.Space{Name: "my-space", Guid: "my-space-guid"},
		FindByNameErr:   context.SpaceFindByNameErr,
	}

	l := NewLogin(ui, configRepo, authRepo, endpointRepo, orgRepo, spaceRepo)
	l.Run(testcmd.NewContext("login", context.Flags))

	return
}

func TestSuccessfullyLoggingInWithPrompts(t *testing.T) {
	context := LoginTestContext{
		Inputs: []string{"api.example.com", "user@example.com", "password", "my-org", "my-space"},
	}

	ui, authRepo, endpointRepo := callLogin(t, context)

	savedConfig := testconfig.SavedConfiguration

	assert.Equal(t, savedConfig.Target, "api.example.com")
	assert.Equal(t, savedConfig.Organization.Guid, "my-org-guid")
	assert.Equal(t, savedConfig.Space.Guid, "my-space-guid")
	assert.Equal(t, savedConfig.AccessToken, "my_access_token")
	assert.Equal(t, savedConfig.RefreshToken, "my_refresh_token")

	assert.Equal(t, endpointRepo.UpdateEndpointEndpoint, "api.example.com")
	assert.Equal(t, authRepo.Email, "user@example.com")
	assert.Equal(t, authRepo.Password, "password")

	assert.True(t, ui.ShowConfigurationCalled)
}

// TODO: make this a matrix test
func TestSuccessfullyLoggingInWithFlags(t *testing.T) {
	context := LoginTestContext{
		Flags: []string{"-a", "api.example.com", "-u", "user@example.com", "-p", "password", "-o", "my-org", "-s", "my-space"},
	}

	ui, authRepo, endpointRepo := callLogin(t, context)

	savedConfig := testconfig.SavedConfiguration

	assert.Equal(t, savedConfig.Target, "api.example.com")
	assert.Equal(t, savedConfig.Organization.Guid, "my-org-guid")
	assert.Equal(t, savedConfig.Space.Guid, "my-space-guid")
	assert.Equal(t, savedConfig.AccessToken, "my_access_token")
	assert.Equal(t, savedConfig.RefreshToken, "my_refresh_token")

	assert.Equal(t, endpointRepo.UpdateEndpointEndpoint, "api.example.com")
	assert.Equal(t, authRepo.Email, "user@example.com")
	assert.Equal(t, authRepo.Password, "password")

	assert.True(t, ui.ShowConfigurationCalled)
}

func TestUnsuccessfullyLoggingInWithAuthError(t *testing.T) {
	context := LoginTestContext{
		Flags:     []string{"-u", "user@example.com"},
		Inputs:    []string{"api.example.com", "password", "password2", "password3"},
		AuthError: true,
	}
	ui, _, _ := callLogin(t, context)

	savedConfig := testconfig.SavedConfiguration

	assert.Equal(t, savedConfig.Target, "api.example.com")
	assert.Empty(t, savedConfig.Organization.Guid)
	assert.Empty(t, savedConfig.Space.Guid)
	assert.Empty(t, savedConfig.AccessToken)
	assert.Empty(t, savedConfig.RefreshToken)

	failIndex := len(ui.Outputs) - 2
	assert.Equal(t, ui.Outputs[failIndex], "FAILED")
	assert.Equal(t, len(ui.PasswordPrompts), 3)
}

func TestUnsuccessfullyLoggingInWithUpdateEndpointError(t *testing.T) {
	context := LoginTestContext{
		Flags:               []string{"-u", "user@example.com"},
		Inputs:              []string{"api.example.com", "user@example.com", "password", "my-org", "my-space"},
		UpdateEndpointError: true,
	}
	ui, _, _ := callLogin(t, context)

	savedConfig := testconfig.SavedConfiguration

	assert.Empty(t, savedConfig.Target)
	assert.Empty(t, savedConfig.Organization.Guid)
	assert.Empty(t, savedConfig.Space.Guid)
	assert.Empty(t, savedConfig.AccessToken)
	assert.Empty(t, savedConfig.RefreshToken)

	failIndex := len(ui.Outputs) - 2
	assert.Equal(t, ui.Outputs[failIndex], "FAILED")
}

func TestUnsuccessfullyLoggingInWithOrgFindByNameErr(t *testing.T) {
	context := LoginTestContext{
		Flags:            []string{"-u", "user@example.com"},
		Inputs:           []string{"api.example.com", "user@example.com", "password", "my-org", "my-space"},
		OrgFindByNameErr: true,
	}
	ui, _, _ := callLogin(t, context)

	savedConfig := testconfig.SavedConfiguration

	assert.Equal(t, savedConfig.Target, "api.example.com")
	assert.Empty(t, savedConfig.Organization.Guid)
	assert.Empty(t, savedConfig.Space.Guid)
	assert.Equal(t, savedConfig.AccessToken, "my_access_token")
	assert.Equal(t, savedConfig.RefreshToken, "my_refresh_token")

	failIndex := len(ui.Outputs) - 2
	assert.Equal(t, ui.Outputs[failIndex], "FAILED")
}

func TestUnsuccessfullyLoggingInWithSpaceFindByNameErr(t *testing.T) {
	context := LoginTestContext{
		Flags:              []string{"-u", "user@example.com"},
		Inputs:             []string{"api.example.com", "user@example.com", "password", "my-org", "my-space"},
		SpaceFindByNameErr: true,
	}
	ui, _, _ := callLogin(t, context)

	savedConfig := testconfig.SavedConfiguration

	assert.Equal(t, savedConfig.Target, "api.example.com")
	assert.Equal(t, savedConfig.Organization.Guid, "my-org-guid")
	assert.Empty(t, savedConfig.Space.Guid)
	assert.Equal(t, savedConfig.AccessToken, "my_access_token")
	assert.Equal(t, savedConfig.RefreshToken, "my_refresh_token")

	failIndex := len(ui.Outputs) - 2
	assert.Equal(t, ui.Outputs[failIndex], "FAILED")
}
