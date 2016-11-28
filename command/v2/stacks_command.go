package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
)

type StacksCommand struct {
	usage           interface{} `usage:"CF_NAME stacks"`
	relatedCommands interface{} `related_commands:"app, push"`
}

func (_ StacksCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ StacksCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
