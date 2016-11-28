package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
)

type SpacesCommand struct {
	usage           interface{} `usage:"CF_NAME spaces"`
	relatedCommands interface{} `related_commands:"target"`
}

func (_ SpacesCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ SpacesCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
