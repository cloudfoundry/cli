package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/flag"
)

type CreateIsolationSegmentCommand struct {
	command.BaseCommand
	RequiredArgs    flag.IsolationSegmentName `positional-args:"yes"`
	usage           interface{}               `usage:"CF_NAME create-isolation-segment SEGMENT_NAME\n\nNOTES:\n   The isolation segment name must match the placement tag applied to the Diego cell."`
	relatedCommands interface{}               `related_commands:"enable-org-isolation, isolation-segments"`
}

func (cmd CreateIsolationSegmentCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Creating isolation segment {{.SegmentName}} as {{.CurrentUser}}...", map[string]interface{}{
		"SegmentName": cmd.RequiredArgs.IsolationSegmentName,
		"CurrentUser": user.Name,
	})

	warnings, err := cmd.Actor.CreateIsolationSegmentByName(v7action.IsolationSegment{
		Name: cmd.RequiredArgs.IsolationSegmentName,
	})
	cmd.UI.DisplayWarnings(warnings)
	if _, ok := err.(actionerror.IsolationSegmentAlreadyExistsError); ok {
		cmd.UI.DisplayWarning("Isolation segment '{{.IsolationSegmentName}}' already exists.", map[string]interface{}{
			"IsolationSegmentName": cmd.RequiredArgs.IsolationSegmentName,
		})
	} else if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
