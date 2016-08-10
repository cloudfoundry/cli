package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type DeleteUserCommand struct {
	RequiredArgs flags.Username `positional-args:"yes"`
	Force        bool           `short:"f" description:"Force deletion without confirmation"`
	usage        interface{}    `usage:"CF_NAME delete-user USERNAME [-f]"`
}

func (_ DeleteUserCommand) Setup() error {
	return nil
}

func (_ DeleteUserCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
