package v7

import (
	"sort"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/ui"
)

type LabelsCommand struct {
	RequiredArgs flag.LabelsArgs `positional-args:"yes"`
	usage        interface{}     `usage:"CF_NAME labels RESOURCE RESOURCE_NAME\n\nEXAMPLES:\n   cf labels app dora \n\nRESOURCES:\n   APP\n\nSEE ALSO:\n   set-label, delete-label"`
	UI           command.UI
	Config       command.Config
	SharedActor  command.SharedActor
	Actor        AppActor
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

	cmd.UI.DisplayNewline()

	app, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.ResourceName, cmd.Config.TargetedSpace().GUID)
	labels := app.Metadata.Labels

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(labels) == 0 {
		cmd.UI.DisplayText("No labels found.")
		return nil
	}

	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	table := [][]string{
		{
			cmd.UI.TranslateText("Key"),
			cmd.UI.TranslateText("Value"),
		},
	}

	for _, key := range keys {
		table = append(table, []string{key, labels[key].Value})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}
