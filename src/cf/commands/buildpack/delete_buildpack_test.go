package buildpack_test

import (
	"cf"
	. "cf/commands/buildpack"
	"cf/net"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
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
	buildpack := cf.Buildpack{}
	buildpack.Name = "my-buildpack"
	buildpack.Guid = "my-buildpack-guid"
	buildpackRepo := &testapi.FakeBuildpackRepository{
		FindByNameBuildpack: buildpack,
	}
	cmd := NewDeleteBuildpack(ui, buildpackRepo)

	ctxt := testcmd.NewContext("delete-buildpack", []string{"my-buildpack"})
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, buildpackRepo.DeleteBuildpackGuid, "my-buildpack-guid")

	testassert.SliceContains(t, ui.Prompts, testassert.Lines{
		{"delete", "my-buildpack"},
	})
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting buildpack", "my-buildpack"},
		{"OK"},
	})
}

func TestDeleteBuildpackNoConfirmation(t *testing.T) {
	ui := &testterm.FakeUI{Inputs: []string{"no"}}
	buildpack := cf.Buildpack{}
	buildpack.Name = "my-buildpack"
	buildpack.Guid = "my-buildpack-guid"
	buildpackRepo := &testapi.FakeBuildpackRepository{
		FindByNameBuildpack: buildpack,
	}
	cmd := NewDeleteBuildpack(ui, buildpackRepo)

	ctxt := testcmd.NewContext("delete-buildpack", []string{"my-buildpack"})
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, buildpackRepo.DeleteBuildpackGuid, "")

	testassert.SliceContains(t, ui.Prompts, testassert.Lines{
		{"delete", "my-buildpack"},
	})
}

func TestDeleteBuildpackThatDoesNotExist(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	buildpack := cf.Buildpack{}
	buildpack.Name = "my-buildpack"
	buildpack.Guid = "my-buildpack-guid"
	buildpackRepo := &testapi.FakeBuildpackRepository{
		FindByNameNotFound:  true,
		FindByNameBuildpack: buildpack,
	}

	ui := &testterm.FakeUI{}
	ctxt := testcmd.NewContext("delete-buildpack", []string{"-f", "my-buildpack"})

	cmd := NewDeleteBuildpack(ui, buildpackRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, buildpackRepo.FindByNameName, "my-buildpack")
	assert.True(t, buildpackRepo.FindByNameNotFound)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting", "my-buildpack"},
		{"OK"},
		{"my-buildpack", "does not exist"},
	})
}

func TestDeleteBuildpackDeleteError(t *testing.T) {
	ui := &testterm.FakeUI{Inputs: []string{"y"}}
	buildpack := cf.Buildpack{}
	buildpack.Name = "my-buildpack"
	buildpack.Guid = "my-buildpack-guid"
	buildpackRepo := &testapi.FakeBuildpackRepository{
		FindByNameBuildpack: buildpack,
		DeleteApiResponse:   net.NewApiResponseWithMessage("failed badly"),
	}

	cmd := NewDeleteBuildpack(ui, buildpackRepo)

	ctxt := testcmd.NewContext("delete-buildpack", []string{"my-buildpack"})
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, buildpackRepo.DeleteBuildpackGuid, "my-buildpack-guid")

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting buildpack", "my-buildpack"},
		{"FAILED"},
		{"my-buildpack", "failed badly"},
	})
}

func TestDeleteBuildpackForceFlagSkipsConfirmation(t *testing.T) {
	ui := &testterm.FakeUI{}
	buildpack := cf.Buildpack{}
	buildpack.Name = "my-buildpack"
	buildpack.Guid = "my-buildpack-guid"
	buildpackRepo := &testapi.FakeBuildpackRepository{
		FindByNameBuildpack: buildpack,
	}

	cmd := NewDeleteBuildpack(ui, buildpackRepo)

	ctxt := testcmd.NewContext("delete-buildpack", []string{"-f", "my-buildpack"})
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, buildpackRepo.DeleteBuildpackGuid, "my-buildpack-guid")

	assert.Equal(t, len(ui.Prompts), 0)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting buildpack", "my-buildpack"},
		{"OK"},
	})
}
