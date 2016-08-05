package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type LogsCommand struct {
	RequiredArgs flags.AppName `positional-args:"yes"`
	Recent       bool          `long:"recent" description:"Dump recent logs instead of tailing"`
}

func (_ LogsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
