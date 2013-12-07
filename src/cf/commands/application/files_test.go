package application_test

import (
	"cf"
	. "cf/commands/application"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
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
	app := cf.Application{}
	app.Name = "my-found-app"
	app.Guid = "my-app-guid"

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: app}
	appFilesRepo := &testapi.FakeAppFilesRepo{FileList: "file 1\nfile 2"}

	ui := callFiles(t, []string{"my-app", "/foo"}, reqFactory, appFilesRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Getting files for app", "my-found-app", "my-org", "my-space", "my-user"},
		{"OK"},
		{"file 1\nfile 2"},
	})

	assert.Equal(t, appFilesRepo.AppGuid, "my-app-guid")
	assert.Equal(t, appFilesRepo.Path, "/foo")
}

func TestListingFilesWithTemplateTokens(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-found-app"
	app.Guid = "my-app-guid"

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: app}
	appFilesRepo := &testapi.FakeAppFilesRepo{FileList: "%s %d %i"}

	ui := callFiles(t, []string{"my-app", "/foo"}, reqFactory, appFilesRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"%s %d %i"},
	})
}

func callFiles(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, appFilesRepo *testapi.FakeAppFilesRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("files", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	space := cf.SpaceFields{}
	space.Name = "my-space"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewFiles(ui, config, appFilesRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
