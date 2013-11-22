package space_test

import (
	"cf"
	. "cf/commands/space"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestCreateSpaceFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	spaceRepo := &testapi.FakeSpaceRepository{}

	fakeUI := callCreateSpace(t, []string{}, reqFactory, spaceRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callCreateSpace(t, []string{"my-space"}, reqFactory, spaceRepo)
	assert.False(t, fakeUI.FailedWithUsage)
}

func TestCreateSpaceRequirements(t *testing.T) {
	spaceRepo := &testapi.FakeSpaceRepository{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	callCreateSpace(t, []string{"my-space"}, reqFactory, spaceRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
	callCreateSpace(t, []string{"my-space"}, reqFactory, spaceRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
	callCreateSpace(t, []string{"my-space"}, reqFactory, spaceRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

}

func TestCreateSpace(t *testing.T) {
	spaceRepo := &testapi.FakeSpaceRepository{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	fakeUI := callCreateSpace(t, []string{"my-space"}, reqFactory, spaceRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Creating space")
	assert.Contains(t, fakeUI.Outputs[0], "my-space")
	assert.Contains(t, fakeUI.Outputs[0], "my-org")
	assert.Contains(t, fakeUI.Outputs[0], "my-user")
	assert.Equal(t, spaceRepo.CreateSpaceName, "my-space")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[2], "TIP")
}

func TestCreateSpaceWhenItAlreadyExists(t *testing.T) {
	spaceRepo := &testapi.FakeSpaceRepository{CreateSpaceExists: true}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	fakeUI := callCreateSpace(t, []string{"my-space"}, reqFactory, spaceRepo)

	assert.Equal(t, len(fakeUI.Outputs), 3)
	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[2], "my-space")
	assert.Contains(t, fakeUI.Outputs[2], "already exists")
}

func callCreateSpace(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-space", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	space := cf.SpaceFields{}
	space.Name = "my-space"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewCreateSpace(ui, config, spaceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
