package serviceauthtoken_test

import (
	"code.cloudfoundry.org/cli/cf/api/apifakes"
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

var _ = Describe("update-service-auth-token command", func() {
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
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("update-service-auth-token").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{Inputs: []string{"y"}}
		authTokenRepo = new(apifakes.OldFakeAuthTokenRepo)
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("update-service-auth-token", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("fails with usage when not provided exactly three args", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			runCommand("some-token-label", "a-provider")
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})

		It("fails when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand("label", "provider", "token")).To(BeFalse())
		})

		It("requires CC API version 2.47 or lower", func() {
			requirementsFactory.NewMaxAPIVersionRequirementReturns(requirements.Failing{Message: "max api 2.47"})
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			Expect(runCommand("one", "two", "three")).To(BeFalse())
		})
	})

	Context("when logged in and the service auth token exists", func() {
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewMaxAPIVersionRequirementReturns(requirements.Passing{})
			foundAuthToken := models.ServiceAuthTokenFields{}
			foundAuthToken.GUID = "found-auth-token-guid"
			foundAuthToken.Label = "found label"
			foundAuthToken.Provider = "found provider"
			authTokenRepo.FindByLabelAndProviderServiceAuthTokenFields = foundAuthToken
		})

		It("updates the service auth token with the provided args", func() {
			runCommand("a label", "a provider", "a value")

			expectedAuthToken := models.ServiceAuthTokenFields{}
			expectedAuthToken.GUID = "found-auth-token-guid"
			expectedAuthToken.Label = "found label"
			expectedAuthToken.Provider = "found provider"
			expectedAuthToken.Token = "a value"

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Updating service auth token as", "my-user"},
				[]string{"OK"},
			))

			Expect(authTokenRepo.FindByLabelAndProviderLabel).To(Equal("a label"))
			Expect(authTokenRepo.FindByLabelAndProviderProvider).To(Equal("a provider"))
			Expect(authTokenRepo.UpdatedServiceAuthTokenFields).To(Equal(expectedAuthToken))
			Expect(authTokenRepo.UpdatedServiceAuthTokenFields).To(Equal(expectedAuthToken))
		})
	})
})
