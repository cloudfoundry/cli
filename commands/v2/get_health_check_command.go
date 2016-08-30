package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type GetHealthCheckCommand struct {
	RequiredArgs flags.AppName `positional-args:"yes"`
	usage        interface{}   `usage:"CF_NAME get-health-check APP_NAME"`
}

func (_ GetHealthCheckCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ GetHealthCheckCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
