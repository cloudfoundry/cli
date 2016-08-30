package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type DeleteServiceCommand struct {
	RequiredArgs    flags.ServiceInstance `positional-args:"yes"`
	Force           bool                  `short:"f" description:"Force deletion without confirmation"`
	usage           interface{}           `usage:"CF_NAME delete-service SERVICE_INSTANCE [-f]"`
	relatedCommands interface{}           `related_commands:"unbind-service, services"`
}

func (_ DeleteServiceCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ DeleteServiceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
