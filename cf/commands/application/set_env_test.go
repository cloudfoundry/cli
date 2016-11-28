package application_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/api/applications/applicationsfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("set-env command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          coreconfig.Repository
		app                 models.Application
		appRepo             *applicationsfakes.FakeRepository
		requirementsFactory *requirementsfakes.FakeFactory
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetApplicationRepository(appRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("set-env").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		app = models.Application{}
		app.Name = "my-app"
		app.GUID = "my-app-guid"
		appRepo = new(applicationsfakes.FakeRepository)
		requirementsFactory = new(requirementsfakes.FakeFactory)
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("set-env", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		BeforeEach(func() {
			applicationReq := new(requirementsfakes.FakeApplicationRequirement)
			applicationReq.GetApplicationReturns(app)
			requirementsFactory.NewApplicationRequirementReturns(applicationReq)
		})

		It("fails when login is not successful", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})

			Expect(runCommand("hey", "gabba", "gabba")).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Failing{Message: "not targeting space"})

			Expect(runCommand("hey", "gabba", "gabba")).To(BeFalse())
		})

		It("fails with usage when not provided with exactly three args", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})

			runCommand("zomg", "too", "many", "args")
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})
	})

	Context("when logged in, a space is targeted and given enough args", func() {
		BeforeEach(func() {
			app.EnvironmentVars = map[string]interface{}{"foo": "bar"}
			applicationReq := new(requirementsfakes.FakeApplicationRequirement)
			applicationReq.GetApplicationReturns(app)
			requirementsFactory.NewApplicationRequirementReturns(applicationReq)
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})
		})

		Context("when it is new", func() {
			It("is created", func() {
				runCommand("my-app", "DATABASE_URL", "mysql://new-example.com/my-db")

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{
						"Setting env variable",
						"DATABASE_URL",
						"mysql://new-example.com/my-db",
						"my-app",
						"my-org",
						"my-space",
						"my-user",
					},
					[]string{"OK"},
					[]string{"TIP"},
				))

				appGUID, params := appRepo.UpdateArgsForCall(0)
				Expect(appGUID).To(Equal(app.GUID))
				Expect(*params.EnvironmentVars).To(Equal(map[string]interface{}{
					"DATABASE_URL": "mysql://new-example.com/my-db",
					"foo":          "bar",
				}))
			})
		})

		Context("when it already exists", func() {
			BeforeEach(func() {
				app.EnvironmentVars["DATABASE_URL"] = "mysql://old-example.com/my-db"
			})

			It("is updated", func() {
				runCommand("my-app", "DATABASE_URL", "mysql://new-example.com/my-db")

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{
						"Setting env variable",
						"DATABASE_URL",
						"mysql://new-example.com/my-db",
						"my-app",
						"my-org",
						"my-space",
						"my-user",
					},
					[]string{"OK"},
					[]string{"TIP"},
				))
			})
		})

		It("allows the variable value to begin with a hyphen", func() {
			runCommand("my-app", "MY_VAR", "--has-a-cool-value")

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{
					"Setting env variable",
					"MY_VAR",
					"--has-a-cool-value",
				},
				[]string{"OK"},
				[]string{"TIP"},
			))
			_, params := appRepo.UpdateArgsForCall(0)
			Expect(*params.EnvironmentVars).To(Equal(map[string]interface{}{
				"MY_VAR": "--has-a-cool-value",
				"foo":    "bar",
			}))
		})

		Context("when setting fails", func() {
			BeforeEach(func() {
				appRepo.UpdateReturns(models.Application{}, errors.New("Error updating app."))
			})

			It("tells the user", func() {
				runCommand("please", "dont", "fail")

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Setting env variable"},
					[]string{"FAILED"},
					[]string{"Error updating app."},
				))
			})
		})

		It("gives the appropriate tip", func() {
			runCommand("my-app", "DATABASE_URL", "mysql://new-example.com/my-db")
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"TIP: Use 'cf restage my-app' to ensure your env variable changes take effect"},
			))
		})
	})
})
