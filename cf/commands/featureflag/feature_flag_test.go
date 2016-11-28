package featureflag_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/api/featureflags/featureflagsfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("feature-flag command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *requirementsfakes.FakeFactory
		configRepo          coreconfig.Repository
		flagRepo            *featureflagsfakes.FakeFeatureFlagRepository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetFeatureFlagRepository(flagRepo)
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("feature-flag").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		flagRepo = new(featureflagsfakes.FakeFeatureFlagRepository)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("feature-flag", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("requires the user to be logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand("foo")).ToNot(HavePassedRequirements())
		})

		It("requires the user to provide a feature flag", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			Expect(runCommand()).ToNot(HavePassedRequirements())
		})
	})

	Describe("when logged in", func() {
		BeforeEach(func() {
			flag := models.FeatureFlag{
				Name:    "route_creation",
				Enabled: false,
			}
			flagRepo.FindByNameReturns(flag, nil)
		})

		It("lists the state of the specified feature flag", func() {
			runCommand("route_creation")
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Retrieving status of route_creation as my-user..."},
				[]string{"Feature", "State"},
				[]string{"route_creation", "disabled"},
			))
		})

		Context("when an error occurs", func() {
			BeforeEach(func() {
				flagRepo.FindByNameReturns(models.FeatureFlag{}, errors.New("An error occurred."))
			})

			It("fails with an error", func() {
				runCommand("route_creation")
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"An error occurred."},
				))
			})
		})
	})
})
