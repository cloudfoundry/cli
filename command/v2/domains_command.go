package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
)

type DomainsCommand struct {
	usage           interface{} `usage:"CF_NAME domains"`
	relatedCommands interface{} `related_commands:"router-groups, create-route, routes"`
}

func (_ DomainsCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ DomainsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
