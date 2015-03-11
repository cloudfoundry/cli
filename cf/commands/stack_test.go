package commands_test

import (
	"errors"

	testapi "github.com/cloudfoundry/cli/cf/api/stacks/fakes"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/cf/commands"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("stack command", func() {
	var (
		ui                  *testterm.FakeUI
		cmd                 ListStack
		repo                *testapi.FakeStackRepository
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config := testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		repo = &testapi.FakeStackRepository{}
		cmd = NewListStack(ui, config, repo)
	})

	Describe("login requirements", func() {
		It("fails if the user is not logged in", func() {
			requirementsFactory.LoginSuccess = false

			Expect(testcmd.RunCommand(cmd, []string{}, requirementsFactory)).To(BeFalse())
		})
	})

	It("lists the stack requested", func() {
		stack1 := models.Stack{
			Name:        "Stack-1",
			Description: "Stack 1 Description",
			Guid:        "Fake-GUID",
		}

		repo.FindByNameReturns(stack1, nil)

		testcmd.RunCommand(cmd, []string{"Stack-1"}, requirementsFactory)

		Expect(ui.Outputs).To(BeInDisplayOrder(
			[]string{"Getting stack 'Stack-1' in org", "my-org", "my-space", "my-user"},
			[]string{"OK"},
			[]string{"Stack-1", "Stack 1 Description"},
		))
	})

	It("informs user if stack is not found", func() {
		repo.FindByNameReturns(models.Stack{}, errors.New("Stack Stack-1 not found"))

		testcmd.RunCommand(cmd, []string{"Stack-1"}, requirementsFactory)

		Expect(ui.Outputs).To(BeInDisplayOrder(
			[]string{"FAILED"},
			[]string{"Stack Stack-1 not found"},
		))
	})
})
