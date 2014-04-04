/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/cf/commands/application/delete_app_test.go
   src/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package user_test

import (
	. "cf/commands/user"
	"cf/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("Create user command", func() {
	var (
		reqFactory *testreq.FakeReqFactory
		ui         *testterm.FakeUI
		userRepo   *testapi.FakeUserRepository
		configRepo configuration.ReadWriter
	)

	BeforeEach(func() {
		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		ui = new(testterm.FakeUI)
		userRepo = &testapi.FakeUserRepository{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		accessToken, _ := testconfig.EncodeAccessToken(configuration.TokenInfo{
			Username: "current-user",
		})
		configRepo.SetAccessToken(accessToken)
	})

	var runCommand = func(args ...string) {
		ctxt := testcmd.NewContext("create-user", args)
		cmd := NewCreateUser(ui, configRepo, userRepo)
		testcmd.RunCommand(cmd, ctxt, reqFactory)
		return
	}

	It("creates a user", func() {
		runCommand("my-user", "my-password")

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Creating user", "my-user", "current-user"},
			{"OK"},
			{"TIP"},
		})

		Expect(userRepo.CreateUserUsername).To(Equal("my-user"))
	})

	It("prints a warning when the given user already exists", func() {
		userRepo.CreateUserExists = true

		runCommand("my-user", "my-password")

		testassert.SliceContains(ui.WarnOutputs, testassert.Lines{
			{"already exists"},
		})

		testassert.SliceDoesNotContain(ui.Outputs, testassert.Lines{
			{"Failed"},
		})
	})

	It("fails when no arguments are passed", func() {
		runCommand()
		Expect(ui.FailedWithUsage).To(BeTrue())
	})

	It("fails when the user is not logged in", func() {
		reqFactory.LoginSuccess = false

		runCommand("my-user", "my-password")
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
})
