package v3

import (
	"net/http"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	sharedV2 "code.cloudfoundry.org/cli/command/v2/shared"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . UnshareServiceActor

type UnshareServiceActor interface {
	UnshareServiceInstanceFromSpace(serviceInstanceName string, sourceSpaceGUID string, sharedToSpaceGUID string) (v3action.Warnings, error)
	CloudControllerAPIVersion() string
}

//go:generate counterfeiter . ServiceInstanceSharedToActorV2

type ServiceInstanceSharedToActorV2 interface {
	GetSharedToSpaceGUID(serviceInstanceName string, sourceSpaceGUID string, sharedToOrgName string, sharedToSpaceName string) (string, v2action.Warnings, error)
}

type V3UnshareServiceCommand struct {
	RequiredArgs    flag.ServiceInstance `positional-args:"yes"`
	OrgName         string               `short:"o" required:"false" description:"Org of the other space (Default: targeted org)"`
	SpaceName       string               `short:"s" required:"true" description:"Space to unshare the service instance from"`
	Force           bool                 `short:"f" description:"Force unshare without confirmation"`
	usage           interface{}          `usage:"cf v3-unshare-service SERVICE_INSTANCE -s OTHER_SPACE [-o OTHER_ORG] [-f]"`
	relatedCommands interface{}          `related_commands:"delete-service, service, services, unbind-service, v3-share-service"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       UnshareServiceActor
	ActorV2     ServiceInstanceSharedToActorV2
}

func (cmd *V3UnshareServiceCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	client, _, err := shared.NewClients(config, ui, true)
	if err != nil {
		if v3Err, ok := err.(ccerror.V3UnexpectedResponseError); ok && v3Err.ResponseCode == http.StatusNotFound {
			return translatableerror.MinimumAPIVersionNotMetError{MinimumVersion: ccversion.MinVersionShareServiceV3}
		}
		return err
	}
	cmd.Actor = v3action.NewActor(client, config, nil, nil)

	ccClientV2, uaaClientV2, err := sharedV2.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.ActorV2 = v2action.NewActor(ccClientV2, uaaClientV2, config)

	return nil
}

func (cmd V3UnshareServiceCommand) Execute(args []string) error {
	err := command.MinimumAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), ccversion.MinVersionShareServiceV3)
	if err != nil {
		return err
	}

	cmd.UI.DisplayText(command.ExperimentalWarning)
	cmd.UI.DisplayNewline()

	err = cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	orgName := cmd.Config.TargetedOrganization().Name

	if cmd.OrgName != "" {
		orgName = cmd.OrgName
	}

	if !cmd.Force {
		cmd.UI.DisplayWarning("WARNING: Unsharing this service instance will remove any service bindings that exist in any spaces that this instance is shared into. This could cause applications to stop working.\n")

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
		"SpaceName":           cmd.SpaceName,
		"Username":            user.Name,
	})

	sharedToSpaceGUID, warningsV2, err := cmd.ActorV2.GetSharedToSpaceGUID(cmd.RequiredArgs.ServiceInstance, cmd.Config.TargetedSpace().GUID, orgName, cmd.SpaceName)
	cmd.UI.DisplayWarnings(warningsV2)

	if err != nil {
		return err
	}

	warnings, err := cmd.Actor.UnshareServiceInstanceFromSpace(cmd.RequiredArgs.ServiceInstance, cmd.Config.TargetedSpace().GUID, sharedToSpaceGUID)
	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
