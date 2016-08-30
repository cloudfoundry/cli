package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type FilesCommand struct {
	RequiredArgs    flags.FilesArgs `positional-args:"yes"`
	Instance        int             `short:"i" description:"Instance"`
	usage           interface{}     `usage:"CF_NAME files APP_NAME [PATH] [-i INSTANCE]\n\nTIP:\n   To list and inspect files of an app running on the Diego backend, use 'CF_NAME ssh'"`
	relatedCommands interface{}     `related_commands:"ssh"`
}

func (_ FilesCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ FilesCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
