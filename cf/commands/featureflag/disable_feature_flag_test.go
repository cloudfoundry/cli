package featureflag_test

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/api/featureflags/featureflagsfakes"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/requirements/requirementsfakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("disable-feature-flag command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *requirementsfakes.FakeFactory
		flagRepo            *featureflagsfakes.FakeFeatureFlagRepository
		configRepo          coreconfig.Repository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetFeatureFlagRepository(flagRepo)
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("disable-feature-flag").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		flagRepo = new(featureflagsfakes.FakeFeatureFlagRepository)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("disable-feature-flag", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("requires the user to be logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand()).ToNot(HavePassedRequirements())
		})

		It("fails with usage if a single feature is not specified", func() {
			runCommand()
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires an argument"},
			))
		})
	})

	Describe("when logged in", func() {
		BeforeEach(func() {
			flagRepo.UpdateReturns(nil)
		})

		It("Sets the flag", func() {
			runCommand("user_org_creation")

			flag, set := flagRepo.UpdateArgsForCall(0)
			Expect(flag).To(Equal("user_org_creation"))
			Expect(set).To(BeFalse())

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Setting status of user_org_creation as my-user..."},
				[]string{"OK"},
				[]string{"Feature user_org_creation Disabled."},
			))
		})

		Context("when an error occurs", func() {
			BeforeEach(func() {
				flagRepo.UpdateReturns(errors.New("An error occurred."))
			})

			It("fails with an error", func() {
				runCommand("i-dont-exist")
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"An error occurred."},
				))
			})
		})
	})
})
