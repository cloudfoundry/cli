package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type UnbindStagingSecurityGroupCommand struct {
	RequiredArgs    flag.SecurityGroup `positional-args:"yes"`
	usage           interface{}        `usage:"CF_NAME unbind-staging-security-group SECURITY_GROUP\n\nTIP: Changes will not apply to existing running applications until they are restarted."`
	relatedCommands interface{}        `related_commands:"apps, restart, staging-security-groups"`
}

func (_ UnbindStagingSecurityGroupCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ UnbindStagingSecurityGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
