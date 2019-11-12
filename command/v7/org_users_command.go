package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . OrgUsersActor

type OrgUsersCommand struct {
	RequiredArgs    flag.Organization `positional-args:"yes"`
	AllUsers        bool              `long:"all-users" short:"a" description:"List all users in the org including Org Users"`
	usage           interface{}       `usage:"CF_NAME org-users ORG"`
	relatedCommands interface{}       `related_commands:"orgs, set-org-role"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       EnvActor
}

func (cmd *OrgUsersCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, _, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil, clock.NewClock())

	return nil
}

func (cmd *OrgUsersCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting users in {{.Org}} as {{.CurrentUser}}...", map[string]interface{}{
		"Org":         cmd.Config.TargetedOrganization(),
		"CurrentUser": user.Name,
	})
	cmd.UI.DisplayNewline()

	return translatableerror.UnrefactoredCommandError{}
}
