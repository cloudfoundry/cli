package buildpack_test

import (
	"cf"
	. "cf/commands/buildpack"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestCreateBuildpackRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	repo, bitsRepo := getRepositories()

	repo.FindByNameBuildpack = cf.Buildpack{}
	callCreateBuildpack([]string{"my-buildpack"}, reqFactory, repo, bitsRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
	callCreateBuildpack([]string{"my-buildpack"}, reqFactory, repo, bitsRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestCreateBuildpack(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	repo, bitsRepo := getRepositories()
	fakeUI := callCreateBuildpack([]string{"my-buildpack", "my.war", "5"}, reqFactory, repo, bitsRepo)

	assert.Equal(t, len(fakeUI.Outputs), 5)
	assert.Contains(t, fakeUI.Outputs[0], "Creating buildpack")
	assert.Contains(t, fakeUI.Outputs[0], "my-buildpack")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[3], "Uploading buildpack")
	assert.Contains(t, fakeUI.Outputs[3], "my-buildpack")
	assert.Contains(t, fakeUI.Outputs[4], "OK")
}

func TestCreateBuildpackWhenItAlreadyExists(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	repo, bitsRepo := getRepositories()

	repo.CreateBuildpackExists = true
	fakeUI := callCreateBuildpack([]string{"my-buildpack", "my.war", "5"}, reqFactory, repo, bitsRepo)

	assert.Equal(t, len(fakeUI.Outputs), 3)
	assert.Contains(t, fakeUI.Outputs[0], "Creating buildpack")
	assert.Contains(t, fakeUI.Outputs[0], "my-buildpack")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[2], "my-buildpack")
	assert.Contains(t, fakeUI.Outputs[2], "already exists")
}

func TestCreateBuildpackWithPosition(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	repo, bitsRepo := getRepositories()
	fakeUI := callCreateBuildpack([]string{"my-buildpack", "my.war", "5"}, reqFactory, repo, bitsRepo)

	assert.Equal(t, len(fakeUI.Outputs), 5)
	assert.Contains(t, fakeUI.Outputs[0], "Creating buildpack")
	assert.Contains(t, fakeUI.Outputs[0], "my-buildpack")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[3], "Uploading buildpack")
	assert.Contains(t, fakeUI.Outputs[3], "my-buildpack")
	assert.Contains(t, fakeUI.Outputs[4], "OK")
}

func TestCreateBuildpackDisabled(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	repo, bitsRepo := getRepositories()
	fakeUI := callCreateBuildpack([]string{"-enabled=false", "my-buildpack", "my.war", "5"}, reqFactory, repo, bitsRepo)

	assert.NotNil(t, repo.CreateBuildpack.Enabled)
	assert.Equal(t, *repo.CreateBuildpack.Enabled, false)

	assert.Equal(t, len(fakeUI.Outputs), 5)
	assert.Contains(t, fakeUI.Outputs[0], "Creating buildpack")
	assert.Contains(t, fakeUI.Outputs[0], "my-buildpack")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[3], "Uploading buildpack")
	assert.Contains(t, fakeUI.Outputs[3], "my-buildpack")
	assert.Contains(t, fakeUI.Outputs[4], "OK")
}

func TestCreateBuildpackWithInvalidPath(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	repo, bitsRepo := getRepositories()

	bitsRepo.UploadBuildpackErr = true
	fakeUI := callCreateBuildpack([]string{"my-buildpack", "bogus/path", "5"}, reqFactory, repo, bitsRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Creating buildpack")
	assert.Contains(t, fakeUI.Outputs[0], "my-buildpack")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[3], "Uploading buildpack")
	assert.Contains(t, fakeUI.Outputs[4], "FAILED")
}

func TestCreateBuildpackFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	repo, bitsRepo := getRepositories()

	fakeUI := callCreateBuildpack([]string{}, reqFactory, repo, bitsRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callCreateBuildpack([]string{"my-buildpack", "my.war", "5"}, reqFactory, repo, bitsRepo)
	assert.False(t, fakeUI.FailedWithUsage)
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
