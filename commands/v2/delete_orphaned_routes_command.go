package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
)

type DeleteOrphanedRoutesCommand struct {
	Force           bool        `short:"f" description:"Force deletion without confirmation"`
	usage           interface{} `usage:"CF_NAME delete-orphaned-routes [-f]"`
	relatedCommands interface{} `related_commands:"delete-route, routes"`
}

func (_ DeleteOrphanedRoutesCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ DeleteOrphanedRoutesCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
