package v7

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type RevisionsCommand struct {
	RequiredArgs flag.AppName `positional-args:"yes"`

	//OptionalArgs flag.RevisionNumber `positional-args:"yes"`

	UI     command.UI
	Config command.Config
}

func (cmd *RevisionsCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui

	return nil
}

func (cmd *RevisionsCommand) Execute(args []string) error {

	//	if cmd.OptionalArgs.RevisionNumber != "" {
	//	cmd.UI.DisplayText("blah")
	//}
	cmd.displayTargetTable()

	return nil
}

// displayTargetTable neatly displays target information.
func (cmd *RevisionsCommand) displayTargetTable() {
	cmd.UI.DisplayTextWithFlavor("Getting revisions for app {{.AppName}} ...", map[string]interface{}{
		"AppName": cmd.RequiredArgs.AppName,
	})
	cmd.UI.DisplayText("")
	cmd.UI.DisplayText("Revision   Last deployed                   Description")
	cmd.UI.DisplayText("v1*        Tue 29 Jan 12:55:30 PST 2019    sha dfl32d")

	cmd.UI.DisplayText("v2         Tue 22 Jan 11:45:30 PST 2019    sha fd982g")
	cmd.UI.DisplayText("")
	cmd.UI.DisplayText("* running")
}
