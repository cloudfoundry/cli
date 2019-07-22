package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . UnsetLabelActor

type UnsetLabelActor interface {
	UpdateApplicationLabelsByApplicationName(string, string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateOrganizationLabelsByOrganizationName(string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateSpaceLabelsBySpaceName(string, string, map[string]types.NullString) (v7action.Warnings, error)
}

type UnsetLabelCommand struct {
	RequiredArgs flag.UnsetLabelArgs `positional-args:"yes"`
	usage        interface{}         `usage:"CF_NAME unset-label RESOURCE RESOURCE_NAME KEY\n\nEXAMPLES:\n   cf unset-label app dora ci_signature_sha2\n\nRESOURCES:\n   app\n\nSEE ALSO:\n   set-label, labels"`
	UI           command.UI
	Config       command.Config
	SharedActor  command.SharedActor
	Actor        UnsetLabelActor
}

func (cmd *UnsetLabelCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)
	ccClient, _, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil, clock.NewClock())
	return nil
}

func (cmd UnsetLabelCommand) Execute(args []string) error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	labels := make(map[string]types.NullString)
	for _, value := range cmd.RequiredArgs.LabelKeys {
		labels[value] = types.NewNullString()
	}

	switch ResourceType(cmd.RequiredArgs.ResourceType) {
	case App:
		err = cmd.executeApp(user.Name, labels)
	case Org:
		err = cmd.executeOrg(user.Name, labels)
	case Space:
		err = cmd.executeSpace(user.Name, labels)
	default:
		err = errors.New(cmd.UI.TranslateText("Unsupported resource type of '{{.ResourceType}}'", map[string]interface{}{"ResourceType": cmd.RequiredArgs.ResourceType}))
	}

	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}

func (cmd UnsetLabelCommand) executeApp(username string, labels map[string]types.NullString) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Removing label(s) for app {{.ResourceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.User}}...", map[string]interface{}{
		"ResourceName": cmd.RequiredArgs.ResourceName,
		"OrgName":      cmd.Config.TargetedOrganization().Name,
		"SpaceName":    cmd.Config.TargetedSpace().Name,
		"User":         username,
	})

	warnings, err := cmd.Actor.UpdateApplicationLabelsByApplicationName(cmd.RequiredArgs.ResourceName, cmd.Config.TargetedSpace().GUID, labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd UnsetLabelCommand) executeOrg(username string, labels map[string]types.NullString) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Removing label(s) for org {{.ResourceName}} as {{.User}}...", map[string]interface{}{
		"ResourceName": cmd.RequiredArgs.ResourceName,
		"User":         username,
	})

	warnings, err := cmd.Actor.UpdateOrganizationLabelsByOrganizationName(cmd.RequiredArgs.ResourceName, labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd UnsetLabelCommand) executeSpace(username string, labels map[string]types.NullString) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Removing label(s) for space {{.ResourceName}} in org {{.OrgName}} as {{.User}}...", map[string]interface{}{
		"ResourceName": cmd.RequiredArgs.ResourceName,
		"OrgName":      cmd.Config.TargetedOrganization().Name,
		"User":         username,
	})

	warnings, err := cmd.Actor.UpdateSpaceLabelsBySpaceName(cmd.RequiredArgs.ResourceName, cmd.Config.TargetedOrganization().GUID, labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}
