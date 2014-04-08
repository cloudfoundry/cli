package user_test

import (
	. "cf/commands/user"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
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

			testassert.SliceContains(ui.Prompts, testassert.Lines{
				{"Really delete the user user-name"},
			})

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Deleting user", "user-name", "admin-user"},
				{"OK"},
			})

			Expect(userRepo.FindByUsernameUsername).To(Equal("user-name"))
			Expect(userRepo.DeleteUserGuid).To(Equal("user-guid"))
		})

		It("does not delete the user when no confirmation is given", func() {
			ui.Inputs = []string{"nope"}
			runCommand("user")

			testassert.SliceContains(ui.Prompts, testassert.Lines{{"Really delete"}})
			Expect(userRepo.FindByUsernameUsername).To(Equal(""))
			Expect(userRepo.DeleteUserGuid).To(Equal(""))
		})

		It("deletes without confirmation when the -f flag is given", func() {
			ui.Inputs = []string{}
			runCommand("-f", "user-name")

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Deleting user", "user-name"},
				{"OK"},
			})

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

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Deleting user", "user-name"},
				{"OK"},
			})

			testassert.SliceContains(ui.WarnOutputs, testassert.Lines{
				{"user-name", "does not exist"},
			})

			Expect(userRepo.FindByUsernameUsername).To(Equal("user-name"))
			Expect(userRepo.DeleteUserGuid).To(Equal(""))
		})
	})
})
