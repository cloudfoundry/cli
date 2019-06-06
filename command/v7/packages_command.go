package v7

import (
	"strings"
	"time"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/ui"
)

//go:generate counterfeiter . PackagesActor

type PackagesActor interface {
	GetApplicationPackages(appName string, spaceGUID string) ([]v7action.Package, v7action.Warnings, error)
}

type PackagesCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME packages APP_NAME"`
	relatedCommands interface{}  `related_commands:"droplets, create-package, app, push"`

	UI          command.UI
	Config      command.Config
	Actor       PackagesActor
	SharedActor command.SharedActor
}

func (cmd *PackagesCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, _, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil)

	return nil
}

func (cmd PackagesCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting packages of app {{.AppName}} in org {{.CurrentOrg}} / space {{.CurrentSpace}} as {{.CurrentUser}}...", map[string]interface{}{
		"AppName":      cmd.RequiredArgs.AppName,
		"CurrentSpace": cmd.Config.TargetedSpace().Name,
		"CurrentOrg":   cmd.Config.TargetedOrganization().Name,
		"CurrentUser":  user.Name,
	})
	cmd.UI.DisplayNewline()

	packages, warnings, err := cmd.Actor.GetApplicationPackages(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(packages) == 0 {
		cmd.UI.DisplayText("No packages found.")
		return nil
	}

	contents := [][]string{}

	for _, pkg := range packages {
		t, err := time.Parse(time.RFC3339, pkg.CreatedAt)
		if err != nil {
			return err
		}

		contents = append(contents, []string{
			pkg.GUID,
			cmd.UI.TranslateText(strings.ToLower(string(pkg.State))),
			cmd.UI.UserFriendlyDate(t),
		})
	}

	table := [][]string{
		{
			cmd.UI.TranslateText("guid"),
			cmd.UI.TranslateText("state"),
			cmd.UI.TranslateText("created"),
		},
	}

	for i := len(contents) - 1; i >= 0; i-- {
		table = append(table, contents[i])
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}
