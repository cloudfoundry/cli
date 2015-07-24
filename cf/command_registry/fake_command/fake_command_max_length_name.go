package fake_command

import (
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/flags"
)

type FakeCommand3 struct {
}

func init() {
	command_registry.Register(FakeCommand3{})
}

func (cmd FakeCommand3) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name: "this-is-a-really-long-command-name-123123123123123123123",
	}
}

func (cmd FakeCommand3) Requirements(_ requirements.Factory, _ flags.FlagContext) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd FakeCommand3) SetDependency(deps command_registry.Dependency, _ bool) command_registry.Command {
	return cmd
}

func (cmd FakeCommand3) Execute(c flags.FlagContext) {
}
