package v7

import (
	"code.cloudfoundry.org/cli/v8/command/flag"
	"code.cloudfoundry.org/cli/v8/resources"
)

type StackCommand struct {
	BaseCommand

	RequiredArgs    flag.StackName `positional-args:"yes"`
	GUID            bool           `long:"guid" description:"Retrieve and display the given stack's guid. All other output for the stack is suppressed."`
	usage           interface{}    `usage:"CF_NAME stack STACK_NAME"`
	relatedCommands interface{}    `related_commands:"app, push, stacks, update-stack"`
}

func (cmd *StackCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	if cmd.GUID {
		return cmd.displayStackGUID()
	}

	return cmd.displayStackInfo()
}

func (cmd *StackCommand) getStack(stackName string) (resources.Stack, error) {
	stack, warnings, err := cmd.Actor.GetStackByName(cmd.RequiredArgs.StackName)
	cmd.UI.DisplayWarnings(warnings)
	return stack, err
}

func (cmd *StackCommand) displayStackGUID() error {
	stack, err := cmd.getStack(cmd.RequiredArgs.StackName)
	if err != nil {
		return err
	}

	cmd.UI.DisplayText(stack.GUID)
	return nil
}

func (cmd *StackCommand) displayStackInfo() error {
	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting info for stack {{.StackName}} as {{.Username}}...", map[string]interface{}{
		"StackName": cmd.RequiredArgs.StackName,
		"Username":  user.Name,
	})
	cmd.UI.DisplayNewline()

	stack, err := cmd.getStack(cmd.RequiredArgs.StackName)
	if err != nil {
		return err
	}

	// Build display table
	displayTable := [][]string{
		{cmd.UI.TranslateText("name:"), stack.Name},
		{cmd.UI.TranslateText("description:"), stack.Description},
	}

	// Add state only if it's present
	if stack.State != "" {
		displayTable = append(displayTable, []string{cmd.UI.TranslateText("state:"), stack.State})

		// Add reason whenever state is not ACTIVE
		if stack.State != resources.StackStateActive {
			displayTable = append(displayTable, []string{cmd.UI.TranslateText("reason:"), stack.StateReason})
		}
	}

	cmd.UI.DisplayKeyValueTable("", displayTable, 3)
	return nil
}
