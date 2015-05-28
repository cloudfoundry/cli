package fake_command

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type FakeCommand2 struct {
	Data string
	req  fakeReq2
	ui   terminal.UI
}

func init() {
	command_registry.Register(FakeCommand2{Data: "FakeCommand2 data", req: fakeReq2{}})
}

func (cmd FakeCommand2) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "fake-non-codegangsta-command2",
		Description: "Description for fake-command2 with bad requirement",
		Usage:       "Usage of fake-command",
	}
}

func (cmd FakeCommand2) Requirements(_ requirements.Factory, _ flags.FlagContext) (reqs []requirements.Requirement, err error) {
	return []requirements.Requirement{cmd.req}, nil
}

func (cmd FakeCommand2) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.req.ui = deps.Ui
	cmd.ui = deps.Ui
	cmd.ui.Say("SetDependency() called, pluginCall " + fmt.Sprintf("%t", pluginCall))

	return cmd
}

func (cmd FakeCommand2) Execute(c flags.FlagContext) {
	cmd.ui.Say("Command Executed")
}

type fakeReq2 struct {
	ui terminal.UI
}

func (f fakeReq2) Execute() bool {
	f.ui.Say("Requirement executed and failed")
	return false
}
