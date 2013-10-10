package organization_test

import (
	"cf"
	. "cf/commands/organization"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestRenameOrgFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	orgRepo := &testapi.FakeOrgRepository{}

	fakeUI := callRenameOrg([]string{}, reqFactory, orgRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callRenameOrg([]string{"foo"}, reqFactory, orgRepo)
	assert.True(t, fakeUI.FailedWithUsage)
}

func TestRenameOrgRequirements(t *testing.T) {
	orgRepo := &testapi.FakeOrgRepository{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	callRenameOrg([]string{"my-org", "my-new-org"}, reqFactory, orgRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.OrganizationName, "my-org")
}

func TestRenameOrgRun(t *testing.T) {
	orgRepo := &testapi.FakeOrgRepository{}

	org := cf.Organization{Name: "my-org", Guid: "my-org-guid"}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Organization: org}
	ui := callRenameOrg([]string{"my-org", "my-new-org"}, reqFactory, orgRepo)

	assert.Contains(t, ui.Outputs[0], "Renaming org")
	assert.Equal(t, orgRepo.RenameOrganization, org)
	assert.Equal(t, orgRepo.RenameNewName, "my-new-org")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func callRenameOrg(args []string, reqFactory *testreq.FakeReqFactory, orgRepo *testapi.FakeOrgRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("rename-org", args)
	cmd := NewRenameOrg(ui, orgRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
