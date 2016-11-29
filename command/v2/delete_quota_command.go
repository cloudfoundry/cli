package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type DeleteQuotaCommand struct {
	RequiredArgs    flag.Quota  `positional-args:"yes"`
	Force           bool        `short:"f" description:"Force deletion without confirmation"`
	usage           interface{} `usage:"CF_NAME delete-quota QUOTA [-f]"`
	relatedCommands interface{} `related_commands:"quotas"`
}

func (_ DeleteQuotaCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ DeleteQuotaCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
