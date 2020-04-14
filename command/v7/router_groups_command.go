package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/util/ui"
)

type RouterGroupsCommand struct {
	BaseCommand

	usage           interface{} `usage:"CF_NAME router-groups"`
	relatedCommands interface{} `related_commands:"create-domain, domains"`
}

func (cmd RouterGroupsCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	currentUser, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting router groups as {{.CurrentUser}}...", map[string]interface{}{
		"CurrentUser": currentUser.Name,
	})

	cmd.UI.DisplayNewline()

	routerGroups, warnings, err := cmd.Actor.GetRouterGroups()
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(routerGroups) == 0 {
		cmd.UI.DisplayText("No router groups found.")
	} else {
		cmd.displayRouterGroupsTable(routerGroups)
	}

	return nil
}

func (cmd RouterGroupsCommand) displayRouterGroupsTable(routerGroups []v7action.RouterGroup) {
	var table = [][]string{
		{
			cmd.UI.TranslateText("name"),
			cmd.UI.TranslateText("type"),
		},
	}

	for _, routerGroup := range routerGroups {
		table = append(table, []string{
			routerGroup.Name,
			routerGroup.Type,
		})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
}
