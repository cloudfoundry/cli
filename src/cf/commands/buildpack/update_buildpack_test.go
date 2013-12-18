package buildpack_test

import (
	. "cf/commands/buildpack"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestUpdateBuildpackRequirements(t *testing.T) {
	repo, bitsRepo := getRepositories()

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
	callUpdateBuildpack([]string{"my-buildpack"}, reqFactory, repo, bitsRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: false}
	callUpdateBuildpack([]string{"my-buildpack", "-p", "buildpack.zip", "extraArg"}, reqFactory, repo, bitsRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: false}
	callUpdateBuildpack([]string{"my-buildpack"}, reqFactory, repo, bitsRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, BuildpackSuccess: true}
	callUpdateBuildpack([]string{"my-buildpack"}, reqFactory, repo, bitsRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestUpdateBuildpack(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
	repo, bitsRepo := getRepositories()

	fakeUI := callUpdateBuildpack([]string{"my-buildpack"}, reqFactory, repo, bitsRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Updating buildpack")
	assert.Contains(t, fakeUI.Outputs[0], "my-buildpack")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestUpdateBuildpackPosition(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
	repo, bitsRepo := getRepositories()

	fakeUI := callUpdateBuildpack([]string{"-i", "999", "my-buildpack"}, reqFactory, repo, bitsRepo)

	assert.Equal(t, *repo.UpdateBuildpack.Position, 999)

	assert.Contains(t, fakeUI.Outputs[0], "Updating buildpack")
	assert.Contains(t, fakeUI.Outputs[0], "my-buildpack")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestUpdateBuildpackEnabled(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
	repo, bitsRepo := getRepositories()

	fakeUI := callUpdateBuildpack([]string{"-enabled", "my-buildpack"}, reqFactory, repo, bitsRepo)

	assert.NotNil(t, repo.UpdateBuildpack.Enabled)
	assert.Equal(t, *repo.UpdateBuildpack.Enabled, true)

	assert.Contains(t, fakeUI.Outputs[0], "Updating buildpack")
	assert.Contains(t, fakeUI.Outputs[0], "my-buildpack")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestUpdateBuildpackPath(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
	repo, bitsRepo := getRepositories()

	fakeUI := callUpdateBuildpack([]string{"-p", "buildpack.zip", "my-buildpack"}, reqFactory, repo, bitsRepo)

	assert.Equal(t, bitsRepo.UploadBuildpackPath, "buildpack.zip")

	assert.Contains(t, fakeUI.Outputs[0], "Updating buildpack")
	assert.Contains(t, fakeUI.Outputs[0], "my-buildpack")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestUpdateBuildpackWithInvalidPath(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
	repo, bitsRepo := getRepositories()
	bitsRepo.UploadBuildpackErr = true

	fakeUI := callUpdateBuildpack([]string{"-p", "bogus/path", "my-buildpack"}, reqFactory, repo, bitsRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Updating buildpack")
	assert.Contains(t, fakeUI.Outputs[0], "my-buildpack")
	assert.Contains(t, fakeUI.Outputs[1], "FAILED")
}

func callUpdateBuildpack(args []string, reqFactory *testreq.FakeReqFactory, fakeRepo *testapi.FakeBuildpackRepository,
	fakeBitsRepo *testapi.FakeBuildpackBitsRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("update-buildpack", args)

	cmd := NewUpdateBuildpack(ui, fakeRepo, fakeBitsRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
