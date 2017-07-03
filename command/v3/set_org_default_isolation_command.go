package v3

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . EnableOrgIsolationActor

type SetOrgDefaultIsolationSegmentActor interface {
	CloudControllerAPIVersion() string
	SetDefaultIsolationSegmentOnOrganizationByName(isolationSegmentName string, orgName string) (v3action.Warnings, error)
}

type SetOrgDefaultIsolationSegmentCommand struct {
	RequiredArgs    flag.OrgIsolationArgs `positional-args:"yes"`
	usage           interface{}           `usage:"CF_NAME set-org-default-isolation-segment ORG_NAME SEGMENT_NAME"`
	relatedCommands interface{}           `related_commands:"create-isolation-segment, isolation-segments, set-org-default-isolation-segment, set-space-isolation-segment"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SetOrgDefaultIsolationSegmentActor
}

func (cmd *SetOrgDefaultIsolationSegmentCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor()

	client, _, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v3action.NewActor(client, config)

	return nil
}

func (cmd SetOrgDefaultIsolationSegmentCommand) Execute(args []string) error {
	// err := command.MinimumAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), command.MinVersionIsolationSegmentV3)
	// if err != nil {
	// 	return err
	// }

	err := cmd.SharedActor.CheckTarget(cmd.Config, false, false)
	if err != nil {
		return shared.HandleError(err)
	}

	_, err = cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	// cmd.UI.DisplayTextWithFlavor("Enabling isolation segment {{.SegmentName}} for org {{.OrgName}} as {{.CurrentUser}}...", map[string]interface{}{
	// 	"SegmentName": cmd.RequiredArgs.IsolationSegmentName,
	// 	"OrgName":     cmd.RequiredArgs.OrganizationName,
	// 	"CurrentUser": user.Name,
	// })

	// warnings, err := cmd.Actor.SetDefaultIsolationSegmentOnOrganizationByName(cmd.RequiredArgs.IsolationSegmentName, cmd.RequiredArgs.OrganizationName)
	// cmd.UI.DisplayWarnings(warnings)
	// if err != nil {
	// 	return shared.HandleError(err)
	// }

	// cmd.UI.DisplayOK()

	return nil
}
