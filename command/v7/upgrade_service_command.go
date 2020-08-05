package v7

import (
	"code.cloudfoundry.org/cli/cf/errors"
	"strings"

	"code.cloudfoundry.org/cli/command/flag"
)

type UpgradeServiceCommand struct {
	BaseCommand

	RequiredArgs flag.ServiceInstance `positional-args:"yes"`
	ForceUpgrade bool                 `short:"f" long:"force" description:"Force upgrade without asking for confirmation"`

	relatedCommands interface{} `related_commands:"services, update-service, update-user-provided-service"`
}

func (cmd UpgradeServiceCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	return errors.New("WIP: Not yet implemented")
}

func (cmd UpgradeServiceCommand) Usage() string {
	return strings.TrimSpace(`
		CF_NAME upgrade-service SERVICE_INSTANCE
	`)
}
