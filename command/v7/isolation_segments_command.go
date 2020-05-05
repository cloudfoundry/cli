package v7

import (
	"strings"

	"code.cloudfoundry.org/cli/util/ui"
)

type IsolationSegmentsCommand struct {
	BaseCommand
	usage           interface{} `usage:"CF_NAME isolation-segments"`
	relatedCommands interface{} `related_commands:"enable-org-isolation, create-isolation-segment"`
}

func (cmd IsolationSegmentsCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting isolation segments as {{.CurrentUser}}...", map[string]interface{}{
		"CurrentUser": user.Name,
	})

	cmd.UI.DisplayNewline()

	summaries, warnings, err := cmd.Actor.GetIsolationSegmentSummaries()
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	table := [][]string{
		{
			cmd.UI.TranslateText("name"),
			cmd.UI.TranslateText("orgs"),
		},
	}

	for _, summary := range summaries {
		table = append(
			table,
			[]string{
				summary.Name,
				strings.Join(summary.EntitledOrgs, ", "),
			},
		)
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
	return nil
}
