package fakecommand

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type FakeCommand1 struct {
	Data string
	req  fakeReq
	ui   terminal.UI
}

func init() {
	commandregistry.Register(FakeCommand1{Data: "FakeCommand1 data", req: fakeReq{ui: nil}})
}

func (cmd FakeCommand1) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "fake-command",
		Description: "Description for fake-command",
		Usage: []string{
			"Usage of fake-command",
		},
	}
}

func (cmd FakeCommand1) Requirements(_ requirements.Factory, _ flags.FlagContext) []requirements.Requirement {
	reqs := []requirements.Requirement{cmd.req}
	return reqs
}

func (cmd FakeCommand1) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	if cmd.ui != nil {
		cmd.ui.Say("SetDependency() called, pluginCall " + fmt.Sprintf("%t", pluginCall))
	}

	cmd.req.ui = deps.UI
	cmd.ui = deps.UI

	return cmd
}

func (cmd FakeCommand1) Execute(c flags.FlagContext) error {
	cmd.ui.Say("Command Executed")
	return nil
}

type fakeReq struct {
	ui terminal.UI
}

func (f fakeReq) Execute() error {
	f.ui.Say("Requirement executed")
	return nil
}
