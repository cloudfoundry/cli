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

	ctxt := testcmd.NewContext("delete-buildpack", []string{"foo.com"})

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.True(t, testcmd.CommandDidPassRequirements)

	// TestDeleteBuildpackNotFound
	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: false}
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, BuildpackSuccess: true}
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestDeleteBuildpackSuccess(t *testing.T) {
	ui := &testterm.FakeUI{Inputs: []string{"y"}}
	buildpackRepo := &testapi.FakeBuildpackRepository{}
	cmd := NewDeleteBuildpack(ui, buildpackRepo)

	ctxt := testcmd.NewContext("delete-buildpack", []string{"foo.com"})
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, buildpackRepo.DeleteBuildpack.Name, "foo.com")

	assert.Contains(t, ui.Prompts[0], "delete")
	assert.Contains(t, ui.Prompts[0], "foo.com")

	assert.Contains(t, ui.Outputs[0], "Deleting buildpack")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteBuildpackNoConfirmation(t *testing.T) {
	ui := &testterm.FakeUI{Inputs: []string{"no"}}
	buildpackRepo := &testapi.FakeBuildpackRepository{}
	cmd := NewDeleteBuildpack(ui, buildpackRepo)

	ctxt := testcmd.NewContext("delete-buildpack", []string{"foo.com"})
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, buildpackRepo.DeleteBuildpack.Name, "")

	assert.Contains(t, ui.Prompts[0], "delete")
	assert.Contains(t, ui.Prompts[0], "foo.com")

	assert.Contains(t, ui.Outputs[0], "Deleting buildpack")
	assert.Contains(t, ui.Outputs[0], "foo.com")

	assert.Equal(t, len(ui.Outputs), 1)
}

func TestDeleteBuildpackDeleteError(t *testing.T) {
	ui := &testterm.FakeUI{Inputs: []string{"y"}}
	buildpackRepo := &testapi.FakeBuildpackRepository{
		FindByNameBuildpack: cf.Buildpack{Name: "foo.com"},
		DeleteApiResponse:   net.NewApiResponseWithMessage("failed badly"),
	}

	cmd := NewDeleteBuildpack(ui, buildpackRepo)

	ctxt := testcmd.NewContext("delete-buildpack", []string{"foo.com"})
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, buildpackRepo.DeleteBuildpack.Name, "foo.com")

	assert.Contains(t, ui.Outputs[0], "Deleting buildpack")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "foo.com")
	assert.Contains(t, ui.Outputs[2], "failed badly")
}

func TestDeleteBuildpackForceFlagSkipsConfirmation(t *testing.T) {
	ui := &testterm.FakeUI{}
	buildpackRepo := &testapi.FakeBuildpackRepository{
		FindByNameBuildpack: cf.Buildpack{Name: "foo.com"},
	}

	cmd := NewDeleteBuildpack(ui, buildpackRepo)

	ctxt := testcmd.NewContext("delete-buildpack", []string{"-f", "foo.com"})
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, buildpackRepo.DeleteBuildpack.Name, "foo.com")

	assert.Equal(t, len(ui.Prompts), 0)
	assert.Contains(t, ui.Outputs[0], "Deleting buildpack")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[1], "OK")
}
