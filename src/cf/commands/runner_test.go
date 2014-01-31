package commands_test

import (
	. "cf/commands"
	"cf/requirements"
	"github.com/codegangsta/cli"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testcmd "testhelpers/commands"
)

type TestCommandFactory struct {
	Cmd     Command
	CmdName string
}

func (f *TestCommandFactory) GetByCmdName(cmdName string) (cmd Command, err error) {
	f.CmdName = cmdName
	cmd = f.Cmd
	return
}

type TestCommand struct {
	Reqs       []requirements.Requirement
	WasRunWith *cli.Context
}

func (cmd *TestCommand) GetRequirements(factory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = cmd.Reqs
	return
}

func (cmd *TestCommand) Run(c *cli.Context) {
	cmd.WasRunWith = c
}

type TestRequirement struct {
	Passes      bool
	WasExecuted bool
}

func (r *TestRequirement) Execute() (success bool) {
	r.WasExecuted = true

	if !r.Passes {
		return false
	}

	return true
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestRun", func() {

			passingReq := TestRequirement{Passes: true}
			failingReq := TestRequirement{Passes: false}
			lastReq := TestRequirement{Passes: true}

			cmd := TestCommand{
				Reqs: []requirements.Requirement{&passingReq, &failingReq, &lastReq},
			}

			cmdFactory := &TestCommandFactory{Cmd: &cmd}
			runner := NewRunner(cmdFactory, nil)

			ctxt := testcmd.NewContext("login", []string{})

			err := runner.RunCmdByName("some-cmd", ctxt)

			assert.Equal(mr.T(), cmdFactory.CmdName, "some-cmd")

			assert.True(mr.T(), passingReq.WasExecuted, ctxt)
			assert.True(mr.T(), failingReq.WasExecuted, ctxt)

			assert.False(mr.T(), lastReq.WasExecuted)
			assert.Nil(mr.T(), cmd.WasRunWith)

			assert.Error(mr.T(), err)
		})
	})
}
