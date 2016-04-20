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

var _ = Describe("Create LDAP user command", func() {
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
		deps.Ui = ui
		deps.Config = config
		deps.RepoLocator = deps.RepoLocator.SetUserRepository(userRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("create-ldap-user").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("create-ldap-user", args, requirementsFactory, updateCommandDependency, false)
	}

	It("creates a user authenticated by LDAP (origin=LDAP)", func() {
		runCommand("my-user", "my-external-id")

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating user", "my-user"},
			[]string{"OK"},
			[]string{"TIP"},
		))

		userName, externalID := userRepo.CreateLDAPArgsForCall(0)
		Expect(userName).To(Equal("my-user"))
		Expect(externalID).To(Equal("my-external-id"))
	})

	Context("when creating the user returns an error", func() {
		It("prints a warning when the given user already exists", func() {
			userRepo.CreateLDAPReturns(errors.NewModelAlreadyExistsError("User", "my-user"))

			runCommand("my-user", "my-external-id")

			Expect(ui.WarnOutputs).To(ContainSubstrings(
				[]string{"already exists"},
			))

			Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"FAILED"}))
		})

		It("fails when any error other than alreadyExists is returned", func() {
			userRepo.CreateLDAPReturns(errors.NewHttpError(403, "403", "Forbidden"))

			runCommand("my-user", "my-external-id")

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

		Expect(runCommand("my-user", "my-external-id")).To(BeFalse())
	})
})
