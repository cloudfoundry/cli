package user_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cli/cf/command_registry"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("Create user command", func() {
	var (
		requirementsFactory *testreq.FakeReqFactory
		ui                  *testterm.FakeUI
		userRepo            *testapi.FakeUserRepository
		config              core_config.Repository
		deps                command_registry.Dependency
	)

	BeforeEach(func() {
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		ui = new(testterm.FakeUI)
		userRepo = &testapi.FakeUserRepository{}
		config = testconfig.NewRepositoryWithDefaults()
		accessToken, _ := testconfig.EncodeAccessToken(core_config.TokenInfo{
			Username: "current-user",
		})
		config.SetAccessToken(accessToken)
	})

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = config
		deps.RepoLocator = deps.RepoLocator.SetUserRepository(userRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("create-user").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("create-user", args, requirementsFactory, updateCommandDependency, false)
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
			userRepo.CreateReturns(errors.NewHttpError(403, "403", "Forbidden"))

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
