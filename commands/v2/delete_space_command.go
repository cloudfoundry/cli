package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type DeleteSpaceCommand struct {
	RequiredArgs flags.Space `positional-args:"yes"`
	Force        bool        `short:"f" description:"Force deletion without confirmation"`
	usage        interface{} `usage:"CF_NAME delete-space SPACE [-f]"`
}

func (_ DeleteSpaceCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ DeleteSpaceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
