package fakecommand

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/requirements"
)

var ErrFakeCommand4 = errors.New("ZOMG command errored")

type FakeCommand4 struct {
	Data string
}

func init() {
	commandregistry.Register(FakeCommand4{Data: "FakeCommand4 data"})
}

func (cmd FakeCommand4) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "fake-command4",
		Description: "Description for fake-command4 will error on run",
		Usage: []string{
			"Usage of fake-command4",
		},
	}
}

func (cmd FakeCommand4) Requirements(_ requirements.Factory, _ flags.FlagContext) ([]requirements.Requirement, error) {
	reqs := []requirements.Requirement{}
	return reqs, nil
}

func (cmd FakeCommand4) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	return cmd
}

func (cmd FakeCommand4) Execute(c flags.FlagContext) error {
	return ErrFakeCommand4
}
