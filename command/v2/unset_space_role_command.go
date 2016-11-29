package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type UnsetSpaceRoleCommand struct {
	RequiredArgs    flag.SetSpaceRoleArgs `positional-args:"yes"`
	usage           interface{}           `usage:"CF_NAME unset-space-role USERNAME ORG SPACE ROLE\n\nROLES:\n   'SpaceManager' - Invite and manage users, and enable features for a given space\n   'SpaceDeveloper' - Create and manage apps and services, and see logs and reports\n   'SpaceAuditor' - View logs, reports, and settings on this space"`
	relatedCommands interface{}           `related_commands:"space-users"`
}

func (_ UnsetSpaceRoleCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ UnsetSpaceRoleCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
