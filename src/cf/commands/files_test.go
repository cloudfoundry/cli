package commands_test

import (
	"cf"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestFilesRequirements(t *testing.T) {
	args := []string{"my-app", "/foo"}
	appFilesRepo := &testhelpers.FakeAppFilesRepo{}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true, Application: cf.Application{}}
	callFiles(args, reqFactory, appFilesRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false, Application: cf.Application{}}
	callFiles(args, reqFactory, appFilesRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: cf.Application{}}
	callFiles(args, reqFactory, appFilesRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")
}

func TestFilesFailsWithUsage(t *testing.T) {
	appFilesRepo := &testhelpers.FakeAppFilesRepo{}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: cf.Application{}}
	ui := callFiles([]string{}, reqFactory, appFilesRepo)

	assert.True(t, ui.FailedWithUsage)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestListingDirectoryEntries(t *testing.T) {
	app := cf.Application{Guid: "my-app-guid"}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: app}
	appFilesRepo := &testhelpers.FakeAppFilesRepo{FileList: "file 1\nfile 2"}

	ui := callFiles([]string{"my-app", "/foo"}, reqFactory, appFilesRepo)

	assert.Contains(t, ui.Outputs[0], "Getting files...")
	assert.Equal(t, appFilesRepo.Application.Guid, "my-app-guid")
	assert.Equal(t, appFilesRepo.Path, "/foo")

	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "file 1\nfile 2")
}

func callFiles(args []string, reqFactory *testhelpers.FakeReqFactory, appFilesRepo *testhelpers.FakeAppFilesRepo) (ui *testhelpers.FakeUI) {
	ui = &testhelpers.FakeUI{}
	ctxt := testhelpers.NewContext("files", args)
	cmd := NewFiles(ui, appFilesRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	return
}
