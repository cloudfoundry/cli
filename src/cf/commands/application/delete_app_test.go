package application_test

import (
	. "cf/commands/application"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("delete app command", func() {
	var (
		cmd                 *DeleteApp
		ui                  *testterm.FakeUI
		app                 models.Application
		configRepo          configuration.ReadWriter
		appRepo             *testapi.FakeApplicationRepository
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		app = models.Application{}
		app.Name = "app-to-delete"
		app.Guid = "app-to-delete-guid"

		ui = &testterm.FakeUI{}
		appRepo = &testapi.FakeApplicationRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}

		ui = &testterm.FakeUI{}

		configRepo = testconfig.NewRepositoryWithDefaults()
		cmd = NewDeleteApp(ui, configRepo, appRepo)
	})

	It("fails requirements when not logged in", func() {
		requirementsFactory.LoginSuccess = false
		testcmd.RunCommand(cmd, testcmd.NewContext("delete", []string{"-f", "delete-this-app-plz"}), requirementsFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("provides the user usage text when no app name is given", func() {
			testcmd.RunCommand(cmd, testcmd.NewContext("delete", []string{}), requirementsFactory)
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		Context("When provided an app that exists", func() {
			BeforeEach(func() {
				appRepo.ReadApp = app
			})

			It("deletes an app when the user confirms", func() {
				ui.Inputs = []string{"y"}
				context := testcmd.NewContext("delete", []string{"app-to-delete"})
				testcmd.RunCommand(cmd, context, requirementsFactory)

				Expect(appRepo.ReadName).To(Equal("app-to-delete"))
				Expect(appRepo.DeletedAppGuid).To(Equal("app-to-delete-guid"))

				testassert.SliceContains(ui.Prompts, testassert.Lines{
					{"Really delete"},
				})
				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Deleting", "app-to-delete", "my-org", "my-space", "my-user"},
					{"OK"},
				})
			})

			It("does not prompt when the -f flag is provided", func() {
				context := testcmd.NewContext("delete", []string{"-f", "app-to-delete"})
				testcmd.RunCommand(cmd, context, requirementsFactory)

				Expect(appRepo.ReadName).To(Equal("app-to-delete"))
				Expect(appRepo.DeletedAppGuid).To(Equal("app-to-delete-guid"))
				Expect(ui.Prompts).To(BeEmpty())

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Deleting", "app-to-delete"},
					{"OK"},
				})
			})
		})

		Context("when the app provided is not found", func() {
			BeforeEach(func() {
				appRepo.ReadNotFound = true
				context := testcmd.NewContext("delete", []string{"-f", "app-to-delete"})
				testcmd.RunCommand(cmd, context, requirementsFactory)
			})

			It("tells the user when the provided app does not exist", func() {
				Expect(appRepo.ReadName).To(Equal("app-to-delete"))
				Expect(appRepo.DeletedAppGuid).To(Equal(""))

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Deleting", "app-to-delete"},
					{"OK"},
					{"app-to-delete", "does not exist"},
				})
			})
		})
	})
})
