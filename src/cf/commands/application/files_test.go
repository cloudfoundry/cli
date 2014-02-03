package application_test

import (
	"cf"
	. "cf/commands/application"
	"cf/configuration"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callFiles(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, appFilesRepo *testapi.FakeAppFilesRepo) (ui *testterm.FakeUI) {
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
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestFilesRequirements", func() {
			args := []string{"my-app", "/foo"}
			appFilesRepo := &testapi.FakeAppFilesRepo{}

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true, Application: cf.Application{}}
			callFiles(mr.T(), args, reqFactory, appFilesRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false, Application: cf.Application{}}
			callFiles(mr.T(), args, reqFactory, appFilesRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: cf.Application{}}
			callFiles(mr.T(), args, reqFactory, appFilesRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
			assert.Equal(mr.T(), reqFactory.ApplicationName, "my-app")
		})
		It("TestFilesFailsWithUsage", func() {

			appFilesRepo := &testapi.FakeAppFilesRepo{}
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: cf.Application{}}
			ui := callFiles(mr.T(), []string{}, reqFactory, appFilesRepo)

			assert.True(mr.T(), ui.FailedWithUsage)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestListingDirectoryEntries", func() {

			app := cf.Application{}
			app.Name = "my-found-app"
			app.Guid = "my-app-guid"

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: app}
			appFilesRepo := &testapi.FakeAppFilesRepo{FileList: "file 1\nfile 2"}

			ui := callFiles(mr.T(), []string{"my-app", "/foo"}, reqFactory, appFilesRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting files for app", "my-found-app", "my-org", "my-space", "my-user"},
				{"OK"},
				{"file 1"},
				{"file 2"},
			})

			assert.Equal(mr.T(), appFilesRepo.AppGuid, "my-app-guid")
			assert.Equal(mr.T(), appFilesRepo.Path, "/foo")
		})
		It("TestListingFilesWithTemplateTokens", func() {

			app := cf.Application{}
			app.Name = "my-found-app"
			app.Guid = "my-app-guid"

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: app}
			appFilesRepo := &testapi.FakeAppFilesRepo{FileList: "%s %d %i"}

			ui := callFiles(mr.T(), []string{"my-app", "/foo"}, reqFactory, appFilesRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"%s %d %i"},
			})
		})
	})
}
