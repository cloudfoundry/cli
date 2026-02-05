package v7

import (
	"slices"
	"strings"

	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/v8/command"
	"code.cloudfoundry.org/cli/v8/command/flag"
	"code.cloudfoundry.org/cli/v8/resources"
)

type UpdateStackCommand struct {
	BaseCommand

	RequiredArgs    flag.StackName `positional-args:"yes"`
	State           string         `long:"state" description:"State to transition the stack to (active, restricted, deprecated, disabled)" required:"true"`
	Reason          string         `long:"reason" description:"Optional plain text describing the stack state change"`
	usage           interface{}    `usage:"CF_NAME update-stack STACK_NAME [--state (active | restricted | deprecated | disabled)] [--reason REASON]\n\nEXAMPLES:\n   CF_NAME update-stack cflinuxfs3 --state disabled\n   CF_NAME update-stack cflinuxfs3 --state deprecated --reason 'Use cflinuxfs4 instead'"`
	relatedCommands interface{}    `related_commands:"stack, stacks"`
}

func (cmd UpdateStackCommand) Execute(args []string) error {
	err := command.MinimumCCAPIVersionCheck(cmd.Config.APIVersion(), ccversion.MinVersionUpdateStack)
	if err != nil {
		return err
	}

	err = cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	// Validate and capitalize the state
	stateValue := strings.ToUpper(cmd.State)

	// Validate against known states
	if !slices.Contains(resources.ValidStackStates, stateValue) {
		return invalidStackStateError{State: cmd.State}
	}

	cmd.UI.DisplayTextWithFlavor("Updating stack {{.StackName}} as {{.Username}}...", map[string]interface{}{
		"StackName": cmd.RequiredArgs.StackName,
		"Username":  user.Name,
	})

	// Get the stack to find its GUID
	stack, warnings, err := cmd.Actor.GetStackByName(cmd.RequiredArgs.StackName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	// Update the stack
	updatedStack, warnings, err := cmd.Actor.UpdateStack(stack.GUID, stateValue, cmd.Reason)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()

	// Display the updated stack info
	displayTable := [][]string{
		{cmd.UI.TranslateText("name:"), updatedStack.Name},
		{cmd.UI.TranslateText("description:"), updatedStack.Description},
		{cmd.UI.TranslateText("state:"), updatedStack.State},
	}

	// Add reason if it's present
	if updatedStack.StateReason != "" {
		displayTable = append(displayTable, []string{cmd.UI.TranslateText("reason:"), updatedStack.StateReason})
	}

	cmd.UI.DisplayKeyValueTable("", displayTable, 3)

	return nil
}

type invalidStackStateError struct {
	State string
}

func (e invalidStackStateError) Error() string {
	return "Invalid state: " + e.State + ". Must be one of: " + strings.Join(resources.ValidStackStatesLowercase(), ", ")
}
