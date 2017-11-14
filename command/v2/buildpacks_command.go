package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type BuildpacksCommand struct {
	usage           interface{} `usage:"CF_NAME buildpacks"`
	relatedCommands interface{} `related_commands:"push"`
}

func (BuildpacksCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (BuildpacksCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
