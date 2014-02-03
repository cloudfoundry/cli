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

func callRename(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, appRepo *testapi.FakeApplicationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("rename", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	space := cf.SpaceFields{}
	space.Name = "my-space"
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewRenameApp(ui, config, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestRenameAppFailsWithUsage", func() {
			reqFactory := &testreq.FakeReqFactory{}
			appRepo := &testapi.FakeApplicationRepository{}

			ui := callRename(mr.T(), []string{}, reqFactory, appRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callRename(mr.T(), []string{"foo"}, reqFactory, appRepo)
			assert.True(mr.T(), ui.FailedWithUsage)
		})
		It("TestRenameRequirements", func() {

			appRepo := &testapi.FakeApplicationRepository{}

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			callRename(mr.T(), []string{"my-app", "my-new-app"}, reqFactory, appRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
			assert.Equal(mr.T(), reqFactory.ApplicationName, "my-app")
		})
		It("TestRenameRun", func() {

			appRepo := &testapi.FakeApplicationRepository{}
			app := cf.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Application: app}
			ui := callRename(mr.T(), []string{"my-app", "my-new-app"}, reqFactory, appRepo)

			assert.Equal(mr.T(), appRepo.UpdateAppGuid, app.Guid)
			assert.Equal(mr.T(), appRepo.UpdateParams.Get("name"), "my-new-app")
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Renaming app", "my-app", "my-new-app", "my-org", "my-space", "my-user"},
				{"OK"},
			})
		})
	})
}
