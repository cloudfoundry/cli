package application_test

import (
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/application"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("env command", func() {
	var (
		ui                  *testterm.FakeUI
		app                 models.Application
		appRepo             *testapi.FakeApplicationRepository
		configRepo          configuration.ReadWriter
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}

		app = models.Application{}
		app.Name = "my-app"
		appRepo = &testapi.FakeApplicationRepository{}
		appRepo.ReadReturns.App = app

		configRepo = testconfig.NewRepositoryWithDefaults()

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
	})

	runCommand := func(args ...string) {
		cmd := NewEnv(ui, configRepo, appRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext(cmd.Metadata().Name, args), requirementsFactory)
	}

	Describe("Requirements", func() {
		It("fails when the user is not logged in", func() {
			requirementsFactory.LoginSuccess = false
			runCommand("my-app")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	It("fails with usage when no app name is given", func() {
		runCommand()

		Expect(ui.FailedWithUsage).To(BeTrue())
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	It("fails with usage when the app cannot be found", func() {
		appRepo.ReadReturns.Error = errors.NewModelNotFoundError("app", "hocus-pocus")
		runCommand("hocus-pocus")

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"FAILED"},
			[]string{"not found"},
		))
	})

	Context("when the app has at least one env var", func() {
		BeforeEach(func() {
			app = models.Application{}
			app.Name = "my-app"
			app.EnvironmentVars = map[string]string{
				"my-key":  "my-value",
				"my-key2": "my-value2",
			}
			appRepo.ReadReturns.App = app
		})

		It("TestEnvListsEnvironmentVariables", func() {

			runCommand("my-app")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting env variables for app", "my-app", "my-org", "my-space", "my-user"},
				[]string{"OK"},
				[]string{"my-key", "my-value", "my-key2", "my-value2"},
			))
		})
	})

	Context("when the app has no environment variables", func() {
		It("shows an empty message", func() {
			runCommand("my-app")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting env variables for app", "my-app"},
				[]string{"OK"},
				[]string{"No env variables exist"},
			))
		})
	})
})
