package buildpack_test

import (
	"cf"
	. "cf/commands/buildpack"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestCreateBuildpackRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	repo, bitsRepo := getRepositories()

	repo.FindByNameBuildpack = cf.Buildpack{}
	callCreateBuildpack([]string{"my-buildpack", "my-dir", "0"}, reqFactory, repo, bitsRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = false
	callCreateBuildpack([]string{"my-buildpack", "my-dir", "0"}, reqFactory, repo, bitsRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestCreateBuildpack(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	repo, bitsRepo := getRepositories()
	ui := callCreateBuildpack([]string{"my-buildpack", "my.war", "5"}, reqFactory, repo, bitsRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating buildpack", "my-buildpack"},
		{"OK"},
		{"Uploading buildpack", "my-buildpack"},
		{"OK"},
	})
	testassert.SliceDoesNotContain(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
	})
}

func TestCreateBuildpackWhenItAlreadyExists(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	repo, bitsRepo := getRepositories()

	repo.CreateBuildpackExists = true
	ui := callCreateBuildpack([]string{"my-buildpack", "my.war", "5"}, reqFactory, repo, bitsRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating buildpack", "my-buildpack"},
		{"OK"},
		{"my-buildpack", "already exists"},
		{"tip", "update-buildpack"},
	})
	testassert.SliceDoesNotContain(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
	})
}

func TestCreateBuildpackWithPosition(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	repo, bitsRepo := getRepositories()
	ui := callCreateBuildpack([]string{"my-buildpack", "my.war", "5"}, reqFactory, repo, bitsRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating buildpack", "my-buildpack"},
		{"OK"},
		{"Uploading buildpack", "my-buildpack"},
		{"OK"},
	})
	testassert.SliceDoesNotContain(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
	})
}

func TestCreateBuildpackEnabled(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	repo, bitsRepo := getRepositories()
	ui := callCreateBuildpack([]string{"--enable", "my-buildpack", "my.war", "5"}, reqFactory, repo, bitsRepo)

	assert.NotNil(t, repo.CreateBuildpack.Enabled)
	assert.Equal(t, *repo.CreateBuildpack.Enabled, true)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"creating buildpack", "my-buildpack"},
		{"OK"},
		{"uploading buildpack", "my-buildpack"},
		{"OK"},
	})
	testassert.SliceDoesNotContain(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
	})
}

func TestCreateBuildpackNoEnableFlag(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	repo, bitsRepo := getRepositories()
	callCreateBuildpack([]string{"my-buildpack", "my.war", "5"}, reqFactory, repo, bitsRepo)

	assert.Nil(t, repo.CreateBuildpack.Enabled)
}

func TestCreateBuildpackDisabled(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	repo, bitsRepo := getRepositories()
	callCreateBuildpack([]string{"--disable", "my-buildpack", "my.war", "5"}, reqFactory, repo, bitsRepo)

	assert.NotNil(t, repo.CreateBuildpack.Enabled)
	assert.Equal(t, *repo.CreateBuildpack.Enabled, false)
}

func TestCreateBuildpackWithInvalidPath(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	repo, bitsRepo := getRepositories()

	bitsRepo.UploadBuildpackErr = true
	ui := callCreateBuildpack([]string{"my-buildpack", "bogus/path", "5"}, reqFactory, repo, bitsRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating buildpack", "my-buildpack"},
		{"OK"},
		{"Uploading buildpack"},
		{"FAILED"},
	})
}

func TestCreateBuildpackFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	repo, bitsRepo := getRepositories()

	ui := callCreateBuildpack([]string{}, reqFactory, repo, bitsRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateBuildpack([]string{"my-buildpack", "my.war", "5"}, reqFactory, repo, bitsRepo)
	assert.False(t, ui.FailedWithUsage)
}

func getRepositories() (*testapi.FakeBuildpackRepository, *testapi.FakeBuildpackBitsRepository) {
	return &testapi.FakeBuildpackRepository{}, &testapi.FakeBuildpackBitsRepository{}
}

func callCreateBuildpack(args []string, reqFactory *testreq.FakeReqFactory, fakeRepo *testapi.FakeBuildpackRepository,
	fakeBitsRepo *testapi.FakeBuildpackBitsRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-buildpack", args)

	cmd := NewCreateBuildpack(ui, fakeRepo, fakeBitsRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
