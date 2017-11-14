package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type OrgUsersCommand struct {
	RequiredArgs    flag.Organization `positional-args:"yes"`
	AllUsers        bool              `short:"a" description:"List all users in the org"`
	usage           interface{}       `usage:"CF_NAME org-users ORG"`
	relatedCommands interface{}       `related_commands:"orgs"`
}

func (OrgUsersCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (OrgUsersCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
