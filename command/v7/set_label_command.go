package v7

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . SetLabelActor

type SetLabelActor interface {
	UpdateApplicationLabelsByApplicationName(string, string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateBuildpackLabelsByBuildpackNameAndStack(string, string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateOrganizationLabelsByOrganizationName(string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateSpaceLabelsBySpaceName(string, string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateStackLabelsByStackName(string, map[string]types.NullString) (v7action.Warnings, error)
}

type SetLabelCommand struct {
	RequiredArgs    flag.SetLabelArgs `positional-args:"yes"`
	usage           interface{}       `usage:"CF_NAME set-label RESOURCE RESOURCE_NAME KEY=VALUE...\n\nEXAMPLES:\n   cf set-label app dora env=production\n   cf set-label org business pci=true public-facing=false\n   cf set-label space business_space public-facing=false owner=jane_doe\n\nRESOURCES:\n   app\n   buildpack\n   org\n   space\n   stack"`
	relatedCommands interface{}       `related_commands:"labels, unset-label"`
	BuildpackStack  string            `long:"stack" short:"s" description:"Specify stack to disambiguate buildpacks with the same name"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SetLabelActor
}

func (cmd *SetLabelCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd SetLabelCommand) Execute(args []string) error {

	labels := make(map[string]types.NullString)
	for _, label := range cmd.RequiredArgs.Labels {
		parts := strings.SplitN(label, "=", 2)
		if len(parts) < 2 {
			return fmt.Errorf("Metadata error: no value provided for label '%s'", label)
		}
		labels[parts[0]] = types.NewNullString(parts[1])
	}

	username, err := cmd.Config.CurrentUserName()
	if err != nil {
		return err
	}

	err = cmd.validateFlags()
	if err != nil {
		return err
	}

	switch cmd.canonicalResourceTypeForName() {
	case App:
		err = cmd.executeApp(username, labels)
	case Buildpack:
		err = cmd.executeBuildpack(username, labels)
	case Org:
		err = cmd.executeOrg(username, labels)
	case Space:
		err = cmd.executeSpace(username, labels)
	case Stack:
		err = cmd.executeStack(username, labels)
	default:
		err = fmt.Errorf("Unsupported resource type of '%s'", cmd.RequiredArgs.ResourceType)
	}

	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}

func (cmd SetLabelCommand) executeApp(username string, labels map[string]types.NullString) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	appName := cmd.RequiredArgs.ResourceName

	preFlavoringText := fmt.Sprintf("Setting label(s) for %s {{.ResourceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.User}}...", strings.ToLower(cmd.RequiredArgs.ResourceType))
	cmd.UI.DisplayTextWithFlavor(
		preFlavoringText,
		map[string]interface{}{
			"ResourceName": appName,
			"OrgName":      cmd.Config.TargetedOrganization().Name,
			"SpaceName":    cmd.Config.TargetedSpace().Name,
			"User":         username,
		},
	)

	warnings, err := cmd.Actor.UpdateApplicationLabelsByApplicationName(appName,
		cmd.Config.TargetedSpace().GUID,
		labels)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	return nil
}

func (cmd SetLabelCommand) executeBuildpack(username string, labels map[string]types.NullString) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	preFlavoringText := fmt.Sprintf("Setting label(s) for %s {{.ResourceName}} as {{.User}}...", strings.ToLower(cmd.RequiredArgs.ResourceType))
	if cmd.BuildpackStack != "" {
		preFlavoringText = fmt.Sprintf("Setting label(s) for %s {{.ResourceName}} with stack {{.StackName}} as {{.User}}...", strings.ToLower(cmd.RequiredArgs.ResourceType))
	}

	cmd.UI.DisplayTextWithFlavor(
		preFlavoringText,
		map[string]interface{}{
			"ResourceName": cmd.RequiredArgs.ResourceName,
			"StackName":    cmd.BuildpackStack,
			"User":         username,
		},
	)

	warnings, err := cmd.Actor.UpdateBuildpackLabelsByBuildpackNameAndStack(cmd.RequiredArgs.ResourceName, cmd.BuildpackStack, labels)
	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd SetLabelCommand) executeOrg(username string, labels map[string]types.NullString) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	preFlavoringText := fmt.Sprintf("Setting label(s) for %s {{.ResourceName}} as {{.User}}...", strings.ToLower(cmd.RequiredArgs.ResourceType))
	cmd.UI.DisplayTextWithFlavor(
		preFlavoringText,
		map[string]interface{}{
			"ResourceName": cmd.RequiredArgs.ResourceName,
			"OrgName":      cmd.Config.TargetedOrganization().Name,
			"User":         username,
		},
	)

	warnings, err := cmd.Actor.UpdateOrganizationLabelsByOrganizationName(cmd.RequiredArgs.ResourceName,
		labels)
	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd SetLabelCommand) executeSpace(username string, labels map[string]types.NullString) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	spaceName := cmd.RequiredArgs.ResourceName

	preFlavoringText := fmt.Sprintf("Setting label(s) for %s {{.ResourceName}} in org {{.OrgName}} as {{.User}}...", strings.ToLower(cmd.RequiredArgs.ResourceType))
	cmd.UI.DisplayTextWithFlavor(
		preFlavoringText,
		map[string]interface{}{
			"ResourceName": spaceName,
			"OrgName":      cmd.Config.TargetedOrganization().Name,
			"User":         username,
		},
	)

	warnings, err := cmd.Actor.UpdateSpaceLabelsBySpaceName(spaceName,
		cmd.Config.TargetedOrganization().GUID,
		labels)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	return nil
}

func (cmd SetLabelCommand) executeStack(username string, labels map[string]types.NullString) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	preFlavoringText := fmt.Sprintf("Setting label(s) for %s {{.ResourceName}} as {{.User}}...", strings.ToLower(cmd.RequiredArgs.ResourceType))
	cmd.UI.DisplayTextWithFlavor(
		preFlavoringText,
		map[string]interface{}{
			"ResourceName": cmd.RequiredArgs.ResourceName,
			"User":         username,
		},
	)

	warnings, err := cmd.Actor.UpdateStackLabelsByStackName(cmd.RequiredArgs.ResourceName, labels)
	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd SetLabelCommand) canonicalResourceTypeForName() ResourceType {
	return ResourceType(strings.ToLower(cmd.RequiredArgs.ResourceType))
}

func (cmd SetLabelCommand) validateFlags() error {
	if cmd.BuildpackStack != "" && cmd.canonicalResourceTypeForName() != Buildpack {
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				cmd.RequiredArgs.ResourceType, "--stack, -s",
			},
		}
	}
	return nil
}
