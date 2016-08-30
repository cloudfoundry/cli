package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type BindSecurityGroupCommand struct {
	RequiredArgs    flags.BindSecurityGroupArgs `positional-args:"yes"`
	usage           interface{}                 `usage:"CF_NAME bind-security-group SECURITY_GROUP ORG [SPACE]\n\nTIP: Changes will not apply to existing running applications until they are restarted."`
	relatedCommands interface{}                 `related_commands:"apps, bind-running-security-group, bind-staging-security-group, restart, security-groups"`
}

func (_ BindSecurityGroupCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ BindSecurityGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
