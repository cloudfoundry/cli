package commands_test

import (
	"cf"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestFilesRequirements(t *testing.T) {
	appFilesRepo := &testhelpers.FakeAppFilesRepo{}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: false, SpaceSuccess: true, Application: cf.Application{}}
	callFiles(reqFactory, appFilesRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true, SpaceSuccess: false, Application: cf.Application{}}
	callFiles(reqFactory, appFilesRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true, SpaceSuccess: true, Application: cf.Application{}}
	callFiles(reqFactory, appFilesRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")
}

func TestListingDirectoryEntries(t *testing.T) {
	app := cf.Application{Guid: "my-app-guid"}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, SpaceSuccess: true, Application: app}
	appFilesRepo := &testhelpers.FakeAppFilesRepo{FileList: "file 1\nfile 2"}

	ui := callFiles(reqFactory, appFilesRepo)

	assert.Contains(t, ui.Outputs[0], "Getting files...")
	assert.Equal(t, appFilesRepo.Application.Guid, "my-app-guid")
	assert.Equal(t, appFilesRepo.Path, "/foo")

	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "file 1\nfile 2")
}

func callFiles(reqFactory *testhelpers.FakeReqFactory, appFilesRepo *testhelpers.FakeAppFilesRepo) (ui *testhelpers.FakeUI) {
	ui = &testhelpers.FakeUI{}
	ctxt := testhelpers.NewContext("files", []string{"--app", "my-app", "--path", "/foo"})
	cmd := NewFiles(ui, appFilesRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	return
}
