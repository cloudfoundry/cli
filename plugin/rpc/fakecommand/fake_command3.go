package fakecommand

import (
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/requirements"
)

type FakeCommand3 struct {
	Data string
}

func init() {
	commandregistry.Register(FakeCommand3{Data: "FakeCommand3 data"})
}

func (cmd FakeCommand3) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "fake-command3",
		Description: "Description for fake-command3",
		Usage: []string{
			"Usage of fake-command3",
		},
	}
}

func (cmd FakeCommand3) Requirements(_ requirements.Factory, _ flags.FlagContext) ([]requirements.Requirement, error) {
	reqs := []requirements.Requirement{}
	return reqs, nil
}

func (cmd FakeCommand3) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	return cmd
}

func (cmd FakeCommand3) Execute(c flags.FlagContext) error {
	panic("this is a test panic for cli_rpc_server_test (panic recovery)")
}
