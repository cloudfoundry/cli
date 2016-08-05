package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type FilesCommand struct {
	RequiredArgs flags.FilesArgs `positional-args:"yes"`
	Instance     int             `short:"i" description:"Instance"`
}

func (_ FilesCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
