package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
)

type PluginsCommand struct {
	Checksum        bool        `long:"checksum" description:"Compute and show the sha1 value of the plugin binary file"`
	usage           interface{} `usage:"CF_NAME plugins [--checksum]"`
	relatedCommands interface{} `related_commands:"install-plugin, repo-plugins, uninstall-plugin"`
}

func (_ PluginsCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ PluginsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
