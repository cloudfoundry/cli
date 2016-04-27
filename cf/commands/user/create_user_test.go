package user_test

import (
	"github.com/cloudfoundry/cli/cf/api/apifakes"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cli/cf/commandregistry"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("Create user command", func() {
	var (
		requirementsFactory *testreq.FakeReqFactory
		ui                  *testterm.FakeUI
		userRepo            *apifakes.FakeUserRepository
		config              coreconfig.Repository
		deps                commandregistry.Dependency
	)

	BeforeEach(func() {
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		ui = new(testterm.FakeUI)
		userRepo = new(apifakes.FakeUserRepository)
		config = testconfig.NewRepositoryWithDefaults()
		accessToken, _ := testconfig.EncodeAccessToken(coreconfig.TokenInfo{
			Username: "current-user",
		})
		config.SetAccessToken(accessToken)
	})

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = config
		deps.RepoLocator = deps.RepoLocator.SetUserRepository(userRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("create-user").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("create-user", args, requirementsFactory, updateCommandDependency, false)
	}

	It("creates a user", func() {
		runCommand("my-user", "my-password")

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating user", "my-user"},
			[]string{"OK"},
			[]string{"TIP"},
		))

		userName, password := userRepo.CreateArgsForCall(0)
		Expect(userName).To(Equal("my-user"))
		Expect(password).To(Equal("my-password"))
	})

	It("creates a user when '--origin uaa' is specified", func() {
		runCommand("my-user", "my-password", "--origin", "uaa")

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating user", "my-user"},
			[]string{"OK"},
			[]string{"TIP"},
		))

		userName, password := userRepo.CreateArgsForCall(0)
		Expect(userName).To(Equal("my-user"))
		Expect(password).To(Equal("my-password"))
	})

	It("creates a user when '--origin ldap' is specified", func() {
		runCommand("my-user", "my-external-id", "--origin", "ldap")

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating user", "my-user"},
			[]string{"OK"},
			[]string{"TIP"},
		))

		userName, origin, externalid := userRepo.CreateExternalArgsForCall(0)
		Expect(userName).To(Equal("my-user"))
		Expect(origin).To(Equal("ldap"))
		Expect(externalid).To(Equal("my-external-id"))
	})

	Context("when creating the user returns an error", func() {
		It("prints a warning when the given user already exists", func() {
			userRepo.CreateReturns(errors.NewModelAlreadyExistsError("User", "my-user"))

			runCommand("my-user", "my-password")

			Expect(ui.WarnOutputs).To(ContainSubstrings(
				[]string{"already exists"},
			))

			Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"FAILED"}))
		})

		It("fails when any error other than alreadyExists is returned", func() {
			userRepo.CreateReturns(errors.NewHTTPError(403, "403", "Forbidden"))

			runCommand("my-user", "my-password")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Forbidden"},
			))

			Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))

		})
	})

	It("fails when no arguments are passed", func() {
		Expect(runCommand()).To(BeFalse())
	})

	It("fails when the user is not logged in", func() {
		requirementsFactory.LoginSuccess = false

		Expect(runCommand("my-user", "my-password")).To(BeFalse())
	})
})
