package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type UnsetOrgRoleCommand struct {
	RequiredArgs    flag.SetOrgRoleArgs `positional-args:"yes"`
	usage           interface{}         `usage:"CF_NAME unset-org-role USERNAME ORG ROLE\n\nROLES:\n   'OrgManager' - Invite and manage users, select and change plans, and set spending limits\n   'BillingManager' - Create and manage the billing account and payment info\n   'OrgAuditor' - Read-only access to org info and reports"`
	relatedCommands interface{}         `related_commands:"org-users, delete-user"`
}

func (_ UnsetOrgRoleCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ UnsetOrgRoleCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
