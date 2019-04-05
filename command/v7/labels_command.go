package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/ui"
)

//go:generate counterfeiter . GetLabelsActor
type GetLabelsActor interface {
	GetApplicationLabels(string, string) (map[string]types.NullString, v7action.Warnings, error)
}

type LabelsCommand struct {
	RequiredArgs flag.LabelsArgs `positional-args:"yes"`
	usage        interface{}     `usage:"CF_NAME labels RESOURCE RESOURCE_NAME\n\n EXAMPLES:\n   cf labels app dora \n\nRESOURCES:\n   APP\n\nSEE ALSO:\n   set-label, delete-label"`
	UI           command.UI
	Config       command.Config
	SharedActor  command.SharedActor
	Actor        GetLabelsActor
}

func (cmd *LabelsCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd LabelsCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	username, err := cmd.Config.CurrentUserName()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting labels for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.ResourceName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  username,
	})

	labels, warnings, err := cmd.Actor.GetApplicationLabels(cmd.RequiredArgs.ResourceName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()

	if len(labels) == 0 {
		cmd.UI.DisplayText("No labels found.")
		return nil
	}

	table := [][]string{
		{
			cmd.UI.TranslateText("Key"),
			cmd.UI.TranslateText("Value"),
		},
	}

	for key, value := range labels {
		table = append(table, []string{key, value.Value})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}
