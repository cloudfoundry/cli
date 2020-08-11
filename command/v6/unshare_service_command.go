package v6

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2v3action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . UnshareServiceActor

type UnshareServiceActor interface {
	UnshareServiceInstanceFromOrganizationNameAndSpaceNameByNameAndSpace(sharedToOrgName string, sharedToSpaceName string, serviceInstanceName string, currentlyTargetedSpaceGUID string) (v2v3action.Warnings, error)
	CloudControllerV3APIVersion() string
}

type UnshareServiceCommand struct {
	RequiredArgs      flag.ServiceInstance `positional-args:"yes"`
	SharedToOrgName   string               `short:"o" required:"false" description:"Org of the other space (Default: targeted org)"`
	SharedToSpaceName string               `short:"s" required:"true" description:"Space to unshare the service instance from"`
	Force             bool                 `short:"f" description:"Force unshare without confirmation"`
	usage             interface{}          `usage:"CF_NAME unshare-service SERVICE_INSTANCE -s OTHER_SPACE [-o OTHER_ORG] [-f]"`
	relatedCommands   interface{}          `related_commands:"delete-service, service, services, share-service, unbind-service"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       UnshareServiceActor
}

func (cmd *UnshareServiceCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config

	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClientV3, uaaClientV3, err := shared.NewV3BasedClients(config, ui, true)
	if err != nil {
		return err
	}

	ccClientV2, uaaClientV2, err := shared.GetNewClientsAndConnectToCF(config, ui)
	if err != nil {
		return err
	}

	cmd.Actor = v2v3action.NewActor(
		v2action.NewActor(ccClientV2, uaaClientV2, config),
		v3action.NewActor(ccClientV3, config, sharedActor, uaaClientV3),
	)

	return nil
}

func (cmd UnshareServiceCommand) Execute(args []string) error {
	err := command.MinimumCCAPIVersionCheck(cmd.Config.APIVersion(), ccversion.MinSupportedV3ClientVersion)
	if err != nil {
		return err
	}

	err = cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	orgName := cmd.Config.TargetedOrganization().Name
	if cmd.SharedToOrgName != "" {
		orgName = cmd.SharedToOrgName
	}

	if !cmd.Force {
		cmd.UI.DisplayWarning("WARNING: Unsharing this service instance will remove any service bindings that exist in any spaces that this instance is shared into. This could cause applications to stop working.")
		cmd.UI.DisplayNewline()

		response, promptErr := cmd.UI.DisplayBoolPrompt(false, "Really unshare the service instance?", map[string]interface{}{
			"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance,
		})
		if promptErr != nil {
			return promptErr
		}
		if !response {
			cmd.UI.DisplayText("Unshare cancelled")
			return nil
		}
	}

	cmd.UI.DisplayTextWithFlavor("Unsharing service instance {{.ServiceInstanceName}} from org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance,
		"OrgName":             orgName,
		"SpaceName":           cmd.SharedToSpaceName,
		"Username":            user.Name,
	})

	warnings, err := cmd.Actor.UnshareServiceInstanceFromOrganizationNameAndSpaceNameByNameAndSpace(orgName, cmd.SharedToSpaceName, cmd.RequiredArgs.ServiceInstance.Value, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		switch err.(type) {
		case actionerror.ServiceInstanceNotSharedToSpaceError:
			cmd.UI.DisplayText("Service instance {{.ServiceInstanceName}} is not shared with space {{.SpaceName}} in organization {{.OrgName}}.", map[string]interface{}{
				"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance,
				"SpaceName":           cmd.SharedToSpaceName,
				"OrgName":             orgName,
			})
		default:
			return err
		}
	}

	cmd.UI.DisplayOK()
	return nil
}
