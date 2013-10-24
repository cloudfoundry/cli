package buildpack_test

import (
	"cf"
	. "cf/commands/buildpack"
	"cf/net"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestDeleteBuildpackGetRequirements(t *testing.T) {
	ui := &testterm.FakeUI{Inputs: []string{"y"}}
	buildpackRepo := &testapi.FakeBuildpackRepository{}
	cmd := NewDeleteBuildpack(ui, buildpackRepo)

	ctxt := testcmd.NewContext("delete-buildpack", []string{"my-buildpack"})

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestDeleteBuildpackSuccess(t *testing.T) {
	ui := &testterm.FakeUI{Inputs: []string{"y"}}
	buildpackRepo := &testapi.FakeBuildpackRepository{
		FindByNameBuildpack: cf.Buildpack{
			Name: "my-buildpack",
		},
	}
	cmd := NewDeleteBuildpack(ui, buildpackRepo)

	ctxt := testcmd.NewContext("delete-buildpack", []string{"my-buildpack"})
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, buildpackRepo.DeleteBuildpack.Name, "my-buildpack")

	assert.Contains(t, ui.Prompts[0], "delete")
	assert.Contains(t, ui.Prompts[0], "my-buildpack")

	assert.Contains(t, ui.Outputs[0], "Deleting buildpack")
	assert.Contains(t, ui.Outputs[0], "my-buildpack")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteBuildpackNoConfirmation(t *testing.T) {
	ui := &testterm.FakeUI{Inputs: []string{"no"}}
	buildpackRepo := &testapi.FakeBuildpackRepository{
		FindByNameBuildpack: cf.Buildpack{
			Name: "my-buildpack",
		},
	}
	cmd := NewDeleteBuildpack(ui, buildpackRepo)

	ctxt := testcmd.NewContext("delete-buildpack", []string{"my-buildpack"})
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, buildpackRepo.DeleteBuildpack.Name, "")

	assert.Contains(t, ui.Prompts[0], "delete")
	assert.Contains(t, ui.Prompts[0], "my-buildpack")
}

func TestDeleteBuildpackThatDoesNotExist(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	buildpackRepo := &testapi.FakeBuildpackRepository{
		FindByNameNotFound:  true,
		FindByNameBuildpack: cf.Buildpack{Name: "my-buildpack"},
	}

	ui := &testterm.FakeUI{}
	ctxt := testcmd.NewContext("delete-buildpack", []string{"-f", "my-buildpack"})

	cmd := NewDeleteBuildpack(ui, buildpackRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, buildpackRepo.FindByNameName, "my-buildpack")
	assert.True(t, buildpackRepo.FindByNameNotFound)
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[0], "my-buildpack")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "my-buildpack")
	assert.Contains(t, ui.Outputs[2], "does not exist")
}

func TestDeleteBuildpackDeleteError(t *testing.T) {
	ui := &testterm.FakeUI{Inputs: []string{"y"}}
	buildpackRepo := &testapi.FakeBuildpackRepository{
		FindByNameBuildpack: cf.Buildpack{Name: "my-buildpack"},
		DeleteApiResponse:   net.NewApiResponseWithMessage("failed badly"),
	}

	cmd := NewDeleteBuildpack(ui, buildpackRepo)

	ctxt := testcmd.NewContext("delete-buildpack", []string{"my-buildpack"})
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, buildpackRepo.DeleteBuildpack.Name, "my-buildpack")

	assert.Contains(t, ui.Outputs[0], "Deleting buildpack")
	assert.Contains(t, ui.Outputs[0], "my-buildpack")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "my-buildpack")
	assert.Contains(t, ui.Outputs[2], "failed badly")
}

func TestDeleteBuildpackForceFlagSkipsConfirmation(t *testing.T) {
	ui := &testterm.FakeUI{}
	buildpackRepo := &testapi.FakeBuildpackRepository{
		FindByNameBuildpack: cf.Buildpack{Name: "my-buildpack"},
	}

	cmd := NewDeleteBuildpack(ui, buildpackRepo)

	ctxt := testcmd.NewContext("delete-buildpack", []string{"-f", "my-buildpack"})
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, buildpackRepo.DeleteBuildpack.Name, "my-buildpack")

	assert.Equal(t, len(ui.Prompts), 0)
	assert.Contains(t, ui.Outputs[0], "Deleting buildpack")
	assert.Contains(t, ui.Outputs[0], "my-buildpack")
	assert.Contains(t, ui.Outputs[1], "OK")
}
