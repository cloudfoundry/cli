package commands_test

import (
	. "cf/commands"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testreq "testhelpers/requirements"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testterm "testhelpers/terminal"
)

var _ = Describe("stacks command", func() {
	var (
		ui         *testterm.FakeUI
		cmd        Stacks
		repo       *testapi.FakeStackRepository
		reqFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config := testconfig.NewRepositoryWithDefaults()
		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		repo = &testapi.FakeStackRepository{}
		cmd = NewStacks(ui, config, repo)
	})

	Describe("login requirements", func() {
		It("fails if the user is not logged in", func() {
			reqFactory.LoginSuccess = false
			context := testcmd.NewContext("stacks", []string{})
			testcmd.RunCommand(cmd, context, reqFactory)
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	It("lists the stacks", func() {
		stack1 := models.Stack{
			Name: "Stack-1",
			Description: "Stack 1 Description",
		}
		stack2 := models.Stack{
			Name: "Stack-2",
			Description: "Stack 2 Description",
		}

		repo.FindAllStacks = []models.Stack{stack1, stack2}
		context := testcmd.NewContext("stacks", []string{})
		testcmd.RunCommand(cmd, context, reqFactory)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Getting stacks in org", "my-org", "my-space", "my-user"},
			{"OK"},
			{"Stack-1", "Stack 1 Description"},
			{"Stack-2", "Stack 2 Description"},
		})
	})
})
