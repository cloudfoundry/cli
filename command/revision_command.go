package command

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type RevisionCommand struct {
	BaseCommand
	usage           interface{}   `usage:"CF_NAME revision APP_NAME [--version VERSION]"`
	RequiredArgs    flag.AppName  `positional-args:"yes"`
	Version         flag.Revision `long:"version" required:"true" description:"The integer representing the specific revision to show"`
	relatedCommands interface{}   `related_commands:"revisions, rollback"`
}

func (cmd RevisionCommand) Execute(_ []string) error {
	cmd.UI.DisplayWarning(command.ExperimentalWarning)
	cmd.UI.DisplayNewline()
	return nil
}
