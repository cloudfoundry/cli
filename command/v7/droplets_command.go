package v7

import (
	"strings"
	"time"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/clock"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/ui"
)

//go:generate counterfeiter . DropletsActor

type DropletsActor interface {
	GetApplicationDroplets(appName string, spaceGUID string) ([]v7action.Droplet, v7action.Warnings, error)
}

type DropletsCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME droplets APP_NAME"`
	relatedCommands interface{}  `related_commands:"set-droplet, create-package, packages, app, push"`

	UI          command.UI
	Config      command.Config
	Actor       DropletsActor
	SharedActor command.SharedActor
}

func (cmd *DropletsCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd DropletsCommand) Execute(args []string) error {

	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting droplets of app {{.AppName}} in org {{.CurrentOrg}} / space {{.CurrentSpace}} as {{.CurrentUser}}...", map[string]interface{}{
		"AppName":      cmd.RequiredArgs.AppName,
		"CurrentSpace": cmd.Config.TargetedSpace().Name,
		"CurrentOrg":   cmd.Config.TargetedOrganization().Name,
		"CurrentUser":  user.Name,
	})
	cmd.UI.DisplayNewline()

	droplets, warnings, err := cmd.Actor.GetApplicationDroplets(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(droplets) == 0 {
		cmd.UI.DisplayText("No droplets found")
		return nil
	}

	table := [][]string{
		{
			cmd.UI.TranslateText("guid"),
			cmd.UI.TranslateText("state"),
			cmd.UI.TranslateText("created"),
		},
	}

	for _, droplet := range droplets {
		t, err := time.Parse(time.RFC3339, droplet.CreatedAt)
		if err != nil {
			return err
		}

		table = append(table, []string{
			droplet.GUID,
			cmd.UI.TranslateText(strings.ToLower(string(droplet.State))),
			cmd.UI.UserFriendlyDate(t),
		})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}
