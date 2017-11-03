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

//go:generate counterfeiter . SetOrgDefaultIsolationSegmentActor

type SetOrgDefaultIsolationSegmentActor interface {
	CloudControllerAPIVersion() string
	GetIsolationSegmentByName(isoSegName string) (v3action.IsolationSegment, v3action.Warnings, error)
	SetOrganizationDefaultIsolationSegment(orgGUID string, isoSegGUID string) (v3action.Warnings, error)
}

//go:generate counterfeiter . SetOrgDefaultIsolationSegmentActorV2

type SetOrgDefaultIsolationSegmentActorV2 interface {
	GetOrganizationByName(orgName string) (v2action.Organization, v2action.Warnings, error)
}

type SetOrgDefaultIsolationSegmentCommand struct {
	RequiredArgs    flag.OrgIsolationArgs `positional-args:"yes"`
	usage           interface{}           `usage:"CF_NAME set-org-default-isolation-segment ORG_NAME SEGMENT_NAME"`
	relatedCommands interface{}           `related_commands:"org, set-space-isolation-segment"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SetOrgDefaultIsolationSegmentActor
	ActorV2     SetOrgDefaultIsolationSegmentActorV2
}

func (cmd *SetOrgDefaultIsolationSegmentCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd SetOrgDefaultIsolationSegmentCommand) Execute(args []string) error {
	err := command.MinimumAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), ccversion.MinVersionIsolationSegmentV3)
	if err != nil {
		return err
	}

	err = cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Setting isolation segment {{.IsolationSegmentName}} to default on org {{.OrgName}} as {{.CurrentUser}}...", map[string]interface{}{
		"IsolationSegmentName": cmd.RequiredArgs.IsolationSegmentName,
		"OrgName":              cmd.RequiredArgs.OrganizationName,
		"CurrentUser":          user.Name,
	})

	org, v2Warnings, err := cmd.ActorV2.GetOrganizationByName(cmd.RequiredArgs.OrganizationName)
	cmd.UI.DisplayWarnings(v2Warnings)
	if err != nil {
		return err
	}

	isoSeg, v3Warnings, err := cmd.Actor.GetIsolationSegmentByName(cmd.RequiredArgs.IsolationSegmentName)
	cmd.UI.DisplayWarnings(v3Warnings)
	if err != nil {
		return err
	}

	v3Warnings, err = cmd.Actor.SetOrganizationDefaultIsolationSegment(org.GUID, isoSeg.GUID)
	cmd.UI.DisplayWarnings(v3Warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText("In order to move running applications to this isolation segment, they must be restarted.")

	return nil
}
