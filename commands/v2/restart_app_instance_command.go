package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type RestartAppInstanceCommand struct {
	RequiredArgs flags.AppInstance `positional-args:"yes"`
}

func (_ RestartAppInstanceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
