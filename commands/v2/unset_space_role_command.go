package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type UnsetSpaceRoleCommand struct {
	RequiredArgs    flags.SetSpaceRoleArgs `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME unset-space-role USERNAME ORG SPACE ROLE\n\nROLES:\n   'SpaceManager' - Invite and manage users, and enable features for a given space\n   'SpaceDeveloper' - Create and manage apps and services, and see logs and reports\n   'SpaceAuditor' - View logs, reports, and settings on this space"`
	relatedCommands interface{}            `related_commands:"space-users"`
}

func (_ UnsetSpaceRoleCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ UnsetSpaceRoleCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
