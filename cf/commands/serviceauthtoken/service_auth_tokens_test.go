package serviceauthtoken_test

import (
	"github.com/cloudfoundry/cli/cf/api/apifakes"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/requirements/requirementsfakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	"github.com/cloudfoundry/cli/cf/commands/serviceauthtoken"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("service-auth-tokens command", func() {
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
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("service-auth-tokens").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{Inputs: []string{"y"}}
		authTokenRepo = new(apifakes.OldFakeAuthTokenRepo)
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("service-auth-tokens", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand()).To(BeFalse())
		})

		Context("when arguments are provided", func() {
			var cmd commandregistry.Command
			var flagContext flags.FlagContext

			BeforeEach(func() {
				cmd = &serviceauthtoken.ListServiceAuthTokens{}
				cmd.SetDependency(deps, false)
				flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
			})

			It("should fail with usage", func() {
				flagContext.Parse("blahblah")

				reqs := cmd.Requirements(requirementsFactory, flagContext)

				err := testcmd.RunRequirements(reqs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Incorrect Usage"))
				Expect(err.Error()).To(ContainSubstring("No argument required"))
			})
		})

		It("requires CC API version 2.47 or greater", func() {
			requirementsFactory.NewMaxAPIVersionRequirementReturns(requirements.Failing{Message: "max api 2.47"})
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			Expect(runCommand()).To(BeFalse())
		})
	})

	Context("when logged in and some service auth tokens exist", func() {
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewMaxAPIVersionRequirementReturns(requirements.Passing{})

			authTokenRepo.FindAllAuthTokens = []models.ServiceAuthTokenFields{
				{Label: "a label", Provider: "a provider"},
				{Label: "a second label", Provider: "a second provider"},
			}
		})

		It("shows you the service auth tokens", func() {
			runCommand()

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Getting service auth tokens as", "my-user"},
				[]string{"OK"},
				[]string{"label", "provider"},
				[]string{"a label", "a provider"},
				[]string{"a second label", "a second provider"},
			))
		})
	})
})
