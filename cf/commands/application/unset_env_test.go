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

   src/github.com/cloudfoundry/cli/cf/commands/application/delete_app_test.go
   src/github.com/cloudfoundry/cli/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package application_test

import (
	"github.com/cloudfoundry/cli/cf/api"
	. "github.com/cloudfoundry/cli/cf/commands/application"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestUnsetEnvRequirements", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		appRepo := &testapi.FakeApplicationRepository{}
		args := []string{"my-app", "DATABASE_URL"}

		requirementsFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		callUnsetEnv(args, requirementsFactory, appRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		requirementsFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: false, TargetedSpaceSuccess: true}
		callUnsetEnv(args, requirementsFactory, appRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		requirementsFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: false}
		callUnsetEnv(args, requirementsFactory, appRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestUnsetEnvWhenApplicationExists", func() {

		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		app.EnvironmentVars = map[string]string{"foo": "bar", "DATABASE_URL": "mysql://example.com/my-db"}
		requirementsFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		appRepo := &testapi.FakeApplicationRepository{}

		args := []string{"my-app", "DATABASE_URL"}
		ui := callUnsetEnv(args, requirementsFactory, appRepo)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Removing env variable", "DATABASE_URL", "my-app", "my-org", "my-space", "my-user"},
			[]string{"OK"},
		))

		Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
		Expect(appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
		Expect(*appRepo.UpdateParams.EnvironmentVars).To(Equal(map[string]string{
			"foo": "bar",
		}))
	})

	It("TestUnsetEnvWhenUnsettingTheEnvFails", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		app.EnvironmentVars = map[string]string{"DATABASE_URL": "mysql://example.com/my-db"}
		requirementsFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		appRepo := &testapi.FakeApplicationRepository{
			UpdateErr: true,
		}
		appRepo.ReadReturns.App = app

		args := []string{"does-not-exist", "DATABASE_URL"}
		ui := callUnsetEnv(args, requirementsFactory, appRepo)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Removing env variable"},
			[]string{"FAILED"},
			[]string{"Error updating app."},
		))
	})

	It("TestUnsetEnvWhenEnvVarDoesNotExist", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		requirementsFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		appRepo := &testapi.FakeApplicationRepository{}

		args := []string{"my-app", "DATABASE_URL"}
		ui := callUnsetEnv(args, requirementsFactory, appRepo)

		Expect(len(ui.Outputs)).To(Equal(3))
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Removing env variable"},
			[]string{"OK"},
			[]string{"DATABASE_URL", "was not set."},
		))
	})

	It("TestUnsetEnvFailsWithUsage", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		requirementsFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		appRepo := &testapi.FakeApplicationRepository{}
		appRepo.ReadReturns.App = app

		args := []string{"my-app", "DATABASE_URL"}
		ui := callUnsetEnv(args, requirementsFactory, appRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())

		args = []string{"my-app"}
		ui = callUnsetEnv(args, requirementsFactory, appRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		args = []string{}
		ui = callUnsetEnv(args, requirementsFactory, appRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())
	})
})

func callUnsetEnv(args []string, requirementsFactory *testreq.FakeReqFactory, appRepo api.ApplicationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewUnsetEnv(ui, configRepo, appRepo)
	testcmd.RunCommand(cmd, args, requirementsFactory)
	return
}
