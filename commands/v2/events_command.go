package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type EventsCommand struct {
	RequiredArgs flags.AppName `positional-args:"yes"`
	usage        interface{}   `usage:"CF_NAME events APP_NAME"`
}

func (_ EventsCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ EventsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
