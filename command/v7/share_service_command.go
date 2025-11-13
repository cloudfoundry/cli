package v7

import (
	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/types"
)

type ShareServiceCommand struct {
	BaseCommand

	RequiredArgs    flag.ServiceInstance `positional-args:"yes"`
	SpaceName       string               `short:"s" required:"true" description:"The space to share the service instance into"`
	OrgName         flag.OptionalString  `short:"o" required:"false" description:"Org of the other space (Default: targeted org)"`
	relatedCommands interface{}          `related_commands:"bind-service, service, services, unshare-service"`
}

func (cmd ShareServiceCommand) Usage() string {
	return "CF_NAME share-service SERVICE_INSTANCE -s OTHER_SPACE [-o OTHER_ORG]"
}

func (cmd ShareServiceCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	if err := cmd.displayIntro(); err != nil {
		return err
	}

	warnings, err := cmd.Actor.ShareServiceInstanceToSpaceAndOrg(
		string(cmd.RequiredArgs.ServiceInstance),
		cmd.Config.TargetedSpace().GUID,
		cmd.Config.TargetedOrganization().GUID,
		v7action.ServiceInstanceSharingParams{
			SpaceName: cmd.SpaceName,
			OrgName:   types.OptionalString(cmd.OrgName),
		})

	cmd.UI.DisplayWarnings(warnings)

	switch err.(type) {
	case nil:
	case ccerror.ServiceInstanceAlreadySharedError:
		cmd.UI.DisplayOK()
		cmd.UI.DisplayTextWithFlavor("A service instance called {{.ServiceInstanceName}} has already been shared", map[string]interface{}{"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance})
		return nil
	default:
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}

func (cmd ShareServiceCommand) displayIntro() error {
	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	orgName := cmd.OrgName.Value
	if !cmd.OrgName.IsSet {
		orgName = cmd.Config.TargetedOrganization().Name
	}

	cmd.UI.DisplayTextWithFlavor(
		"Sharing service instance {{.ServiceInstanceName}} into org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance,
			"OrgName":             orgName,
			"SpaceName":           cmd.SpaceName,
			"Username":            user.Name,
		},
	)

	return nil
}
