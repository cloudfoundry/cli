package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type UnbindRunningSecurityGroupCommand struct {
	RequiredArgs    flag.SecurityGroup `positional-args:"yes"`
	usage           interface{}        `usage:"CF_NAME unbind-running-security-group SECURITY_GROUP\n\nTIP: Changes will not apply to existing running applications until they are restarted."`
	relatedCommands interface{}        `related_commands:"apps, restart, running-security-groups"`
}

func (_ UnbindRunningSecurityGroupCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ UnbindRunningSecurityGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
