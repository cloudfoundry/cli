package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/util/ui"
)

type SpacesCommand struct {
	BaseCommand

	usage           interface{} `usage:"CF_NAME spaces [--labels SELECTOR]\n\nEXAMPLES:\n   CF_NAME spaces\n   CF_NAME spaces --labels 'environment in (production,staging),tier in (backend)'\n   CF_NAME spaces --labels 'env=dev,!chargeback-code,tier in (backend,worker)'"`
	relatedCommands interface{} `related_commands:"create-space, set-space-role, space, space-users"`
	Labels          string      `long:"labels" description:"Selector to filter spaces by labels"`
}

func (cmd SpacesCommand) Execute([]string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting spaces in org {{.OrgName}} as {{.CurrentUser}}...", map[string]interface{}{
		"OrgName":     cmd.Config.TargetedOrganization().Name,
		"CurrentUser": user.Name,
	})
	cmd.UI.DisplayNewline()

	spaces, warnings, err := cmd.Actor.GetOrganizationSpacesWithLabelSelector(cmd.Config.TargetedOrganization().GUID, cmd.Labels)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(spaces) == 0 {
		cmd.UI.DisplayText("No spaces found.")
	} else {
		cmd.displaySpaces(spaces)
	}

	return nil
}

func (cmd SpacesCommand) displaySpaces(spaces []v7action.Space) {
	table := [][]string{{cmd.UI.TranslateText("name")}}

	for _, space := range spaces {
		table = append(table, []string{space.Name})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
}
