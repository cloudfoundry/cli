package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flags"
)

type CreateUserCommand struct {
	RequiredArgs    flags.Authentication `positional-args:"yes"`
	usage           interface{}          `usage:"CF_NAME create-user USERNAME PASSWORD"`
	relatedCommands interface{}          `related_commands:"passwd, set-org-role, set-space-role"`
}

func (_ CreateUserCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ CreateUserCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
