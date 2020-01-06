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
	UpdateDomainLabelsByDomainName(string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateOrganizationLabelsByOrganizationName(string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateRouteLabels(string, string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateSpaceLabelsBySpaceName(string, string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateStackLabelsByStackName(string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateServiceBrokerLabelsByServiceBrokerName(string, map[string]types.NullString) (v7action.Warnings, error)
}

type SetLabelCommand struct {
	RequiredArgs    flag.SetLabelArgs `positional-args:"yes"`
	usage           interface{}       `usage:"CF_NAME set-label RESOURCE RESOURCE_NAME KEY=VALUE...\n\nEXAMPLES:\n   cf set-label app dora env=production\n   cf set-label org business pci=true public-facing=false\n   cf set-label buildpack go_buildpack go=1.12 -s cflinuxfs3\n\nRESOURCES:\n   app\n   buildpack\n   domain\n   org\n   route\n   service-broker\n   space\n   stack"`
	relatedCommands interface{}       `related_commands:"labels, unset-label"`
	BuildpackStack  string            `long:"stack" short:"s" description:"Specify stack to disambiguate buildpacks with the same name"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SetLabelActor

	username string
	labels   map[string]types.NullString
}

func (cmd *SetLabelCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)
	ccClient, _, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil, clock.NewClock())
	return nil
}

func (cmd SetLabelCommand) Execute(args []string) error {
	cmd.labels = make(map[string]types.NullString)
	for _, label := range cmd.RequiredArgs.Labels {
		parts := strings.SplitN(label, "=", 2)
		if len(parts) < 2 {
			return fmt.Errorf("Metadata error: no value provided for label '%s'", label)
		}
		cmd.labels[parts[0]] = types.NewNullString(parts[1])
	}

	var err error
	cmd.username, err = cmd.Config.CurrentUserName()
	if err != nil {
		return err
	}

	if err = cmd.validateFlags(); err != nil {
		return err
	}

	switch cmd.canonicalResourceTypeForName() {
	case App:
		err = cmd.executeApp()
	case Buildpack:
		err = cmd.executeBuildpack()
	case Domain:
		err = cmd.executeDomain()
	case Org:
		err = cmd.executeOrg()
	case Route:
		err = cmd.executeRoute()
	case Space:
		err = cmd.executeSpace()
	case Stack:
		err = cmd.executeStack()
	case ServiceBroker:
		err = cmd.executeServiceBroker()
	default:
		err = fmt.Errorf("Unsupported resource type of '%s'", cmd.RequiredArgs.ResourceType)
	}

	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}

func (cmd SetLabelCommand) executeApp() error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	cmd.displayMessageWithOrgAndSpace()

	warnings, err := cmd.Actor.UpdateApplicationLabelsByApplicationName(
		cmd.RequiredArgs.ResourceName,
		cmd.Config.TargetedSpace().GUID,
		cmd.labels,
	)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd SetLabelCommand) executeBuildpack() error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	preFlavoringText := "Setting label(s) for {{.ResourceType}} {{.ResourceName}} as {{.User}}..."
	if cmd.BuildpackStack != "" {
		preFlavoringText = "Setting label(s) for {{.ResourceType}} {{.ResourceName}} with stack {{.StackName}} as {{.User}}..."
	}

	cmd.UI.DisplayTextWithFlavor(
		preFlavoringText,
		map[string]interface{}{
			"ResourceType": strings.ToLower(cmd.RequiredArgs.ResourceType),
			"ResourceName": cmd.RequiredArgs.ResourceName,
			"StackName":    cmd.BuildpackStack,
			"User":         cmd.username,
		},
	)

	warnings, err := cmd.Actor.UpdateBuildpackLabelsByBuildpackNameAndStack(cmd.RequiredArgs.ResourceName, cmd.BuildpackStack, cmd.labels)
	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd SetLabelCommand) executeOrg() error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	cmd.displayMessage()

	warnings, err := cmd.Actor.UpdateOrganizationLabelsByOrganizationName(
		cmd.RequiredArgs.ResourceName,
		cmd.labels,
	)
	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd SetLabelCommand) executeRoute() error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	cmd.displayMessageWithOrgAndSpace()

	warnings, err := cmd.Actor.UpdateRouteLabels(
		cmd.RequiredArgs.ResourceName,
		cmd.Config.TargetedSpace().GUID,
		cmd.labels,
	)
	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd SetLabelCommand) executeSpace() error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor(
		"Setting label(s) for {{.ResourceType}} {{.ResourceName}} in org {{.OrgName}} as {{.User}}...",
		map[string]interface{}{
			"ResourceType": strings.ToLower(cmd.RequiredArgs.ResourceType),
			"ResourceName": cmd.RequiredArgs.ResourceName,
			"OrgName":      cmd.Config.TargetedOrganization().Name,
			"User":         cmd.username,
		},
	)

	warnings, err := cmd.Actor.UpdateSpaceLabelsBySpaceName(
		cmd.RequiredArgs.ResourceName,
		cmd.Config.TargetedOrganization().GUID,
		cmd.labels,
	)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd SetLabelCommand) executeStack() error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	cmd.displayMessage()

	warnings, err := cmd.Actor.UpdateStackLabelsByStackName(cmd.RequiredArgs.ResourceName, cmd.labels)
	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd SetLabelCommand) executeDomain() error {
	cmd.displayMessage()

	warnings, err := cmd.Actor.UpdateDomainLabelsByDomainName(cmd.RequiredArgs.ResourceName, cmd.labels)
	cmd.UI.DisplayWarnings(warnings)
	return err
}

func (cmd SetLabelCommand) executeServiceBroker() error {
	if err := cmd.SharedActor.CheckTarget(false, false); err != nil {
		return err
	}

	cmd.displayMessage()

	warnings, err := cmd.Actor.UpdateServiceBrokerLabelsByServiceBrokerName(cmd.RequiredArgs.ResourceName, cmd.labels)
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

func (cmd SetLabelCommand) displayMessage() {
	cmd.UI.DisplayTextWithFlavor(
		"Setting label(s) for {{.ResourceType}} {{.ResourceName}} as {{.User}}...",
		map[string]interface{}{
			"ResourceType": strings.ToLower(cmd.RequiredArgs.ResourceType),
			"ResourceName": cmd.RequiredArgs.ResourceName,
			"User":         cmd.username,
		},
	)
}

func (cmd SetLabelCommand) displayMessageWithOrgAndSpace() {
	cmd.UI.DisplayTextWithFlavor(
		"Setting label(s) for {{.ResourceType}} {{.ResourceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.User}}...",
		map[string]interface{}{
			"ResourceType": strings.ToLower(cmd.RequiredArgs.ResourceType),
			"ResourceName": cmd.RequiredArgs.ResourceName,
			"OrgName":      cmd.Config.TargetedOrganization().Name,
			"SpaceName":    cmd.Config.TargetedSpace().Name,
			"User":         cmd.username,
		},
	)
}
