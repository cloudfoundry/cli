package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type CreateSpaceCommand struct {
	RequiredArgs flags.Space `positional-args:"yes"`
	Organization string      `short:"o" description:"Organization"`
	Quota        string      `short:"q" description:"Quota to assign to the newly created space"`
}

func (_ CreateSpaceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
