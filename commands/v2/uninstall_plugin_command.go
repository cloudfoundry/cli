package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type UninstallPluginCommand struct {
	RequiredArgs    flags.PluginName `positional-args:"yes"`
	usage           interface{}      `usage:"CF_NAME uninstall-plugin PLUGIN-NAME"`
	relatedCommands interface{}      `related_commands:"plugins"`
}

func (_ UninstallPluginCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ UninstallPluginCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
