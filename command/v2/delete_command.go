package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type DeleteCommand struct {
	RequiredArgs       flag.AppName `positional-args:"yes"`
	ForceDelete        bool         `short:"f" description:"Force deletion without confirmation"`
	DeleteMappedRoutes bool         `short:"r" description:"Also delete any mapped routes"`
	usage              interface{}  `usage:"CF_NAME delete APP_NAME [-r] [-f]"`
	relatedCommands    interface{}  `related_commands:"apps, scale, stop"`
}

func (_ DeleteCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ DeleteCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
