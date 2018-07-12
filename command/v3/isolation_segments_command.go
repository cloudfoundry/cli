package v3

import (
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3/shared"
	"code.cloudfoundry.org/cli/util/ui"
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
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	client, _, err := shared.NewClients(config, ui, true)
	if err != nil {
		if v3Err, ok := err.(ccerror.V3UnexpectedResponseError); ok && v3Err.ResponseCode == http.StatusNotFound {
			return translatableerror.MinimumAPIVersionNotMetError{MinimumVersion: ccversion.MinVersionIsolationSegmentV3}
		}

		return err
	}
	cmd.Actor = v3action.NewActor(client, config, sharedActor, nil)

	return nil
}

func (cmd IsolationSegmentsCommand) Execute(args []string) error {
	err := command.MinimumAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), ccversion.MinVersionIsolationSegmentV3)
	if err != nil {
		return err
	}

	err = cmd.SharedActor.CheckTarget(false, false)
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

	summaries, warnings, err := cmd.Actor.GetIsolationSegmentSummaries()
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
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

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
	return nil
}
