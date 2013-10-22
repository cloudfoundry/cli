package application_test

import (
	"cf"
	. "cf/commands/application"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestFilesRequirements(t *testing.T) {
	args := []string{"my-app", "/foo"}
	appFilesRepo := &testapi.FakeAppFilesRepo{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true, Application: cf.Application{}}
	callFiles(t, args, reqFactory, appFilesRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false, Application: cf.Application{}}
	callFiles(t, args, reqFactory, appFilesRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: cf.Application{}}
	callFiles(t, args, reqFactory, appFilesRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")
}

func TestFilesFailsWithUsage(t *testing.T) {
	appFilesRepo := &testapi.FakeAppFilesRepo{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: cf.Application{}}
	ui := callFiles(t, []string{}, reqFactory, appFilesRepo)

	assert.True(t, ui.FailedWithUsage)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestListingDirectoryEntries(t *testing.T) {
	app := cf.Application{Name: "my-found-app", Guid: "my-app-guid"}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: app}
	appFilesRepo := &testapi.FakeAppFilesRepo{FileList: "file 1\nfile 2"}

	ui := callFiles(t, []string{"my-app", "/foo"}, reqFactory, appFilesRepo)

	assert.Contains(t, ui.Outputs[0], "Getting files for app")
	assert.Contains(t, ui.Outputs[0], "my-found-app")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Equal(t, appFilesRepo.Application.Guid, "my-app-guid")
	assert.Equal(t, appFilesRepo.Path, "/foo")

	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "file 1\nfile 2")
}

func callFiles(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, appFilesRepo *testapi.FakeAppFilesRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("files", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewFiles(ui, config, appFilesRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
