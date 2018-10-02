package v3

import (
	"net/http"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2v3action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . ShareServiceActor

type ShareServiceActor interface {
	ShareServiceInstanceToSpaceNameByNameAndSpaceAndOrganizationName(sharedToSpaceName string, serviceInstanceName string, sourceSpaceGUID string, sharedToOrgName string) (v2v3action.Warnings, error)
	ShareServiceInstanceToSpaceNameByNameAndSpaceAndOrganization(sharedToSpaceName string, serviceInstanceName string, sourceSpaceGUID string, sharedToOrgGUID string) (v2v3action.Warnings, error)
	CloudControllerV3APIVersion() string
}

type ShareServiceCommand struct {
	RequiredArgs    flag.ServiceInstance `positional-args:"yes"`
	OrgName         string               `short:"o" required:"false" description:"Org of the other space (Default: targeted org)"`
	SpaceName       string               `short:"s" required:"true" description:"Space to share the service instance into"`
	usage           interface{}          `usage:"CF_NAME share-service SERVICE_INSTANCE -s OTHER_SPACE [-o OTHER_ORG]"`
	relatedCommands interface{}          `related_commands:"bind-service, service, services, unshare-service"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       ShareServiceActor
}

func (cmd *ShareServiceCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config

	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClientV3, uaaClientV3, err := shared.NewV3BasedClients(config, ui, true, "")
	if err != nil {
		if v3Err, ok := err.(ccerror.V3UnexpectedResponseError); ok && v3Err.ResponseCode == http.StatusNotFound {
			return translatableerror.MinimumCFAPIVersionNotMetError{MinimumVersion: ccversion.MinVersionShareServiceV3}
		}
		return err
	}

	ccClientV2, uaaClientV2, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}

	cmd.Actor = v2v3action.NewActor(
		v2action.NewActor(ccClientV2, uaaClientV2, config),
		v3action.NewActor(ccClientV3, config, sharedActor, uaaClientV3),
	)

	return nil
}

func (cmd ShareServiceCommand) Execute(args []string) error {
	err := command.MinimumCCAPIVersionCheck(cmd.Actor.CloudControllerV3APIVersion(), ccversion.MinVersionShareServiceV3)
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
	if cmd.OrgName != "" {
		orgName = cmd.OrgName
	}

	cmd.UI.DisplayTextWithFlavor("Sharing service instance {{.ServiceInstanceName}} into org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance,
		"OrgName":             orgName,
		"SpaceName":           cmd.SpaceName,
		"Username":            user.Name,
	})

	var warnings v2v3action.Warnings

	if cmd.OrgName != "" {
		warnings, err = cmd.Actor.ShareServiceInstanceToSpaceNameByNameAndSpaceAndOrganizationName(cmd.SpaceName, cmd.RequiredArgs.ServiceInstance, cmd.Config.TargetedSpace().GUID, cmd.OrgName)
	} else {
		warnings, err = cmd.Actor.ShareServiceInstanceToSpaceNameByNameAndSpaceAndOrganization(cmd.SpaceName, cmd.RequiredArgs.ServiceInstance, cmd.Config.TargetedSpace().GUID, cmd.Config.TargetedOrganization().GUID)
	}

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		switch err.(type) {
		case actionerror.ServiceInstanceAlreadySharedError:
			cmd.UI.DisplayText("Service instance {{.ServiceInstanceName}} is already shared with that space.", map[string]interface{}{
				"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance,
			})
		default:
			return err
		}
	}

	cmd.UI.DisplayOK()
	return nil
}
