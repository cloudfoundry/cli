package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type InstallPluginCommand struct {
	OptionalArgs         flags.InstallPluginArgs `positional-args:"yes"`
	RegisteredRepository string                  `short:"r" description:"Name of a registered repository where the specified plugin is located"`
	Force                string                  `short:"f" description:"Force install of plugin without confirmation"`
}

func (_ InstallPluginCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
