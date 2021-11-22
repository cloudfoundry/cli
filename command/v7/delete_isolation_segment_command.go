package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/flag"
)

type DeleteIsolationSegmentCommand struct {
	BaseCommand

	RequiredArgs    flag.IsolationSegmentName `positional-args:"yes"`
	Force           bool                      `short:"f" description:"Force deletion without confirmation"`
	usage           interface{}               `usage:"CF_NAME delete-isolation-segment SEGMENT_NAME"`
	relatedCommands interface{}               `related_commands:"disable-org-isolation, isolation-segments"`
}

func (cmd DeleteIsolationSegmentCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
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

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Deleting isolation segment {{.SegmentName}} as {{.CurrentUser}}...", map[string]interface{}{
		"SegmentName": cmd.RequiredArgs.IsolationSegmentName,
		"CurrentUser": user.Name,
	})

	warnings, err := cmd.Actor.DeleteIsolationSegmentByName(cmd.RequiredArgs.IsolationSegmentName)
	cmd.UI.DisplayWarnings(warnings)
	if _, ok := err.(actionerror.IsolationSegmentNotFoundError); ok {
		cmd.UI.DisplayWarning("Isolation segment {{.IsolationSegmentName}} does not exist.", map[string]interface{}{
			"IsolationSegmentName": cmd.RequiredArgs.IsolationSegmentName,
		})
	} else if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
