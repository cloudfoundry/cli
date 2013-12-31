package organization_test

import (
	"cf"
	. "cf/commands/organization"
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

func TestCreateOrgFailsWithUsage(t *testing.T) {
	orgRepo := &testapi.FakeOrgRepository{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	ui := callCreateOrg(t, []string{}, reqFactory, orgRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateOrg(t, []string{"my-org"}, reqFactory, orgRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestCreateOrgRequirements(t *testing.T) {
	orgRepo := &testapi.FakeOrgRepository{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	callCreateOrg(t, []string{"my-org"}, reqFactory, orgRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
	callCreateOrg(t, []string{"my-org"}, reqFactory, orgRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestCreateOrg(t *testing.T) {
	orgRepo := &testapi.FakeOrgRepository{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	ui := callCreateOrg(t, []string{"my-org"}, reqFactory, orgRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating org", "my-org", "my-user"},
		{"OK"},
	})
	assert.Equal(t, orgRepo.CreateName, "my-org")
}

func TestCreateOrgWhenAlreadyExists(t *testing.T) {
	orgRepo := &testapi.FakeOrgRepository{CreateOrgExists: true}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	ui := callCreateOrg(t, []string{"my-org"}, reqFactory, orgRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating org", "my-org"},
		{"OK"},
		{"my-org", "already exists"},
	})
}

func callCreateOrg(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, orgRepo *testapi.FakeOrgRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-org", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	space := cf.SpaceFields{}
	space.Name = "my-space"

	organization := cf.OrganizationFields{}
	organization.Name = "my-org"

	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: organization,
		AccessToken:        token,
	}

	cmd := NewCreateOrg(fakeUI, config, orgRepo)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
