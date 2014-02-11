package commands_test

import (
	. "cf/commands"
	"cf/models"
	. "github.com/onsi/ginkgo"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testterm "testhelpers/terminal"
)

func callStacks(stackRepo *testapi.FakeStackRepository) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("stacks", []string{})
	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewStacks(ui, configRepo, stackRepo)
	testcmd.RunCommand(cmd, ctxt, nil)

	return
}
func init() {
	Describe("stacks command", func() {
		It("lists the stacks", func() {
			stack1 := models.Stack{}
			stack1.Name = "Stack-1"
			stack1.Description = "Stack 1 Description"

			stack2 := models.Stack{}
			stack2.Name = "Stack-2"
			stack2.Description = "Stack 2 Description"

			stackRepo := &testapi.FakeStackRepository{
				FindAllStacks: []models.Stack{stack1, stack2},
			}

			ui := callStacks(stackRepo)
			testassert.SliceContains(GinkgoT(), ui.Outputs, testassert.Lines{
				{"Getting stacks in org", "my-org", "my-space", "my-user"},
				{"OK"},
				{"Stack-1", "Stack 1 Description"},
				{"Stack-2", "Stack 2 Description"},
			})
		})
	})
}
