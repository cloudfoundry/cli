package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type DeleteQuotaCommand struct {
	RequiredArgs    flags.Quota `positional-args:"yes"`
	Force           bool        `short:"f" description:"Force deletion without confirmation"`
	usage           interface{} `usage:"CF_NAME delete-quota QUOTA [-f]"`
	relatedCommands interface{} `related_commands:"quotas"`
}

func (_ DeleteQuotaCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ DeleteQuotaCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
