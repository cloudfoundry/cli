package application_test

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/api/applications/applicationsfakes"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/requirements/requirementsfakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("set-health-check command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *requirementsfakes.FakeFactory
		appRepo             *applicationsfakes.FakeRepository
		configRepo          coreconfig.Repository
		deps                commandregistry.Dependency
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		appRepo = new(applicationsfakes.FakeRepository)
	})

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetApplicationRepository(appRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("get-health-check").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("get-health-check", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("fails with usage when called without enough arguments", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})

			runCommand()
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"get-health-check"},
				[]string{"Incorrect Usage", "Requires", "argument"},
			))
		})

		It("fails requirements when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand("my-app")).To(BeFalse())
		})

		It("fails if a space is not targeted", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Failing{Message: "not targeting space"})
			Expect(runCommand("my-app")).To(BeFalse())
		})
	})

	Describe("getting health_check_type", func() {
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})
		})

		Context("when application is not found", func() {
			It("Fails", func() {
				appRequirement := new(requirementsfakes.FakeApplicationRequirement)
				appRequirement.ExecuteReturns(errors.New("no app"))
				requirementsFactory.NewApplicationRequirementReturns(appRequirement)
				Expect(runCommand("non-exist-app")).To(BeFalse())
			})
		})

		Context("when application exists", func() {
			BeforeEach(func() {
				app := models.Application{}
				app.Name = "my-app"
				app.GUID = "my-app-guid"
				app.HealthCheckType = "port"

				applicationReq := new(requirementsfakes.FakeApplicationRequirement)
				applicationReq.GetApplicationReturns(app)
				requirementsFactory.NewApplicationRequirementReturns(applicationReq)
			})

			It("shows the health_check_type", func() {
				runCommand("my-app")

				Expect(ui.Outputs()).To(ContainSubstrings([]string{"Getting", "my-app", "health_check_type"}))
				Expect(ui.Outputs()).To(ContainSubstrings([]string{"port"}))
			})
		})
	})

})
