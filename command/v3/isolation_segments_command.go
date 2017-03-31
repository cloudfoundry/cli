package v3

import (
	"strings"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . IsolationSegmentsActor

type IsolationSegmentsActor interface {
	CloudControllerAPIVersion() string
	GetIsolationSegmentSummaries() ([]v3action.IsolationSegmentSummary, v3action.Warnings, error)
}

type IsolationSegmentsCommand struct {
	usage           interface{} `usage:"CF_NAME isolation-segments"`
	relatedCommands interface{} `related_commands:"enable-org-isolation, create-isolation-segment"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       IsolationSegmentsActor
}

func (cmd *IsolationSegmentsCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor()

	client, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v3action.NewActor(client, config)

	return nil
}

func (cmd IsolationSegmentsCommand) Execute(args []string) error {
	err := command.MinimumAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), "3.11.0")
	if err != nil {
		return err
	}

	err = cmd.SharedActor.CheckTarget(cmd.Config, false, false)
	if err != nil {
		return shared.HandleError(err)
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting isolation segments as {{.CurrentUser}}...", map[string]interface{}{
		"CurrentUser": user.Name,
	})

	summaries, warnings, err := cmd.Actor.GetIsolationSegmentSummaries()
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}
	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()

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

	cmd.UI.DisplayTableWithHeader("", table, 3)
	return nil
}
