/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/cf/commands/application/delete_app_test.go
   src/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package application_test

import (
	"cf/api"
	. "cf/commands/application"
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

var _ = Describe("set-env command", func() {
	It("TestSetEnvRequirements", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		appRepo := &testapi.FakeApplicationRepository{}
		args := []string{"my-app", "DATABASE_URL", "mysql://example.com/my-db"}

		requirementsFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		callSetEnv(args, requirementsFactory, appRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		requirementsFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: false, TargetedSpaceSuccess: true}
		callSetEnv(args, requirementsFactory, appRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		testcmd.CommandDidPassRequirements = true

		requirementsFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: false}
		callSetEnv(args, requirementsFactory, appRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	It("TestRunWhenApplicationExists", func() {

		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		app.EnvironmentVars = map[string]string{"foo": "bar"}
		requirementsFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		appRepo := &testapi.FakeApplicationRepository{}

		args := []string{"my-app", "DATABASE_URL", "mysql://example.com/my-db"}
		ui := callSetEnv(args, requirementsFactory, appRepo)

		Expect(len(ui.Outputs)).To(Equal(3))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{
				"Setting env variable",
				"DATABASE_URL",
				"mysql://example.com/my-db",
				"my-app",
				"my-org",
				"my-space",
				"my-user",
			},
			{"OK"},
			{"TIP"},
		})

		Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
		Expect(appRepo.UpdateAppGuid).To(Equal(app.Guid))
		Expect(*appRepo.UpdateParams.EnvironmentVars).To(Equal(map[string]string{
			"DATABASE_URL": "mysql://example.com/my-db",
			"foo":          "bar",
		}))
	})

	It("TestSetEnvWhenItAlreadyExists", func() {

		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		app.EnvironmentVars = map[string]string{"DATABASE_URL": "mysql://example.com/my-db"}
		requirementsFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		appRepo := &testapi.FakeApplicationRepository{}

		args := []string{"my-app", "DATABASE_URL", "mysql://example2.com/my-db"}
		ui := callSetEnv(args, requirementsFactory, appRepo)

		Expect(len(ui.Outputs)).To(Equal(3))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{
				"Setting env variable",
				"DATABASE_URL",
				"mysql://example2.com/my-db",
				"my-app",
				"my-org",
				"my-space",
				"my-user",
			},
			{"OK"},
			{"TIP"},
		})

		Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
		Expect(appRepo.UpdateAppGuid).To(Equal(app.Guid))
		Expect(*appRepo.UpdateParams.EnvironmentVars).To(Equal(map[string]string{
			"DATABASE_URL": "mysql://example2.com/my-db",
		}))
	})

	It("TestRunWhenSettingTheEnvFails", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		requirementsFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		appRepo := &testapi.FakeApplicationRepository{UpdateErr: true}
		appRepo.ReadReturns.App = app

		args := []string{"does-not-exist", "DATABASE_URL", "mysql://example.com/my-db"}
		ui := callSetEnv(args, requirementsFactory, appRepo)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Setting env variable"},
			{"FAILED"},
			{"Error updating app."},
		})
	})

	It("TestSetEnvFailsWithUsage", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		requirementsFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		appRepo := &testapi.FakeApplicationRepository{}
		appRepo.ReadReturns.App = app

		args := []string{"my-app", "DATABASE_URL", "..."}
		ui := callSetEnv(args, requirementsFactory, appRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())

		args = []string{"my-app", "DATABASE_URL"}
		ui = callSetEnv(args, requirementsFactory, appRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		args = []string{"my-app"}
		ui = callSetEnv(args, requirementsFactory, appRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		args = []string{}
		ui = callSetEnv(args, requirementsFactory, appRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())
	})
})

func callSetEnv(args []string, requirementsFactory *testreq.FakeReqFactory, appRepo api.ApplicationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("set-env", args)
	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewSetEnv(ui, configRepo, appRepo)
	testcmd.RunCommand(cmd, ctxt, requirementsFactory)
	return
}
