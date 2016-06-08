package featureflag_test

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/api/featureflags/featureflagsfakes"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("enable-feature-flag command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
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
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		flagRepo = new(featureflagsfakes.FakeFeatureFlagRepository)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("enable-feature-flag", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("requires the user to be logged in", func() {
			requirementsFactory.LoginSuccess = false
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
