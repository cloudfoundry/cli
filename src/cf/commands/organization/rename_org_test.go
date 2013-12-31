package organization_test

import (
	"cf"
	"cf/commands/organization"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestRenameOrgFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	orgRepo := &testapi.FakeOrgRepository{}

	ui := callRenameOrg(t, []string{}, reqFactory, orgRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callRenameOrg(t, []string{"foo"}, reqFactory, orgRepo)
	assert.True(t, ui.FailedWithUsage)
}

func TestRenameOrgRequirements(t *testing.T) {
	orgRepo := &testapi.FakeOrgRepository{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	callRenameOrg(t, []string{"my-org", "my-new-org"}, reqFactory, orgRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.OrganizationName, "my-org")
}

func TestRenameOrgRun(t *testing.T) {
	orgRepo := &testapi.FakeOrgRepository{}

	org := cf.Organization{}
	org.Name = "my-org"
	org.Guid = "my-org-guid"
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Organization: org}
	ui := callRenameOrg(t, []string{"my-org", "my-new-org"}, reqFactory, orgRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Renaming org", "my-org", "my-new-org", "my-user"},
		{"OK"},
	})

	assert.Equal(t, orgRepo.RenameOrganizationGuid, "my-org-guid")
	assert.Equal(t, orgRepo.RenameNewName, "my-new-org")
}

func callRenameOrg(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, orgRepo *testapi.FakeOrgRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("rename-org", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	spaceFields := cf.SpaceFields{}
	spaceFields.Name = "my-space"

	orgFields := cf.OrganizationFields{}
	orgFields.Name = "my-org"

	config := &configuration.Configuration{
		SpaceFields:        spaceFields,
		OrganizationFields: orgFields,
		AccessToken:        token,
	}

	cmd := organization.NewRenameOrg(ui, config, orgRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
