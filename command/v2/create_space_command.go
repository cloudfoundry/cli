package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type CreateSpaceCommand struct {
	RequiredArgs    flag.Space  `positional-args:"yes"`
	Organization    string      `short:"o" description:"Organization"`
	Quota           string      `short:"q" description:"Quota to assign to the newly created space"`
	usage           interface{} `usage:"CF_NAME create-space SPACE [-o ORG] [-q SPACE_QUOTA]"`
	relatedCommands interface{} `related_commands:"target, space-quotas, spaces"`
}

func (_ CreateSpaceCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ CreateSpaceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
