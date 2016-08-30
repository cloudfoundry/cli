package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
)

type ServicesCommand struct {
	usage           interface{} `usage:"CF_NAME services"`
	relatedCommands interface{} `related_commands:"create-service, marketplace"`
}

func (_ ServicesCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ ServicesCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
