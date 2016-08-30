package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
)

type StacksCommand struct {
	usage           interface{} `usage:"CF_NAME stacks"`
	relatedCommands interface{} `related_commands:"app, push"`
}

func (_ StacksCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ StacksCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
