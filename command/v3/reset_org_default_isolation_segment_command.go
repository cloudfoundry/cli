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

//go:generate counterfeiter . ResetOrgDefaultIsolationSegmentActor

type ResetOrgDefaultIsolationSegmentActor interface {
	CloudControllerAPIVersion() string
	ResetOrganizationDefaultIsolationSegment(orgGUID string) (v3action.Warnings, error)
}

//go:generate counterfeiter . ResetOrgDefaultIsolationSegmentActorV2

type ResetOrgDefaultIsolationSegmentActorV2 interface {
	GetOrganizationByName(orgName string) (v2action.Organization, v2action.Warnings, error)
}

type ResetOrgDefaultIsolationSegmentCommand struct {
	RequiredArgs    flag.ResetOrgDefaultIsolationArgs `positional-args:"yes"`
	usage           interface{}                       `usage:"CF_NAME reset-org-default-isolation-segment ORG_NAME"`
	relatedCommands interface{}                       `related_commands:"org, restart"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       ResetOrgDefaultIsolationSegmentActor
	ActorV2     ResetOrgDefaultIsolationSegmentActorV2
}

func (cmd *ResetOrgDefaultIsolationSegmentCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	client, _, err := shared.NewClients(config, ui, true)
	if err != nil {
		if v3Err, ok := err.(ccerror.V3UnexpectedResponseError); ok && v3Err.ResponseCode == http.StatusNotFound {
			return translatableerror.MinimumAPIVersionNotMetError{MinimumVersion: ccversion.MinVersionIsolationSegmentV3}
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

func (cmd ResetOrgDefaultIsolationSegmentCommand) Execute(args []string) error {
	err := command.MinimumAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), ccversion.MinVersionIsolationSegmentV3)
	if err != nil {
		return err
	}

	err = cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Resetting default isolation segment of org {{.OrgName}} as {{.CurrentUser}}...", map[string]interface{}{
		"OrgName":     cmd.RequiredArgs.OrgName,
		"CurrentUser": user.Name,
	})

	organization, v2Warnings, err := cmd.ActorV2.GetOrganizationByName(cmd.RequiredArgs.OrgName)
	cmd.UI.DisplayWarnings(v2Warnings)
	if err != nil {
		return err
	}

	warnings, err := cmd.Actor.ResetOrganizationDefaultIsolationSegment(organization.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()

	cmd.UI.DisplayText("Applications in spaces of this org that have no isolation segment assigned will be placed in the platform default isolation segment.")
	cmd.UI.DisplayText("Running applications need a restart to be moved there.")

	return nil
}
