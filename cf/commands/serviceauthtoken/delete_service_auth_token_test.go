package serviceauthtoken_test

import (
	"github.com/cloudfoundry/cli/cf/api/apifakes"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/requirements/requirementsfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("delete-service-auth-token command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          coreconfig.Repository
		authTokenRepo       *apifakes.OldFakeAuthTokenRepo
		requirementsFactory *requirementsfakes.FakeFactory
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetServiceAuthTokenRepository(authTokenRepo)
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("delete-service-auth-token").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{Inputs: []string{"y"}}
		authTokenRepo = new(apifakes.OldFakeAuthTokenRepo)
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		requirementsFactory.NewMaxAPIVersionRequirementReturns(requirements.Passing{})
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("delete-service-auth-token", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("fails with usage when fewer than two arguments are given", func() {
			runCommand("yurp")
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})

		It("fails when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand()).To(BeFalse())
		})

		It("requires CC API version 2.47 or lower", func() {
			requirementsFactory.NewMaxAPIVersionRequirementReturns(requirements.Failing{Message: "max api 2.47"})
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			Expect(runCommand("one", "two")).To(BeFalse())
		})
	})

	Context("when the service auth token exists", func() {
		BeforeEach(func() {
			authTokenRepo.FindByLabelAndProviderServiceAuthTokenFields = models.ServiceAuthTokenFields{
				GUID:     "the-guid",
				Label:    "a label",
				Provider: "a provider",
			}
		})

		It("deletes the service auth token", func() {
			runCommand("a label", "a provider")
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Deleting service auth token as", "my-user"},
				[]string{"OK"},
			))

			Expect(authTokenRepo.FindByLabelAndProviderLabel).To(Equal("a label"))
			Expect(authTokenRepo.FindByLabelAndProviderProvider).To(Equal("a provider"))
			Expect(authTokenRepo.DeletedServiceAuthTokenFields.GUID).To(Equal("the-guid"))
		})

		It("does nothing when the user does not confirm", func() {
			ui.Inputs = []string{"nope"}
			runCommand("a label", "a provider")

			Expect(ui.Prompts).To(ContainSubstrings(
				[]string{"Really delete", "service auth token", "a label", "a provider"},
			))
			Expect(ui.Outputs()).To(BeEmpty())
			Expect(authTokenRepo.DeletedServiceAuthTokenFields).To(Equal(models.ServiceAuthTokenFields{}))
		})

		It("does not prompt the user when the -f flag is given", func() {
			ui.Inputs = []string{}
			runCommand("-f", "a label", "a provider")

			Expect(ui.Prompts).To(BeEmpty())
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Deleting"},
				[]string{"OK"},
			))

			Expect(authTokenRepo.DeletedServiceAuthTokenFields.GUID).To(Equal("the-guid"))
		})
	})

	Context("when the service auth token does not exist", func() {
		BeforeEach(func() {
			authTokenRepo.FindByLabelAndProviderAPIResponse = errors.NewModelNotFoundError("Service Auth Token", "")
		})

		It("warns the user when the specified service auth token does not exist", func() {
			runCommand("a label", "a provider")

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Deleting service auth token as", "my-user"},
				[]string{"OK"},
			))

			Expect(ui.WarnOutputs).To(ContainSubstrings([]string{"does not exist"}))
		})
	})

	Context("when there is an error deleting the service auth token", func() {
		BeforeEach(func() {
			authTokenRepo.FindByLabelAndProviderAPIResponse = errors.New("OH NOES")
		})

		It("shows the user an error", func() {
			runCommand("a label", "a provider")
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Deleting service auth token as", "my-user"},
				[]string{"FAILED"},
				[]string{"OH NOES"},
			))
		})
	})
})
