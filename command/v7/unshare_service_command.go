package v7

import "code.cloudfoundry.org/cli/command/flag"

type UnshareServiceCommand struct {
	BaseCommand

	RequiredArgs    flag.ShareServiceArgs `positional-args:"yes"`
	OrgName         string                `short:"o" required:"false" description:"Org of the other space (Default: targeted org)"`
	Force           bool                  `short:"f" description:"Force unshare without confirmation"`
	relatedCommands interface{}           `related_commands:"delete-service, service, services, share-service, unbind-service"`
}

func (cmd UnshareServiceCommand) Usage() string {
	return "CF_NAME unshare-service SERVICE_INSTANCE OTHER_SPACE [-o OTHER_ORG] [-f]"
}

func (cmd UnshareServiceCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	return nil
}
