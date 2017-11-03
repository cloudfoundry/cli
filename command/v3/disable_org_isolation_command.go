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
	cmd.SharedActor = sharedaction.NewActor(config)

	client, _, err := shared.NewClients(config, ui, true)
	if err != nil {
		if v3Err, ok := err.(ccerror.V3UnexpectedResponseError); ok && v3Err.ResponseCode == http.StatusNotFound {
			return translatableerror.MinimumAPIVersionNotMetError{MinimumVersion: ccversion.MinVersionIsolationSegmentV3}
		}

		return err
	}
	cmd.Actor = v3action.NewActor(client, config, nil, nil)

	return nil
}

func (cmd DisableOrgIsolationCommand) Execute(args []string) error {
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

	cmd.UI.DisplayTextWithFlavor("Removing entitlement to isolation segment {{.SegmentName}} from org {{.OrgName}} as {{.CurrentUser}}...", map[string]interface{}{
		"SegmentName": cmd.RequiredArgs.IsolationSegmentName,
		"OrgName":     cmd.RequiredArgs.OrganizationName,
		"CurrentUser": user.Name,
	})

	warnings, err := cmd.Actor.RevokeIsolationSegmentFromOrganizationByName(cmd.RequiredArgs.IsolationSegmentName, cmd.RequiredArgs.OrganizationName)

	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
