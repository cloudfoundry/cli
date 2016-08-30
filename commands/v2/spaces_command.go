package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
)

type SpacesCommand struct {
	usage           interface{} `usage:"CF_NAME spaces"`
	relatedCommands interface{} `related_commands:"target"`
}

func (_ SpacesCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ SpacesCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
