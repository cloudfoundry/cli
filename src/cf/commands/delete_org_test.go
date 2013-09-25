package commands_test

import (
	"cf"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestDeleteOrganizationConfirmingWithY(t *testing.T) {
	org := cf.Organization{Name: "org-to-dellete", Guid: "org-to-delete-guid"}
	orgRepo := &testhelpers.FakeOrgRepository{FindByNameOrganization: org}

	ui := deleteOrganization("y", []string{"org-to-delete"}, orgRepo)

	assert.Contains(t, ui.Prompts[0], "Really delete")

	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Equal(t, orgRepo.FindByNameName, "org-to-delete")
	assert.Equal(t, orgRepo.DeletedOrganization, orgRepo.FindByNameOrganization)
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteOrganizationConfirmingWithYes(t *testing.T) {
	org := cf.Organization{Name: "org-to-dellete", Guid: "org-to-delete-guid"}
	orgRepo := &testhelpers.FakeOrgRepository{FindByNameOrganization: org}

	ui := deleteOrganization("Yes", []string{"org-to-delete"}, orgRepo)

	assert.Contains(t, ui.Prompts[0], "Really delete")

	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Equal(t, orgRepo.FindByNameName, "org-to-delete")
	assert.Equal(t, orgRepo.DeletedOrganization, orgRepo.FindByNameOrganization)
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteTargetedOrganizationClearsConfig(t *testing.T) {
	configRepo := &testhelpers.FakeConfigRepository{}
	config, _ := configRepo.Get()
	config.Organization = cf.Organization{Name: "org-to-delete", Guid: "org-to-delete-guid"}
	config.Space = cf.Space{Name: "space-to-delete"}
	configRepo.Save()

	org := cf.Organization{Name: "org-to-dellete", Guid: "org-to-delete-guid"}
	orgRepo := &testhelpers.FakeOrgRepository{FindByNameOrganization: org}
	deleteOrganization("Yes", []string{"org-to-delete"}, orgRepo)

	updatedConfig, err := configRepo.Get()
	assert.NoError(t, err)

	assert.Equal(t, updatedConfig.Organization, cf.Organization{})
	assert.Equal(t, updatedConfig.Space, cf.Space{})
}

func TestDeleteUntargetedOrganizationDoesNotClearConfig(t *testing.T) {
	org := cf.Organization{Name: "org-to-dellete", Guid: "org-to-delete-guid"}
	orgRepo := &testhelpers.FakeOrgRepository{FindByNameOrganization: org}

	configRepo := &testhelpers.FakeConfigRepository{}
	config, _ := configRepo.Get()
	config.Organization = cf.Organization{Name: "some-other-org", Guid: "some-other-org-guid"}
	config.Space = cf.Space{Name: "some-other-space"}
	configRepo.Save()

	deleteOrganization("Yes", []string{"org-to-delete"}, orgRepo)

	updatedConfig, err := configRepo.Get()
	assert.NoError(t, err)

	assert.Equal(t, updatedConfig.Organization.Name, "some-other-org")
	assert.Equal(t, updatedConfig.Space.Name, "some-other-space")
}

func TestDeleteOrganizationWithForceOption(t *testing.T) {
	org := cf.Organization{Name: "org-to-delete", Guid: "org-to-delete-guid"}
	orgRepo := &testhelpers.FakeOrgRepository{FindByNameOrganization: org}

	ui := deleteOrganization("Yes", []string{"-f", "org-to-delete"}, orgRepo)

	assert.Equal(t, len(ui.Prompts), 0)
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[0], "org-to-delete")
	assert.Equal(t, orgRepo.FindByNameName, "org-to-delete")
	assert.Equal(t, orgRepo.DeletedOrganization, orgRepo.FindByNameOrganization)
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteOrganizationCommandFailsWithUsage(t *testing.T) {
	orgRepo := &testhelpers.FakeOrgRepository{}
	ui := deleteOrganization("Yes", []string{}, orgRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = deleteOrganization("Yes", []string{"org-to-delete"}, orgRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestDeleteOrganizationWhenOrgDoesNotExist(t *testing.T) {
	orgRepo := &testhelpers.FakeOrgRepository{DidNotFindOrganizationByName: true}
	ui := deleteOrganization("y", []string{"org-to-delete"}, orgRepo)

	assert.Equal(t, len(ui.Outputs), 3)
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[0], "org-to-delete")
	assert.Equal(t, orgRepo.FindByNameName, "org-to-delete")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "org-to-delete")
	assert.Contains(t, ui.Outputs[2], "was already deleted.")
}

func deleteOrganization(confirmation string, args []string, orgRepo *testhelpers.FakeOrgRepository) (ui *testhelpers.FakeUI) {
	reqFactory := &testhelpers.FakeReqFactory{}
	configRepo := &testhelpers.FakeConfigRepository{}

	ui = &testhelpers.FakeUI{
		Inputs: []string{confirmation},
	}
	ctxt := testhelpers.NewContext("delete-org", args)
	cmd := NewDeleteOrg(ui, orgRepo, configRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
