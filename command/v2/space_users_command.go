package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type SpaceUsersCommand struct {
	RequiredArgs    flag.OrgSpace `positional-args:"yes"`
	usage           interface{}   `usage:"CF_NAME space-users ORG SPACE"`
	relatedCommands interface{}   `related_commands:"org-users, set-space-role, unset-space-role, orgs, spaces"`
}

func (SpaceUsersCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (SpaceUsersCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
