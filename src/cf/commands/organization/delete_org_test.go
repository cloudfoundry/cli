package organization_test

import (
	"cf"
	. "cf/commands/organization"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestDeleteOrgConfirmingWithY(t *testing.T) {
	org := cf.Organization{Name: "org-to-dellete", Guid: "org-to-delete-guid"}
	orgRepo := &testapi.FakeOrgRepository{FindByNameOrganization: org}

	ui := deleteOrg(t, "y", []string{"org-to-delete"}, orgRepo)

	assert.Contains(t, ui.Prompts[0], "Really delete")

	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Equal(t, orgRepo.FindByNameName, "org-to-delete")
	assert.Equal(t, orgRepo.DeletedOrganization, orgRepo.FindByNameOrganization)
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteOrgConfirmingWithYes(t *testing.T) {
	org := cf.Organization{Name: "org-to-delete", Guid: "org-to-delete-guid"}
	orgRepo := &testapi.FakeOrgRepository{FindByNameOrganization: org}

	ui := deleteOrg(t, "Yes", []string{"org-to-delete"}, orgRepo)

	assert.Contains(t, ui.Prompts[0], "Really delete")

	assert.Contains(t, ui.Outputs[0], "Deleting org")
	assert.Contains(t, ui.Outputs[0], "org-to-delete")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Equal(t, orgRepo.FindByNameName, "org-to-delete")
	assert.Equal(t, orgRepo.DeletedOrganization, orgRepo.FindByNameOrganization)
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteTargetedOrganizationClearsConfig(t *testing.T) {
	configRepo := &testconfig.FakeConfigRepository{}
	config, _ := configRepo.Get()
	config.Organization = cf.Organization{Name: "org-to-delete", Guid: "org-to-delete-guid"}
	config.Space = cf.Space{Name: "space-to-delete"}
	configRepo.Save()

	org := cf.Organization{Name: "org-to-dellete", Guid: "org-to-delete-guid"}
	orgRepo := &testapi.FakeOrgRepository{FindByNameOrganization: org}
	deleteOrg(t, "Yes", []string{"org-to-delete"}, orgRepo)

	updatedConfig, err := configRepo.Get()
	assert.NoError(t, err)

	assert.Equal(t, updatedConfig.Organization, cf.Organization{})
	assert.Equal(t, updatedConfig.Space, cf.Space{})
}

func TestDeleteUntargetedOrganizationDoesNotClearConfig(t *testing.T) {
	org := cf.Organization{Name: "org-to-dellete", Guid: "org-to-delete-guid"}
	orgRepo := &testapi.FakeOrgRepository{FindByNameOrganization: org}

	configRepo := &testconfig.FakeConfigRepository{}
	config, _ := configRepo.Get()
	config.Organization = cf.Organization{Name: "some-other-org", Guid: "some-other-org-guid"}
	config.Space = cf.Space{Name: "some-other-space"}
	configRepo.Save()

	deleteOrg(t, "Yes", []string{"org-to-delete"}, orgRepo)

	updatedConfig, err := configRepo.Get()
	assert.NoError(t, err)

	assert.Equal(t, updatedConfig.Organization.Name, "some-other-org")
	assert.Equal(t, updatedConfig.Space.Name, "some-other-space")
}

func TestDeleteOrgWithForceOption(t *testing.T) {
	org := cf.Organization{Name: "org-to-delete", Guid: "org-to-delete-guid"}
	orgRepo := &testapi.FakeOrgRepository{FindByNameOrganization: org}

	ui := deleteOrg(t, "Yes", []string{"-f", "org-to-delete"}, orgRepo)

	assert.Equal(t, len(ui.Prompts), 0)
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[0], "org-to-delete")
	assert.Equal(t, orgRepo.FindByNameName, "org-to-delete")
	assert.Equal(t, orgRepo.DeletedOrganization, orgRepo.FindByNameOrganization)
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteOrgCommandFailsWithUsage(t *testing.T) {
	orgRepo := &testapi.FakeOrgRepository{}
	ui := deleteOrg(t, "Yes", []string{}, orgRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = deleteOrg(t, "Yes", []string{"org-to-delete"}, orgRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestDeleteOrgWhenOrgDoesNotExist(t *testing.T) {
	orgRepo := &testapi.FakeOrgRepository{FindByNameNotFound: true}
	ui := deleteOrg(t, "y", []string{"org-to-delete"}, orgRepo)

	assert.Equal(t, len(ui.Outputs), 3)
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[0], "org-to-delete")
	assert.Equal(t, orgRepo.FindByNameName, "org-to-delete")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "org-to-delete")
	assert.Contains(t, ui.Outputs[2], "does not exist.")
}

func deleteOrg(t *testing.T, confirmation string, args []string, orgRepo *testapi.FakeOrgRepository) (ui *testterm.FakeUI) {
	reqFactory := &testreq.FakeReqFactory{}
	configRepo := &testconfig.FakeConfigRepository{}

	ui = &testterm.FakeUI{
		Inputs: []string{confirmation},
	}
	ctxt := testcmd.NewContext("delete-org", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewDeleteOrg(ui, config, orgRepo, configRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
