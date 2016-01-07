package fake_command

import (
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/flags"
)

type FakeCommand2 struct {
	Data string
}

func (cmd FakeCommand2) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "fake-command2",
		ShortName:   "fc2",
		Description: "Description for fake-command2",
		Usage:       "Usage of fake-command2",
	}
}

func (cmd FakeCommand2) Requirements(_ requirements.Factory, _ flags.FlagContext) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd FakeCommand2) SetDependency(deps command_registry.Dependency, _ bool) command_registry.Command {
	return cmd
}

func (cmd FakeCommand2) Execute(c flags.FlagContext) {
}
