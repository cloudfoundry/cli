package v7

import (
	"fmt"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/util/ui"
)

type DropletsCommand struct {
	BaseCommand

	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME droplets APP_NAME"`
	relatedCommands interface{}  `related_commands:"set-droplet, create-package, packages, app, push"`
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

		dropletGUID := droplet.GUID
		if droplet.IsCurrent {
			dropletGUID = fmt.Sprintf(`%s (current)`, dropletGUID)
		}

		table = append(table, []string{
			dropletGUID,
			cmd.UI.TranslateText(strings.ToLower(string(droplet.State))),
			cmd.UI.UserFriendlyDate(t),
		})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}
