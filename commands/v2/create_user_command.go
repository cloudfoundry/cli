package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type CreateUserCommand struct {
	RequiredArgs flags.Authentication `positional-args:"yes"`
	usage        interface{}          `usage:"CF_NAME create-user USERNAME PASSWORD"`
}

func (_ CreateUserCommand) Setup() error {
	return nil
}

func (_ CreateUserCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
