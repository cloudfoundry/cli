package commands_test

import (
	"cf"
	. "cf/commands"
	"cf/configuration"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testterm "testhelpers/terminal"
)

func callStacks(t mr.TestingT, stackRepo *testapi.FakeStackRepository) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	ctxt := testcmd.NewContext("stacks", []string{})

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	space := cf.SpaceFields{}
	space.Name = "my-space"

	org := cf.OrganizationFields{}
	org.Name = "my-org"

	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewStacks(ui, config, stackRepo)
	testcmd.RunCommand(cmd, ctxt, nil)

	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestStacks", func() {
			stack1 := cf.Stack{}
			stack1.Name = "Stack-1"
			stack1.Description = "Stack 1 Description"

			stack2 := cf.Stack{}
			stack2.Name = "Stack-2"
			stack2.Description = "Stack 2 Description"

			stackRepo := &testapi.FakeStackRepository{
				FindAllStacks: []cf.Stack{stack1, stack2},
			}

			ui := callStacks(mr.T(), stackRepo)

			assert.Equal(mr.T(), len(ui.Outputs), 6)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting stacks in org", "my-org", "my-space", "my-user"},
				{"OK"},
				{"Stack-1", "Stack 1 Description"},
				{"Stack-2", "Stack 2 Description"},
			})
		})
	})
}
