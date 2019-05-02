package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/types"
	"fmt"
)

//go:generate counterfeiter . DeleteLabelActor

type DeleteLabelActor interface {
	UpdateApplicationLabelsByApplicationName(string, string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateOrganizationLabelsByOrganizationName(string, map[string]types.NullString) (v7action.Warnings, error)
}

type DeleteLabelCommand struct {
	RequiredArgs flag.DeleteLabelArgs `positional-args:"yes"`
	usage        interface{}          `usage:"CF_NAME delete-label RESOURCE RESOURCE_NAME KEY\n\nEXAMPLES:\n   cf delete-label app dora ci_signature_sha2\n\nRESOURCES:\n   app\n\nSEE ALSO:\n   set-label, labels"`
	UI           command.UI
	Config       command.Config
	SharedActor  command.SharedActor
	Actor        DeleteLabelActor
}

func (cmd *DeleteLabelCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)
	ccClient, _, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil)
	return nil
}

func (cmd DeleteLabelCommand) Execute(args []string) error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	labels := make(map[string]types.NullString)
	for _, value := range cmd.RequiredArgs.LabelKeys {
		labels[value] = types.NewNullString()
	}

	switch cmd.RequiredArgs.ResourceType {
	case "app":
		err = cmd.executeApp(user.Name, labels)
	case "org":
		err = cmd.executeOrg(user.Name, labels)
	default:
		err = fmt.Errorf("Unsupported resource type of '%s'", cmd.RequiredArgs.ResourceType)
	}

	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}

func (cmd DeleteLabelCommand) executeApp(username string, labels map[string]types.NullString) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Deleting label(s) for app {{.ResourceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.User}}...", map[string]interface{}{
		"ResourceName": cmd.RequiredArgs.ResourceName,
		"OrgName":      cmd.Config.TargetedOrganization().Name,
		"SpaceName":    cmd.Config.TargetedSpace().Name,
		"User":         username,
	})

	warnings, err := cmd.Actor.UpdateApplicationLabelsByApplicationName(cmd.RequiredArgs.ResourceName, cmd.Config.TargetedSpace().GUID, labels)

	for _, warning := range warnings {
		cmd.UI.DisplayWarning(warning)
	}

	return err
}

func (cmd DeleteLabelCommand) executeOrg(username string, labels map[string]types.NullString) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Deleting label(s) for org {{.ResourceName}} as {{.User}}...", map[string]interface{}{
		"ResourceName": cmd.RequiredArgs.ResourceName,
		"User":         username,
	})

	warnings, err := cmd.Actor.UpdateOrganizationLabelsByOrganizationName(cmd.RequiredArgs.ResourceName, labels)

	for _, warning := range warnings {
		cmd.UI.DisplayWarning(warning)
	}

	return err
}
