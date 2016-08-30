package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type QuotaCommand struct {
	RequiredArgs    flags.Quota `positional-args:"yes"`
	usage           interface{} `usage:"CF_NAME quota QUOTA"`
	relatedCommands interface{} `related_commands:"org, quotas"`
}

func (_ QuotaCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ QuotaCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
