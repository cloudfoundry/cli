package command_runner_test

import (
	"github.com/cloudfoundry/cli/cf/command"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	. "github.com/cloudfoundry/cli/cf/command_runner"
	"github.com/cloudfoundry/cli/cf/requirements"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	"github.com/codegangsta/cli"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type TestCommandFactory struct {
	Cmd     command.Command
	CmdName string
}

func (f *TestCommandFactory) GetByCmdName(cmdName string) (cmd command.Command, err error) {
	f.CmdName = cmdName
	cmd = f.Cmd
	return
}

func (fake *TestCommandFactory) CommandMetadatas() []command_metadata.CommandMetadata {
	return []command_metadata.CommandMetadata{}
}

type TestCommand struct {
	Reqs       []requirements.Requirement
	WasRunWith *cli.Context
}

func (cmd *TestCommand) GetRequirements(_ requirements.Factory, _ *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = cmd.Reqs
	return
}

func (command *TestCommand) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{}
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

var _ = Describe("Requirements runner", func() {
	It("runs", func() {
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

		Expect(cmdFactory.CmdName).To(Equal("some-cmd"))

		Expect(passingReq.WasExecuted).To(BeTrue())
		Expect(failingReq.WasExecuted).To(BeTrue())

		Expect(lastReq.WasExecuted).To(BeFalse())
		Expect(cmd.WasRunWith).To(BeNil())

		Expect(err).To(HaveOccurred())
	})
})
