package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type CreateSpaceCommand struct {
	RequiredArgs    flag.Space  `positional-args:"yes"`
	Organization    string      `short:"o" description:"Organization"`
	Quota           string      `short:"q" description:"Quota to assign to the newly created space"`
	usage           interface{} `usage:"CF_NAME create-space SPACE [-o ORG] [-q SPACE_QUOTA]"`
	relatedCommands interface{} `related_commands:"set-space-isolation-segment, space-quotas, spaces, target"`
}

func (CreateSpaceCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (CreateSpaceCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
