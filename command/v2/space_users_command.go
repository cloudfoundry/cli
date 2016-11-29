package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type SpaceUsersCommand struct {
	RequiredArgs    flag.OrgSpace `positional-args:"yes"`
	usage           interface{}   `usage:"CF_NAME space-users ORG SPACE"`
	relatedCommands interface{}   `related_commands:"org-users, set-space-role, unset-space-role, orgs, spaces"`
}

func (_ SpaceUsersCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ SpaceUsersCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
