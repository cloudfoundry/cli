package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type BindSecurityGroupCommand struct {
	RequiredArgs    flag.BindSecurityGroupArgs `positional-args:"yes"`
	usage           interface{}                `usage:"CF_NAME bind-security-group SECURITY_GROUP ORG [SPACE]\n\nTIP: Changes will not apply to existing running applications until they are restarted."`
	relatedCommands interface{}                `related_commands:"apps, bind-running-security-group, bind-staging-security-group, restart, security-groups"`
}

func (_ BindSecurityGroupCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ BindSecurityGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
