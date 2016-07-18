package fakecommand

import (
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/flags"
	"github.com/cloudfoundry/cli/cf/requirements"
)

type FakeCommand2 struct {
	Data string
}

func (cmd FakeCommand2) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "fake-command2",
		ShortName:   "fc2",
		Description: "Description for fake-command2",
		Usage: []string{
			"Usage of fake-command2",
		},
	}
}

func (cmd FakeCommand2) Requirements(_ requirements.Factory, _ flags.FlagContext) []requirements.Requirement {
	return []requirements.Requirement{}
}

func (cmd FakeCommand2) SetDependency(deps commandregistry.Dependency, _ bool) commandregistry.Command {
	return cmd
}

func (cmd FakeCommand2) Execute(c flags.FlagContext) error {
	return nil
}
