package user_test

import (
	. "cf/commands/user"
	"cf/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"

	. "testhelpers/matchers"
)

var _ = Describe("Create user command", func() {
	var (
		requirementsFactory *testreq.FakeReqFactory
		ui                  *testterm.FakeUI
		userRepo            *testapi.FakeUserRepository
		configRepo          configuration.ReadWriter
	)

	BeforeEach(func() {
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		ui = new(testterm.FakeUI)
		userRepo = &testapi.FakeUserRepository{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		accessToken, _ := testconfig.EncodeAccessToken(configuration.TokenInfo{
			Username: "current-user",
		})
		configRepo.SetAccessToken(accessToken)
	})

	runCommand := func(args ...string) {
		ctxt := testcmd.NewContext("create-user", args)
		cmd := NewCreateUser(ui, configRepo, userRepo)
		testcmd.RunCommand(cmd, ctxt, requirementsFactory)
		return
	}

	It("creates a user", func() {
		runCommand("my-user", "my-password")

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating user", "my-user", "current-user"},
			[]string{"OK"},
			[]string{"TIP"},
		))

		Expect(userRepo.CreateUserUsername).To(Equal("my-user"))
	})

	It("prints a warning when the given user already exists", func() {
		userRepo.CreateUserExists = true

		runCommand("my-user", "my-password")

		Expect(ui.WarnOutputs).To(ContainSubstrings(
			[]string{"already exists"},
		))

		Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"FAILED"}))
	})

	It("fails when no arguments are passed", func() {
		runCommand()
		Expect(ui.FailedWithUsage).To(BeTrue())
	})

	It("fails when the user is not logged in", func() {
		requirementsFactory.LoginSuccess = false

		runCommand("my-user", "my-password")
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
})
