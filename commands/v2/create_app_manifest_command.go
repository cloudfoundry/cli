package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type CreateAppManifestCommand struct {
	RequiredArgs flags.AppName `positional-args:"yes"`
	FilePath     string        `short:"p" description:"Specify a path for file creation. If path not specified, manifest file is created in current working directory."`
}

func (_ CreateAppManifestCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
