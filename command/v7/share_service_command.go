package v7

import "code.cloudfoundry.org/cli/command/flag"

type ShareServiceCommand struct {
	BaseCommand

	RequiredArgs    flag.ShareServiceArgs `positional-args:"yes"`
	OrgName         string                `short:"o" required:"false" description:"Org of the other space (Default: targeted org)"`
	relatedCommands interface{}           `related_commands:"bind-service, service, services, unshare-service"`
}

func (cmd ShareServiceCommand) Usage() string {
	return "CF_NAME share-service SERVICE_INSTANCE OTHER_SPACE [-o OTHER_ORG]"
}

func (cmd ShareServiceCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	return nil
}
