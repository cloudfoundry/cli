package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/types"
	"errors"
)

type UnshareServiceCommand struct {
	BaseCommand

	RequiredArgs    flag.ShareServiceArgs `positional-args:"yes"`
	SpaceName       string                `short:"s" required:"true" description:"Space to unshare the service instance from"`
	OrgName         flag.OptionalString   `short:"o" required:"false" description:"Org of the other space (Default: targeted org)"`
	Force           bool                  `short:"f" description:"Force unshare without confirmation"`
	relatedCommands interface{}           `related_commands:"delete-service, service, services, share-service, unbind-service"`
}

func (cmd UnshareServiceCommand) Usage() string {
	return "CF_NAME unshare-service SERVICE_INSTANCE -s OTHER_SPACE [-o OTHER_ORG] [-f]"
}

func (cmd UnshareServiceCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	cmd.displayIntro()

	warnings, err := cmd.Actor.UnshareServiceInstanceFromSpaceAndOrg(
		cmd.RequiredArgs.ServiceInstance,
		cmd.Config.TargetedSpace().GUID,
		cmd.Config.TargetedOrganization().GUID,
		v7action.ServiceInstanceSharingParams{
			SpaceName: cmd.SpaceName,
			OrgName:   types.OptionalString(cmd.OrgName),
		})

	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		return err
	}
	return errors.New("Not yet implemented")
}

func (cmd UnshareServiceCommand) displayIntro() error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	orgName := cmd.OrgName.Value
	if !cmd.OrgName.IsSet {
		orgName = cmd.Config.TargetedOrganization().Name
	}

	cmd.UI.DisplayTextWithFlavor(
		"Unsharing service instance {{.ServiceInstanceName}} from org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance,
			"OrgName":             orgName,
			"SpaceName":           cmd.SpaceName,
			"Username":            user.Name,
		},
	)

	return nil
}
