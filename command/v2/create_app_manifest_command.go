package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type CreateAppManifestCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	FilePath        string       `short:"p" description:"Specify a path for file creation. If path not specified, manifest file is created in current working directory."`
	usage           interface{}  `usage:"CF_NAME create-app-manifest APP_NAME [-p /path/to/<app-name>-manifest.yml]"`
	relatedCommands interface{}  `related_commands:"apps, push"`
}

func (_ CreateAppManifestCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ CreateAppManifestCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
