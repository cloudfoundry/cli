package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type OrgUsersCommand struct {
	RequiredArgs    flag.Organization `positional-args:"yes"`
	AllUsers        bool              `short:"a" description:"List all users in the org"`
	usage           interface{}       `usage:"CF_NAME org-users ORG"`
	relatedCommands interface{}       `related_commands:"orgs"`
}

func (_ OrgUsersCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ OrgUsersCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
