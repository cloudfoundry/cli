package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type DeleteServiceCommand struct {
	BaseCommand

	RequiredArgs    flag.ServiceInstance `positional-args:"yes"`
	Force           bool                 `short:"f" description:"Force deletion without confirmation"`
	relatedCommands interface{}          `related_commands:"unbind-service, services"`
}

func (cmd DeleteServiceCommand) Execute(args []string) error {
	return cmd.SharedActor.CheckTarget(true, true)
}

func (cmd DeleteServiceCommand) Usage() string {
	return "CF_NAME delete-service SERVICE_INSTANCE [-f]"
}
