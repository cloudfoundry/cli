package v7

import (
	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/command/flag"
	"code.cloudfoundry.org/cli/v8/types"
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

	if !cmd.Force {
		cmd.UI.DisplayWarning(
			`WARNING: Unsharing this service instance will remove any existing bindings originating from the service instance in the space "{{.SpaceName}}". This could cause apps to stop working.`,
			map[string]interface{}{"SpaceName": cmd.SpaceName},
		)

		unshare, err := cmd.displayPrompt()
		if err != nil {
			return err
		}

		if !unshare {
			cmd.UI.DisplayText("Unshare cancelled")
			return nil
		}
	}

	if err := cmd.displayIntro(); err != nil {
		return err
	}

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

	cmd.UI.DisplayOK()

	return nil
}

func (cmd UnshareServiceCommand) displayIntro() error {
	user, err := cmd.Actor.GetCurrentUser()
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

func (cmd UnshareServiceCommand) displayPrompt() (bool, error) {
	cmd.UI.DisplayNewline()
	unshare, err := cmd.UI.DisplayBoolPrompt(
		false,
		"Really unshare the service instance {{.ServiceInstanceName}} from space {{.SpaceName}}?",
		map[string]interface{}{
			"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance,
			"SpaceName":           cmd.SpaceName,
		})
	if err != nil {
		return false, err
	}

	return unshare, nil
}
