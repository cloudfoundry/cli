package featureflag_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/api/featureflags/featureflagsfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("enable-feature-flag command", func() {
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
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("enable-feature-flag").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		flagRepo = new(featureflagsfakes.FakeFeatureFlagRepository)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("enable-feature-flag", args, requirementsFactory, updateCommandDependency, false, ui)
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
			Expect(set).To(BeTrue())

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Setting status of user_org_creation as my-user..."},
				[]string{"OK"},
				[]string{"Feature user_org_creation Enabled."},
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
