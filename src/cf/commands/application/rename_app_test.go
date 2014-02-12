package application_test

import (
	. "cf/commands/application"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("Testing with ginkgo", func() {
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
		Expect(reqFactory.ApplicationName).To(Equal("my-app"))
	})
	It("TestRenameRun", func() {

		appRepo := &testapi.FakeApplicationRepository{}
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Application: app}
		ui := callRename(mr.T(), []string{"my-app", "my-new-app"}, reqFactory, appRepo)

		Expect(appRepo.UpdateAppGuid).To(Equal(app.Guid))
		Expect(*appRepo.UpdateParams.Name).To(Equal("my-new-app"))
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Renaming app", "my-app", "my-new-app", "my-org", "my-space", "my-user"},
			{"OK"},
		})
	})
})

func callRename(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, appRepo *testapi.FakeApplicationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("rename", args)

	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewRenameApp(ui, configRepo, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
