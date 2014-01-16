package buildpack_test

import (
	. "cf/commands/buildpack"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
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

	ui := callUpdateBuildpack([]string{"my-buildpack"}, reqFactory, repo, bitsRepo)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Updating buildpack", "my-buildpack"},
		{"OK"},
	})
}

func TestUpdateBuildpackPosition(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
	repo, bitsRepo := getRepositories()

	ui := callUpdateBuildpack([]string{"-i", "999", "my-buildpack"}, reqFactory, repo, bitsRepo)

	assert.Equal(t, *repo.UpdateBuildpack.Position, 999)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Updating buildpack", "my-buildpack"},
		{"OK"},
	})
}

func TestUpdateBuildpackNoPosition(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
	repo, bitsRepo := getRepositories()

	callUpdateBuildpack([]string{"my-buildpack"}, reqFactory, repo, bitsRepo)

	assert.Nil(t, repo.UpdateBuildpack.Position)
}

func TestUpdateBuildpackEnabled(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
	repo, bitsRepo := getRepositories()

	fakeUI := callUpdateBuildpack([]string{"--enable", "my-buildpack"}, reqFactory, repo, bitsRepo)

	assert.NotNil(t, repo.UpdateBuildpack.Enabled)
	assert.Equal(t, *repo.UpdateBuildpack.Enabled, true)

	assert.Contains(t, fakeUI.Outputs[0], "Updating buildpack")
	assert.Contains(t, fakeUI.Outputs[0], "my-buildpack")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestUpdateBuildpackDisabled(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
	repo, bitsRepo := getRepositories()

	callUpdateBuildpack([]string{"--disable", "my-buildpack"}, reqFactory, repo, bitsRepo)

	assert.NotNil(t, repo.UpdateBuildpack.Enabled)
	assert.Equal(t, *repo.UpdateBuildpack.Enabled, false)
}

func TestUpdateBuildpackNoEnable(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
	repo, bitsRepo := getRepositories()

	callUpdateBuildpack([]string{"my-buildpack"}, reqFactory, repo, bitsRepo)

	assert.Nil(t, repo.UpdateBuildpack.Enabled)
}

func TestUpdateBuildpackPath(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
	repo, bitsRepo := getRepositories()

	ui := callUpdateBuildpack([]string{"-p", "buildpack.zip", "my-buildpack"}, reqFactory, repo, bitsRepo)

	assert.Equal(t, bitsRepo.UploadBuildpackPath, "buildpack.zip")

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Updating buildpack", "my-buildpack"},
		{"OK"},
	})
}

func TestUpdateBuildpackWithInvalidPath(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
	repo, bitsRepo := getRepositories()
	bitsRepo.UploadBuildpackErr = true

	ui := callUpdateBuildpack([]string{"-p", "bogus/path", "my-buildpack"}, reqFactory, repo, bitsRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Updating buildpack", "my-buildpack"},
		{"FAILED"},
	})
}

func callUpdateBuildpack(args []string, reqFactory *testreq.FakeReqFactory, fakeRepo *testapi.FakeBuildpackRepository,
	fakeBitsRepo *testapi.FakeBuildpackBitsRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("update-buildpack", args)

	cmd := NewUpdateBuildpack(ui, fakeRepo, fakeBitsRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
