package v3

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . DeleteIsolationSegmentActor

type DeleteIsolationSegmentActor interface {
	CloudControllerAPIVersion() string
	DeleteIsolationSegmentByName(name string) (v3action.Warnings, error)
}

type DeleteIsolationSegmentCommand struct {
	RequiredArgs    flag.IsolationSegmentName `positional-args:"yes"`
	Force           bool                      `short:"f" description:"Force deletion without confirmation"`
	usage           interface{}               `usage:"CF_NAME delete-isolation-segment SEGMENT_NAME"`
	relatedCommands interface{}               `related_commands:"disable-org-isolation, isolation-segments"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       DeleteIsolationSegmentActor
}

func (cmd *DeleteIsolationSegmentCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd DeleteIsolationSegmentCommand) Execute(args []string) error {
	err := command.MinimumAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), "3.11.0")
	if err != nil {
		return err
	}

	err = cmd.SharedActor.CheckTarget(cmd.Config, false, false)
	if err != nil {
		return shared.HandleError(err)
	}

	if !cmd.Force {
		deleteSegment, promptErr := cmd.UI.DisplayBoolPrompt(false, "Really delete the isolation segment {{.IsolationSegmentName}}?", map[string]interface{}{
			"IsolationSegmentName": cmd.RequiredArgs.IsolationSegmentName,
		})

		if promptErr != nil {
			return promptErr
		}

		if !deleteSegment {
			cmd.UI.DisplayText("Delete cancelled")
			return nil
		}
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Deleting isolation segment {{.SegmentName}} as {{.CurrentUser}}...", map[string]interface{}{
		"SegmentName": cmd.RequiredArgs.IsolationSegmentName,
		"CurrentUser": user.Name,
	})

	warnings, err := cmd.Actor.DeleteIsolationSegmentByName(cmd.RequiredArgs.IsolationSegmentName)
	cmd.UI.DisplayWarnings(warnings)
	if _, ok := err.(v3action.IsolationSegmentNotFoundError); ok {
		cmd.UI.DisplayWarning("Isolation segment {{.IsolationSegmentName}} does not exist.", map[string]interface{}{
			"IsolationSegmentName": cmd.RequiredArgs.IsolationSegmentName,
		})
	} else if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
