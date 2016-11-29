package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type AuthCommand struct {
	RequiredArgs    flag.Authentication `positional-args:"yes"`
	usage           interface{}         `usage:"CF_NAME auth USERNAME PASSWORD\n\nWARNING:\n   Providing your password as a command line option is highly discouraged\n   Your password may be visible to others and may be recorded in your shell history\n\nEXAMPLES:\n   CF_NAME auth name@example.com \"my password\" (use quotes for passwords with a space)\n   CF_NAME auth name@example.com \"\\\"password\\\"\" (escape quotes if used in password)"`
	relatedCommands interface{}         `related_commands:"api, login, target"`
}

func (_ AuthCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ AuthCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
