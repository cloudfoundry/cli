package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type SetQuotaCommand struct {
	RequiredArgs    flag.SetOrgQuotaArgs `positional-args:"yes"`
	usage           interface{}          `usage:"CF_NAME set-quota ORG QUOTA\n\nTIP:\n   View allowable quotas with 'CF_NAME quotas'"`
	relatedCommands interface{}          `related_commands:"orgs, quotas"`
}

func (_ SetQuotaCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ SetQuotaCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
