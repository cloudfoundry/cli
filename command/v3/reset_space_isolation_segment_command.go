package v3

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	sharedV2 "code.cloudfoundry.org/cli/command/v2/shared"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . ResetSpaceIsolationSegmentActor

type ResetSpaceIsolationSegmentActor interface {
	CloudControllerAPIVersion() string
	ResetSpaceIsolationSegment(orgGUID string, spaceGUID string) (string, v3action.Warnings, error)
}

//go:generate counterfeiter . ResetSpaceIsolationSegmentActorV2

type ResetSpaceIsolationSegmentActorV2 interface {
	GetSpaceByOrganizationAndName(orgGUID string, spaceName string) (v2action.Space, v2action.Warnings, error)
}

type ResetSpaceIsolationSegmentCommand struct {
	RequiredArgs    flag.ResetSpaceIsolationArgs `positional-args:"yes"`
	usage           interface{}                  `usage:"CF_NAME reset-space-isolation-segment SPACE_NAME"`
	relatedCommands interface{}                  `related_commands:"org, restart, space"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       ResetSpaceIsolationSegmentActor
	ActorV2     ResetSpaceIsolationSegmentActorV2
}

func (cmd *ResetSpaceIsolationSegmentCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor()

	ccClient, _, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v3action.NewActor(ccClient, config)

	ccClientV2, uaaClientV2, err := sharedV2.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.ActorV2 = v2action.NewActor(ccClientV2, uaaClientV2, config)

	return nil
}

func (cmd ResetSpaceIsolationSegmentCommand) Execute(args []string) error {
	err := command.MinimumAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), command.MinVersionIsolationSegmentV3)
	if err != nil {
		return err
	}

	err = cmd.SharedActor.CheckTarget(cmd.Config, true, false)
	if err != nil {
		return shared.HandleError(err)
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Resetting isolation segment assignment of space {{.SpaceName}} in org {{.OrgName}} as {{.CurrentUser}}...", map[string]interface{}{
		"SpaceName":   cmd.RequiredArgs.SpaceName,
		"OrgName":     cmd.Config.TargetedOrganization().Name,
		"CurrentUser": user.Name,
	})

	space, v2Warnings, err := cmd.ActorV2.GetSpaceByOrganizationAndName(cmd.Config.TargetedOrganization().GUID, cmd.RequiredArgs.SpaceName)
	cmd.UI.DisplayWarnings(v2Warnings)
	if err != nil {
		return sharedV2.HandleError(err)
	}

	newIsolationSegmentName, warnings, err := cmd.Actor.ResetSpaceIsolationSegment(cmd.Config.TargetedOrganization().GUID, space.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()

	if newIsolationSegmentName == "" {
		cmd.UI.DisplayText("Applications in this space will be placed in the platform default isolation segment.")
	} else {
		cmd.UI.DisplayText("Applications in this space will be placed in isolation segment {{.orgIsolationSegment}}.", map[string]interface{}{
			"orgIsolationSegment": newIsolationSegmentName,
		})
	}
	cmd.UI.DisplayText("Running applications need a restart to be moved there.")

	return nil
}
