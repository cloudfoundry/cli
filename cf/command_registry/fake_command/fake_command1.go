package fake_command

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
)

type FakeCommand1 struct {
	Data string
}

func init() {
	command_registry.Register(FakeCommand1{Data: "FakeCommand1 data"})
}

func (cmd FakeCommand1) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &cliFlags.BoolFlag{ShortName: "f", Usage: "Usage for BoolFlag"}
	fs["boolFlag"] = &cliFlags.BoolFlag{Name: "BoolFlag", Usage: "Usage for BoolFlag"}
	fs["intFlag"] = &cliFlags.IntFlag{Name: "intFlag", Usage: "Usage for intFlag"}

	return command_registry.CommandMetadata{
		Name:        "fake-command",
		ShortName:   "fc1",
		Description: "Description for fake-command",
		Usage:       "CF_NAME Usage of fake-command",
		Flags:       fs,
	}
}

func (cmd FakeCommand1) Requirements(_ requirements.Factory, _ flags.FlagContext) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd FakeCommand1) SetDependency(deps command_registry.Dependency, _ bool) command_registry.Command {
	return cmd
}

func (cmd FakeCommand1) Execute(c flags.FlagContext) {
	fmt.Println("This is fake-command")
}
