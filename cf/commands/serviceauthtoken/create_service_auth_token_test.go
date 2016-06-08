package serviceauthtoken_test

import (
	"github.com/cloudfoundry/cli/cf/api/apifakes"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create-service-auth-token command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          coreconfig.Repository
		authTokenRepo       *apifakes.OldFakeAuthTokenRepo
		requirementsFactory *testreq.FakeReqFactory
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetServiceAuthTokenRepository(authTokenRepo)
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("create-service-auth-token").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		authTokenRepo = new(apifakes.OldFakeAuthTokenRepo)
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("create-service-auth-token", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("fails with usage when not invoked with exactly three args", func() {
			requirementsFactory.LoginSuccess = true
			runCommand("whoops", "i-accidentally-an-arg")

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})

		It("fails when not logged in", func() {
			Expect(runCommand("just", "enough", "args")).To(BeFalse())
		})

		It("requires CC API version 2.47 or greater", func() {
			requirementsFactory.MaxAPIVersionSuccess = false
			requirementsFactory.LoginSuccess = true
			Expect(runCommand("one", "two", "three")).To(BeFalse())
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.MaxAPIVersionSuccess = true
		})

		It("creates a service auth token, obviously", func() {
			runCommand("a label", "a provider", "a value")

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Creating service auth token as", "my-user"},
				[]string{"OK"},
			))

			authToken := models.ServiceAuthTokenFields{}
			authToken.Label = "a label"
			authToken.Provider = "a provider"
			authToken.Token = "a value"
			Expect(authTokenRepo.CreatedServiceAuthTokenFields).To(Equal(authToken))
		})
	})
})
