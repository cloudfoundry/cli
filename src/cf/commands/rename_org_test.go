package commands_test

import (
	"cf"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestRenameOrgFailsWithUsage(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{}
	orgRepo := &testhelpers.FakeOrgRepository{}

	fakeUI := callRenameOrg([]string{}, reqFactory, orgRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callRenameOrg([]string{"foo"}, reqFactory, orgRepo)
	assert.True(t, fakeUI.FailedWithUsage)
}

func TestRenameOrgRequirements(t *testing.T) {
	orgRepo := &testhelpers.FakeOrgRepository{}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	callRenameOrg([]string{"my-org", "my-new-org"}, reqFactory, orgRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.OrganizationName, "my-org")
}

func TestRenameOrgRun(t *testing.T) {
	orgRepo := &testhelpers.FakeOrgRepository{}

	org := cf.Organization{Name: "my-org", Guid: "my-org-guid"}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, Organization: org}
	ui := callRenameOrg([]string{"my-org", "my-new-org"}, reqFactory, orgRepo)

	assert.Contains(t, ui.Outputs[0], "Renaming org")
	assert.Equal(t, orgRepo.RenameOrganization, org)
	assert.Equal(t, orgRepo.RenameNewName, "my-new-org")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func callRenameOrg(args []string, reqFactory *testhelpers.FakeReqFactory, orgRepo *testhelpers.FakeOrgRepository) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("rename-org", args)
	cmd := NewRenameOrg(ui, orgRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
