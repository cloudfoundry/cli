package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type DeleteServiceKeyCommand struct {
	RequiredArgs    flag.ServiceInstanceKey `positional-args:"yes"`
	Force           bool                    `short:"f" description:"Force deletion without confirmation"`
	usage           interface{}             `usage:"CF_NAME delete-service-key SERVICE_INSTANCE SERVICE_KEY [-f]\n\nEXAMPLES:\n   CF_NAME delete-service-key mydb mykey"`
	relatedCommands interface{}             `related_commands:"service-keys"`
}

func (_ DeleteServiceKeyCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ DeleteServiceKeyCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
