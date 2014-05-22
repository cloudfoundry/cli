package commands_test

import (
	. "github.com/cloudfoundry/cli/cf/commands"
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

var _ = Describe("stacks command", func() {
	var (
		ui                  *testterm.FakeUI
		cmd                 ListStacks
		repo                *testapi.FakeStackRepository
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config := testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		repo = &testapi.FakeStackRepository{}
		cmd = NewListStacks(ui, config, repo)
	})

	Describe("login requirements", func() {
		It("fails if the user is not logged in", func() {
			requirementsFactory.LoginSuccess = false
			testcmd.RunCommand(cmd, []string{}, requirementsFactory)
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	It("lists the stacks", func() {
		stack1 := models.Stack{
			Name:        "Stack-1",
			Description: "Stack 1 Description",
		}
		stack2 := models.Stack{
			Name:        "Stack-2",
			Description: "Stack 2 Description",
		}

		repo.FindAllStacks = []models.Stack{stack1, stack2}
		testcmd.RunCommand(cmd, []string{}, requirementsFactory)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Getting stacks in org", "my-org", "my-space", "my-user"},
			[]string{"OK"},
			[]string{"Stack-1", "Stack 1 Description"},
			[]string{"Stack-2", "Stack 2 Description"},
		))
	})
})
