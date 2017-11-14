package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type UnsetOrgRoleCommand struct {
	RequiredArgs    flag.SetOrgRoleArgs `positional-args:"yes"`
	usage           interface{}         `usage:"CF_NAME unset-org-role USERNAME ORG ROLE\n\nROLES:\n   'OrgManager' - Invite and manage users, select and change plans, and set spending limits\n   'BillingManager' - Create and manage the billing account and payment info\n   'OrgAuditor' - Read-only access to org info and reports"`
	relatedCommands interface{}         `related_commands:"org-users, delete-user"`
}

func (UnsetOrgRoleCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (UnsetOrgRoleCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
