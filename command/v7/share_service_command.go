package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/types"
)

type ShareServiceCommand struct {
	BaseCommand

	RequiredArgs    flag.ShareServiceArgs `positional-args:"yes"`
	OrgName         flag.OptionalString   `short:"o" required:"false" description:"Org of the other space (Default: targeted org)"`
	relatedCommands interface{}           `related_commands:"bind-service, service, services, unshare-service"`
}

func (cmd ShareServiceCommand) Usage() string {
	return "CF_NAME share-service SERVICE_INSTANCE OTHER_SPACE [-o OTHER_ORG]"
}

func (cmd ShareServiceCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	cmd.displayIntro()

	warnings, err := cmd.Actor.ShareServiceInstanceToSpaceAndOrg(
		cmd.RequiredArgs.ServiceInstance,
		cmd.Config.TargetedSpace().GUID,
		cmd.Config.TargetedOrganization().GUID,
		v7action.ServiceInstanceSharingParams{
			SpaceName: cmd.RequiredArgs.SpaceName,
			OrgName:   types.OptionalString(cmd.OrgName),
		})

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	//cmd.UI.DisplayOK()

	return nil
}

func (cmd ShareServiceCommand) displayIntro() error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	orgName := cmd.OrgName.Value
	if !cmd.OrgName.IsSet {
		orgName = cmd.Config.TargetedOrganization().Name
	}

	cmd.UI.DisplayTextWithFlavor(
		"Sharing service instance {{.ServiceInstanceName}} to org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance,
			"OrgName":             orgName,
			"SpaceName":           cmd.RequiredArgs.SpaceName,
			"Username":            user.Name,
		},
	)

	return nil
}
