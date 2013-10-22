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

func TestDeleteSpaceConfirmingWithY(t *testing.T) {
	ui, spaceRepo := deleteSpace(t, "y", []string{"space-to-delete"})

	assert.Equal(t, spaceRepo.FindByNameName, "space-to-delete")
	assert.Contains(t, ui.Prompts[0], "Really delete")

	assert.Contains(t, ui.Outputs[0], "Deleting space ")
	assert.Contains(t, ui.Outputs[0], "space-to-delete")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Equal(t, spaceRepo.DeletedSpace, spaceRepo.FindByNameSpace)
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteSpaceConfirmingWithYes(t *testing.T) {
	ui, spaceRepo := deleteSpace(t, "Yes", []string{"space-to-delete"})

	assert.Equal(t, spaceRepo.FindByNameName, "space-to-delete")
	assert.Contains(t, ui.Prompts[0], "Really delete")

	assert.Contains(t, ui.Outputs[0], "Deleting space ")
	assert.Contains(t, ui.Outputs[0], "space-to-delete")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Equal(t, spaceRepo.DeletedSpace, spaceRepo.FindByNameSpace)
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteSpaceWithForceOption(t *testing.T) {
	space := cf.Space{Name: "space-to-delete", Guid: "space-to-delete-guid"}
	reqFactory := &testreq.FakeReqFactory{}
	spaceRepo := &testapi.FakeSpaceRepository{FindByNameSpace: space}
	configRepo := &testconfig.FakeConfigRepository{}
	config, _ := configRepo.Get()

	ui := &testterm.FakeUI{}
	ctxt := testcmd.NewContext("delete", []string{"-f", "space-to-delete"})

	cmd := NewDeleteSpace(ui, config, spaceRepo, configRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, spaceRepo.FindByNameName, "space-to-delete")
	assert.Equal(t, len(ui.Prompts), 0)
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[0], "space-to-delete")
	assert.Equal(t, spaceRepo.DeletedSpace, spaceRepo.FindByNameSpace)
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteSpaceWhenSpaceDoesNotExist(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	spaceRepo := &testapi.FakeSpaceRepository{FindByNameNotFound: true}
	configRepo := &testconfig.FakeConfigRepository{}
	config, _ := configRepo.Get()

	ui := &testterm.FakeUI{}
	ctxt := testcmd.NewContext("delete", []string{"-f", "space-to-delete"})

	cmd := NewDeleteSpace(ui, config, spaceRepo, configRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, len(ui.Outputs), 3)
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "space-to-delete")
	assert.Contains(t, ui.Outputs[2], "does not exist.")
}

func TestDeleteSpaceWhenSpaceIsTargeted(t *testing.T) {
	space := cf.Space{Name: "space-to-delete", Guid: "space-to-delete-guid"}
	reqFactory := &testreq.FakeReqFactory{}
	spaceRepo := &testapi.FakeSpaceRepository{FindByNameSpace: space}
	configRepo := &testconfig.FakeConfigRepository{}

	config, _ := configRepo.Get()
	config.Space = space
	configRepo.Save()

	ui := &testterm.FakeUI{}
	ctxt := testcmd.NewContext("delete", []string{"-f", "space-to-delete"})

	cmd := NewDeleteSpace(ui, config, spaceRepo, configRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	config, _ = configRepo.Get()
	assert.Equal(t, config.HasSpace(), false)
}

func TestDeleteSpaceWhenSpaceNotTargeted(t *testing.T) {
	space := cf.Space{Name: "space-to-delete", Guid: "space-to-delete-guid"}
	reqFactory := &testreq.FakeReqFactory{}
	spaceRepo := &testapi.FakeSpaceRepository{FindByNameSpace: space}
	configRepo := &testconfig.FakeConfigRepository{}

	config, _ := configRepo.Get()
	config.Space = cf.Space{Name: "do-not-delete", Guid: "do-not-delete-guid"}
	configRepo.Save()

	ui := &testterm.FakeUI{}
	ctxt := testcmd.NewContext("delete", []string{"-f", "space-to-delete"})

	cmd := NewDeleteSpace(ui, config, spaceRepo, configRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	config, _ = configRepo.Get()
	assert.Equal(t, config.HasSpace(), true)
}

func TestDeleteSpaceCommandWith(t *testing.T) {
	ui, _ := deleteSpace(t, "Yes", []string{})
	assert.True(t, ui.FailedWithUsage)

	ui, _ = deleteSpace(t, "Yes", []string{"space-to-delete"})
	assert.False(t, ui.FailedWithUsage)
}

func TestDeleteSpaceCommandFailsWithUsage(t *testing.T) {
	ui, _ := deleteSpace(t, "Yes", []string{})
	assert.True(t, ui.FailedWithUsage)

	ui, _ = deleteSpace(t, "Yes", []string{"space-to-delete"})
	assert.False(t, ui.FailedWithUsage)
}

func deleteSpace(t *testing.T, confirmation string, args []string) (ui *testterm.FakeUI, spaceRepo *testapi.FakeSpaceRepository) {
	space := cf.Space{Name: "space-to-delete", Guid: "space-to-delete-guid"}
	reqFactory := &testreq.FakeReqFactory{}
	spaceRepo = &testapi.FakeSpaceRepository{FindByNameSpace: space}
	configRepo := &testconfig.FakeConfigRepository{}

	ui = &testterm.FakeUI{
		Inputs: []string{confirmation},
	}
	ctxt := testcmd.NewContext("delete-space", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewDeleteSpace(ui, config, spaceRepo, configRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
