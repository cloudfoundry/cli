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

func defaultDeleteSpaceSpace() cf.Space {
	space := cf.Space{}
	space.Name = "space-to-delete"
	space.Guid = "space-to-delete-guid"
	return space
}
func defaultDeleteSpaceReqFactory() *testreq.FakeReqFactory {
	return &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, Space: defaultDeleteSpaceSpace()}
}

func TestDeleteSpaceRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
	deleteSpace(t, []string{"y"}, []string{"my-space"}, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
	deleteSpace(t, []string{"y"}, []string{"my-space"}, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	deleteSpace(t, []string{"y"}, []string{"my-space"}, reqFactory)
	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.SpaceName, "my-space")
}

func TestDeleteSpaceConfirmingWithY(t *testing.T) {
	ui, spaceRepo := deleteSpace(t, []string{"y"}, []string{"space-to-delete"}, defaultDeleteSpaceReqFactory())

	assert.Contains(t, ui.Prompts[0], "Really delete")

	assert.Contains(t, ui.Outputs[0], "Deleting space ")
	assert.Contains(t, ui.Outputs[0], "space-to-delete")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Equal(t, spaceRepo.DeletedSpaceGuid, "space-to-delete-guid")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteSpaceConfirmingWithYes(t *testing.T) {
	ui, spaceRepo := deleteSpace(t, []string{"Yes"}, []string{"space-to-delete"}, defaultDeleteSpaceReqFactory())

	assert.Contains(t, ui.Prompts[0], "Really delete")

	assert.Contains(t, ui.Outputs[0], "Deleting space ")
	assert.Contains(t, ui.Outputs[0], "space-to-delete")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Equal(t, spaceRepo.DeletedSpaceGuid, "space-to-delete-guid")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteSpaceWithForceOption(t *testing.T) {
	ui, spaceRepo := deleteSpace(t, []string{}, []string{"-f", "space-to-delete"}, defaultDeleteSpaceReqFactory())

	assert.Equal(t, len(ui.Prompts), 0)
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[0], "space-to-delete")
	assert.Equal(t, spaceRepo.DeletedSpaceGuid, "space-to-delete-guid")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteSpaceWhenSpaceIsTargeted(t *testing.T) {
	reqFactory := defaultDeleteSpaceReqFactory()
	spaceRepo := &testapi.FakeSpaceRepository{}
	configRepo := &testconfig.FakeConfigRepository{}

	config, _ := configRepo.Get()
	config.SpaceFields = defaultDeleteSpaceSpace().SpaceFields
	configRepo.Save()

	ui := &testterm.FakeUI{}
	ctxt := testcmd.NewContext("delete", []string{"-f", "space-to-delete"})

	cmd := NewDeleteSpace(ui, config, spaceRepo, configRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	config, _ = configRepo.Get()
	assert.Equal(t, config.HasSpace(), false)
}

func TestDeleteSpaceWhenSpaceNotTargeted(t *testing.T) {
	reqFactory := defaultDeleteSpaceReqFactory()
	spaceRepo := &testapi.FakeSpaceRepository{}
	configRepo := &testconfig.FakeConfigRepository{}

	config, _ := configRepo.Get()
	otherSpace := cf.SpaceFields{}
	otherSpace.Name = "do-not-delete"
	otherSpace.Guid = "do-not-delete-guid"
	config.SpaceFields = otherSpace
	configRepo.Save()

	ui := &testterm.FakeUI{}
	ctxt := testcmd.NewContext("delete", []string{"-f", "space-to-delete"})

	cmd := NewDeleteSpace(ui, config, spaceRepo, configRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	config, _ = configRepo.Get()
	assert.Equal(t, config.HasSpace(), true)
}

func TestDeleteSpaceCommandWith(t *testing.T) {
	ui, _ := deleteSpace(t, []string{"Yes"}, []string{}, defaultDeleteSpaceReqFactory())
	assert.True(t, ui.FailedWithUsage)

	ui, _ = deleteSpace(t, []string{"Yes"}, []string{"space-to-delete"}, defaultDeleteSpaceReqFactory())
	assert.False(t, ui.FailedWithUsage)
}

func deleteSpace(t *testing.T, inputs []string, args []string, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI, spaceRepo *testapi.FakeSpaceRepository) {
	spaceRepo = &testapi.FakeSpaceRepository{}
	configRepo := &testconfig.FakeConfigRepository{}

	ui = &testterm.FakeUI{
		Inputs: inputs,
	}
	ctxt := testcmd.NewContext("delete-space", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	space := cf.SpaceFields{}
	space.Name = "my-space"
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewDeleteSpace(ui, config, spaceRepo, configRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
