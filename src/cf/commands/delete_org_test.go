package commands_test

import (
	"cf"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestDeleteOrganizationConfirmingWithY(t *testing.T) {
	ui, reqFactory, orgRepo := deleteOrganization("y", []string{"org-to-delete"})

	assert.Equal(t, reqFactory.OrganizationName, "org-to-delete")
	assert.Contains(t, ui.Prompts[0], "Really delete")

	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Equal(t, orgRepo.DeletedOrganization, reqFactory.Organization)
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteOrganizationConfirmingWithYes(t *testing.T) {
	ui, reqFactory, orgRepo := deleteOrganization("Yes", []string{"org-to-delete"})

	assert.Equal(t, reqFactory.OrganizationName, "org-to-delete")
	assert.Contains(t, ui.Prompts[0], "Really delete")

	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Equal(t, orgRepo.DeletedOrganization, reqFactory.Organization)
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteOrganizationWithForceOption(t *testing.T) {
	org := cf.Organization{Name: "org-to-delete", Guid: "org-to-delete-guid"}
	reqFactory := &testhelpers.FakeReqFactory{Organization: org}
	orgRepo := &testhelpers.FakeOrgRepository{}

	ui := &testhelpers.FakeUI{}
	ctxt := testhelpers.NewContext("delete", []string{"-f", "org-to-delete"})

	cmd := NewDeleteOrg(ui, orgRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, reqFactory.OrganizationName, "org-to-delete")
	assert.Equal(t, len(ui.Prompts), 0)
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[0], "org-to-delete")
	assert.Equal(t, orgRepo.DeletedOrganization, reqFactory.Organization)
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteOrganizationCommandFailsWithUsage(t *testing.T) {
	ui, _, _ := deleteOrganization("Yes", []string{})
	assert.True(t, ui.FailedWithUsage)

	ui, _, _ = deleteOrganization("Yes", []string{"org-to-delete"})
	assert.False(t, ui.FailedWithUsage)
}

func deleteOrganization(confirmation string, args []string) (ui *testhelpers.FakeUI, reqFactory *testhelpers.FakeReqFactory, orgRepo *testhelpers.FakeOrgRepository) {
	org := cf.Organization{Name: "org-to-dellete", Guid: "org-to-delete-guid"}
	reqFactory = &testhelpers.FakeReqFactory{Organization: org}
	orgRepo = &testhelpers.FakeOrgRepository{}
	ui = &testhelpers.FakeUI{
		Inputs: []string{confirmation},
	}
	ctxt := testhelpers.NewContext("delete-org", args)
	cmd := NewDeleteOrg(ui, orgRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
