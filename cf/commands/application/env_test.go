package application_test

import (
	"code.cloudfoundry.org/cli/cf/api/applications/applicationsfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
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

var _ = Describe("env command", func() {
	var (
		ui                  *testterm.FakeUI
		app                 models.Application
		appRepo             *applicationsfakes.FakeRepository
		configRepo          coreconfig.Repository
		requirementsFactory *requirementsfakes.FakeFactory
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetApplicationRepository(appRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("env").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}

		app = models.Application{}
		app.Name = "my-app"
		appRepo = new(applicationsfakes.FakeRepository)
		appRepo.ReadReturns(app, nil)

		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("env", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("Requirements", func() {
		It("fails when the user is not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand("my-app")).To(BeFalse())
		})

		It("fails if a space is not targeted", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Failing{Message: "not targeting space"})
			Expect(runCommand("my-app")).To(BeFalse())
		})
	})

	It("fails with usage when no app name is given", func() {
		passed := runCommand()

		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires", "argument"},
		))
		Expect(passed).To(BeFalse())
	})

	It("fails with usage when the app cannot be found", func() {
		appRepo.ReadReturns(models.Application{}, errors.NewModelNotFoundError("app", "hocus-pocus"))
		runCommand("hocus-pocus")

		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"FAILED"},
			[]string{"not found"},
		))
	})

	Context("when the app has at least one env var", func() {
		BeforeEach(func() {
			app = models.Application{}
			app.Name = "my-app"
			app.GUID = "the-app-guid"

			appRepo.ReadReturns(app, nil)
			appRepo.ReadEnvReturns(&models.Environment{
				Environment: map[string]interface{}{
					"my-key":     "my-value",
					"my-key2":    "my-value2",
					"first-key":  0,
					"first-bool": false,
				},
				System: map[string]interface{}{
					"VCAP_SERVICES": map[string]interface{}{
						"pump-yer-brakes": "drive-slow",
					},
				},
				Application: map[string]interface{}{
					"VCAP_APPLICATION": map[string]interface{}{
						"dis-be-an-app-field": "wit-an-app-value",
						"app-key-1":           0,
						"app-key-2":           false,
					},
				},
			}, nil)
		})

		It("lists those environment variables, in sorted order for provided services", func() {
			runCommand("my-app")
			Expect(appRepo.ReadEnvArgsForCall(0)).To(Equal("the-app-guid"))
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Getting env variables for app", "my-app", "my-org", "my-space", "my-user"},
				[]string{"OK"},
				[]string{"System-Provided:"},
				[]string{"VCAP_SERVICES", ":", "{"},
				[]string{"pump-yer-brakes", ":", "drive-slow"},
				[]string{"}"},
				[]string{"User-Provided:"},
				[]string{"first-bool", "false"},
				[]string{"first-key", "0"},
				[]string{"my-key", "my-value"},
				[]string{"my-key2", "my-value2"},
			))
		})
		It("displays the application env info under the System env column", func() {
			runCommand("my-app")
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Getting env variables for app", "my-app", "my-org", "my-space", "my-user"},
				[]string{"OK"},
				[]string{"System-Provided:"},
				[]string{"VCAP_SERVICES", ":", "{"},
				[]string{"pump-yer-brakes", ":", "drive-slow"},
				[]string{"}"},
				[]string{"VCAP_APPLICATION", ":", "{"},
				[]string{"dis-be-an-app-field", ":", "wit-an-app-value"},
				[]string{"app-key-1", ":", "0"},
				[]string{"app-key-2", ":", "false"},
				[]string{"}"},
			))
		})
	})

	Context("when the app has no user-defined environment variables", func() {
		It("shows an empty message", func() {
			appRepo.ReadEnvReturns(&models.Environment{}, nil)
			runCommand("my-app")

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Getting env variables for app", "my-app"},
				[]string{"OK"},
				[]string{"No", "system-provided", "env variables", "have been set"},
				[]string{"No", "env variables", "have been set"},
			))
		})
	})

	Context("when the app has no environment variables", func() {
		It("informs the user that each group is empty", func() {
			appRepo.ReadEnvReturns(&models.Environment{}, nil)

			runCommand("my-app")
			Expect(ui.Outputs()).To(ContainSubstrings([]string{"No system-provided env variables have been set"}))
			Expect(ui.Outputs()).To(ContainSubstrings([]string{"No user-defined env variables have been set"}))
			Expect(ui.Outputs()).To(ContainSubstrings([]string{"No running env variables have been set"}))
			Expect(ui.Outputs()).To(ContainSubstrings([]string{"No staging env variables have been set"}))
		})
	})

	Context("when the app has at least one running and staging environment variable", func() {
		BeforeEach(func() {
			app = models.Application{}
			app.Name = "my-app"
			app.GUID = "the-app-guid"

			appRepo.ReadReturns(app, nil)
			appRepo.ReadEnvReturns(&models.Environment{
				Running: map[string]interface{}{
					"running-key-1": "running-value-1",
					"running-key-2": "running-value-2",
					"running":       true,
					"number":        37,
				},
				Staging: map[string]interface{}{
					"staging-key-1": "staging-value-1",
					"staging-key-2": "staging-value-2",
					"staging":       false,
					"number":        42,
				},
			}, nil)
		})

		It("lists the environment variables", func() {
			runCommand("my-app")
			Expect(appRepo.ReadEnvArgsForCall(0)).To(Equal("the-app-guid"))
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Getting env variables for app", "my-app", "my-org", "my-space", "my-user"},
				[]string{"OK"},
				[]string{"Running Environment Variable Groups:"},
				[]string{"running-key-1", ":", "running-value-1"},
				[]string{"running-key-2", ":", "running-value-2"},
				[]string{"running", ":", "true"},
				[]string{"number", ":", "37"},
				[]string{"Staging Environment Variable Groups:"},
				[]string{"staging-key-1", ":", "staging-value-1"},
				[]string{"staging-key-2", ":", "staging-value-2"},
				[]string{"staging", ":", "false"},
				[]string{"number", ":", "42"},
			))
		})
	})

	Context("when reading the environment variables returns an error", func() {
		It("tells you about that error", func() {
			appRepo.ReadEnvReturns(nil, errors.New("BOO YOU CANT DO THAT; GO HOME; you're drunk"))
			runCommand("whatever")
			Expect(ui.Outputs()).To(ContainSubstrings([]string{"you're drunk"}))
		})
	})
})
