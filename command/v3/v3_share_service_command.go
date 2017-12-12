package v3

import (
	"net/http"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . ShareServiceActor

type ShareServiceActor interface {
	ShareServiceInstanceInSpaceByOrganizationAndSpaceName(serviceInstanceName string, sourceSpaceGUID string, sharedToOrgGUID string, sharedToSpaceName string) (v3action.Warnings, error)
	ShareServiceInstanceInSpaceByOrganizationNameAndSpaceName(serviceInstanceName string, sourceSpaceGUID string, sharedToOrgName string, sharedToSpaceName string) (v3action.Warnings, error)
	CloudControllerAPIVersion() string
}

type V3ShareServiceCommand struct {
	RequiredArgs    flag.ServiceInstance `positional-args:"yes"`
	OrgName         string               `short:"o" required:"false" description:"Org of the other space (Default: targeted org)"`
	SpaceName       string               `short:"s" required:"true" description:"Space to share the service instance into"`
	usage           interface{}          `usage:"cf v3-share-service SERVICE_INSTANCE -s OTHER_SPACE [-o OTHER_ORG]"`
	relatedCommands interface{}          `related_commands:"bind-service, service, services, v3-unshare-service"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       ShareServiceActor
}

func (cmd *V3ShareServiceCommand) Setup(config command.Config, ui command.UI) error {
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

	return nil
}

func (cmd V3ShareServiceCommand) Execute(args []string) error {
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

	cmd.UI.DisplayTextWithFlavor("Sharing service instance {{.ServiceInstanceName}} into org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance,
		"OrgName":             orgName,
		"SpaceName":           cmd.SpaceName,
		"Username":            user.Name,
	})

	var warnings v3action.Warnings

	if cmd.OrgName != "" {
		warnings, err = cmd.Actor.ShareServiceInstanceInSpaceByOrganizationNameAndSpaceName(cmd.RequiredArgs.ServiceInstance, cmd.Config.TargetedSpace().GUID, cmd.OrgName, cmd.SpaceName)
	} else {
		warnings, err = cmd.Actor.ShareServiceInstanceInSpaceByOrganizationAndSpaceName(cmd.RequiredArgs.ServiceInstance, cmd.Config.TargetedSpace().GUID, cmd.Config.TargetedOrganization().GUID, cmd.SpaceName)
	}

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
