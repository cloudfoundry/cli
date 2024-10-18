package v7

import (
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/util/ui"
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

	user, err := cmd.Actor.GetCurrentUser()
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

func (cmd SpacesCommand) displaySpaces(spaces []resources.Space) {
	table := [][]string{{cmd.UI.TranslateText("name")}}

	for _, space := range spaces {
		table = append(table, []string{space.Name})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
}
