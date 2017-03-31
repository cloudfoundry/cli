package v3

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . DisableOrgIsolationActor

type DisableOrgIsolationActor interface {
	CloudControllerAPIVersion() string
	RevokeIsolationSegmentFromOrganizationByName(isolationSegmentName string, orgName string) (v3action.Warnings, error)
}
type DisableOrgIsolationCommand struct {
	RequiredArgs    flag.OrgIsolationArgs `positional-args:"yes"`
	usage           interface{}           `usage:"CF_NAME disable-org-isolation ORG_NAME SEGMENT_NAME"`
	relatedCommands interface{}           `related_commands:"enable-org-isolation, isolation-segments"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       DisableOrgIsolationActor
}

func (cmd *DisableOrgIsolationCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor()

	client, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v3action.NewActor(client, config)

	return nil
}

func (cmd DisableOrgIsolationCommand) Execute(args []string) error {
	err := command.MinimumAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), "3.11.0")
	if err != nil {
		return err
	}

	err = cmd.SharedActor.CheckTarget(cmd.Config, false, false)
	if err != nil {
		return shared.HandleError(err)
	}
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Removing entitlement to isolation segment {{.SegmentName}} from org {{.OrgName}} as {{.CurrentUser}}...", map[string]interface{}{
		"SegmentName": cmd.RequiredArgs.IsolationSegmentName,
		"OrgName":     cmd.RequiredArgs.OrganizationName,
		"CurrentUser": user.Name,
	})

	warnings, err := cmd.Actor.RevokeIsolationSegmentFromOrganizationByName(cmd.RequiredArgs.IsolationSegmentName, cmd.RequiredArgs.OrganizationName)

	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayOK()

	return nil
}
