package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type UnbindSecurityGroupCommand struct {
	RequiredArgs    flag.BindSecurityGroupArgs `positional-args:"yes"`
	usage           interface{}                `usage:"CF_NAME unbind-security-group SECURITY_GROUP ORG SPACE\n\nTIP: Changes will not apply to existing running applications until they are restarted."`
	relatedCommands interface{}                `related_commands:"apps, restart, security-groups"`
}

func (_ UnbindSecurityGroupCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ UnbindSecurityGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
