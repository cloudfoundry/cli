package user_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/user"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("delete-user command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          configuration.ReadWriter
		userRepo            *testapi.FakeUserRepository
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{Inputs: []string{"y"}}
		userRepo = &testapi.FakeUserRepository{}
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		configRepo = testconfig.NewRepositoryWithDefaults()

		token, err := testconfig.EncodeAccessToken(configuration.TokenInfo{
			UserGuid: "admin-user-guid",
			Username: "admin-user",
		})
		Expect(err).ToNot(HaveOccurred())
		configRepo.SetAccessToken(token)
	})

	runCommand := func(args ...string) {
		cmd := NewDeleteUser(ui, configRepo, userRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext("delete-user", args), requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory.LoginSuccess = false
			runCommand("my-user")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails with usage when no arguments are given", func() {
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})

	Context("when the given user exists", func() {
		BeforeEach(func() {
			userRepo.FindByUsernameUserFields = models.UserFields{
				Username: "user-name",
				Guid:     "user-guid",
			}
		})

		It("deletes a user with the given name", func() {
			runCommand("user-name")

			Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the user user-name"}))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Deleting user", "user-name", "admin-user"},
				[]string{"OK"},
			))

			Expect(userRepo.FindByUsernameUsername).To(Equal("user-name"))
			Expect(userRepo.DeleteUserGuid).To(Equal("user-guid"))
		})

		It("does not delete the user when no confirmation is given", func() {
			ui.Inputs = []string{"nope"}
			runCommand("user")

			Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete"}))
			Expect(userRepo.FindByUsernameUsername).To(Equal(""))
			Expect(userRepo.DeleteUserGuid).To(Equal(""))
		})

		It("deletes without confirmation when the -f flag is given", func() {
			ui.Inputs = []string{}
			runCommand("-f", "user-name")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Deleting user", "user-name"},
				[]string{"OK"},
			))

			Expect(userRepo.FindByUsernameUsername).To(Equal("user-name"))
			Expect(userRepo.DeleteUserGuid).To(Equal("user-guid"))
		})
	})

	Context("when the given user does not exist", func() {
		BeforeEach(func() {
			userRepo.FindByUsernameNotFound = true
		})

		It("prints a warning", func() {
			runCommand("-f", "user-name")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Deleting user", "user-name"},
				[]string{"OK"},
			))

			Expect(ui.WarnOutputs).To(ContainSubstrings([]string{"user-name", "does not exist"}))

			Expect(userRepo.FindByUsernameUsername).To(Equal("user-name"))
			Expect(userRepo.DeleteUserGuid).To(Equal(""))
		})
	})
})
