package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type EnableServiceAccessCommand struct {
	RequiredArgs    flag.Service `positional-args:"yes"`
	Organization    string       `short:"o" description:"Enable access for a specified organization"`
	ServicePlan     string       `short:"p" description:"Enable access to a specified service plan"`
	usage           interface{}  `usage:"CF_NAME enable-service-access SERVICE [-p PLAN] [-o ORG]"`
	relatedCommands interface{}  `related_commands:"marketplace, service-access, service-brokers"`
}

func (_ EnableServiceAccessCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ EnableServiceAccessCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
