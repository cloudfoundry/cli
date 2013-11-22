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

func TestCreateOrgFailsWithUsage(t *testing.T) {
	orgRepo := &testapi.FakeOrgRepository{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	fakeUI := callCreateOrg(t, []string{}, reqFactory, orgRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callCreateOrg(t, []string{"my-org"}, reqFactory, orgRepo)
	assert.False(t, fakeUI.FailedWithUsage)
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
	fakeUI := callCreateOrg(t, []string{"my-org"}, reqFactory, orgRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Creating org")
	assert.Contains(t, fakeUI.Outputs[0], "my-org")
	assert.Contains(t, fakeUI.Outputs[0], "my-user")
	assert.Equal(t, orgRepo.CreateName, "my-org")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestCreateOrgWhenAlreadyExists(t *testing.T) {
	orgRepo := &testapi.FakeOrgRepository{CreateOrgExists: true}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	fakeUI := callCreateOrg(t, []string{"my-org"}, reqFactory, orgRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Creating org")
	assert.Contains(t, fakeUI.Outputs[0], "my-org")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[2], "my-org")
	assert.Contains(t, fakeUI.Outputs[2], "already exists")
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
