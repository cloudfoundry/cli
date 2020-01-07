package v7

import (
	"errors"
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

type UnsetLabelCommand struct {
	RequiredArgs    flag.UnsetLabelArgs `positional-args:"yes"`
	BuildpackStack  string              `long:"stack" short:"s" description:"Specify stack to disambiguate buildpacks with the same name"`
	usage           interface{}         `usage:"CF_NAME unset-label RESOURCE RESOURCE_NAME KEY...\n\nEXAMPLES:\n   cf unset-label app dora ci_signature_sha2\n   cf unset-label org business pci public-facing\n   cf unset-label buildpack go_buildpack go -s cflinuxfs3\n\nRESOURCES:\n   app\n   buildpack\n   domain\n   org\n   route\n   service-broker\n   space\n   stack"`
	relatedCommands interface{}         `related_commands:"labels, set-label"`
	UI              command.UI
	Config          command.Config
	SharedActor     command.SharedActor
	Actor           SetLabelActor

	username string
	labels   map[string]types.NullString
}

func (cmd *UnsetLabelCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd UnsetLabelCommand) Execute(args []string) error {
	cmd.labels = make(map[string]types.NullString)
	for _, value := range cmd.RequiredArgs.LabelKeys {
		cmd.labels[value] = types.NewNullString()
	}

	var err error
	cmd.username, err = cmd.Config.CurrentUserName()
	if err != nil {
		return err
	}

	err = cmd.validateFlags()
	if err != nil {
		return err
	}

	resourceTypeString := strings.ToLower(cmd.RequiredArgs.ResourceType)
	switch ResourceType(resourceTypeString) {
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
	case ServiceBroker:
		err = cmd.executeServiceBroker()
	case Space:
		err = cmd.executeSpace()
	case Stack:
		err = cmd.executeStack()

	default:
		err = errors.New(cmd.UI.TranslateText("Unsupported resource type of '{{.ResourceType}}'", map[string]interface{}{"ResourceType": cmd.RequiredArgs.ResourceType}))
	}

	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}

func (cmd UnsetLabelCommand) executeApp() error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	cmd.displayMessageWithOrgAndSpace()

	warnings, err := cmd.Actor.UpdateApplicationLabelsByApplicationName(cmd.RequiredArgs.ResourceName, cmd.Config.TargetedSpace().GUID, cmd.labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd UnsetLabelCommand) executeDomain() error {
	cmd.displayMessage()

	warnings, err := cmd.Actor.UpdateDomainLabelsByDomainName(cmd.RequiredArgs.ResourceName, cmd.labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd UnsetLabelCommand) executeRoute() error {
	cmd.displayMessageWithOrgAndSpace()

	warnings, err := cmd.Actor.UpdateRouteLabels(cmd.RequiredArgs.ResourceName, cmd.Config.TargetedSpace().GUID, cmd.labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd UnsetLabelCommand) executeBuildpack() error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	var template string
	if cmd.BuildpackStack == "" {
		template = "Removing label(s) for buildpack {{.ResourceName}} as {{.User}}..."
	} else {
		template = "Removing label(s) for buildpack {{.ResourceName}} with stack {{.StackName}} as {{.User}}..."
	}
	cmd.UI.DisplayTextWithFlavor(template, map[string]interface{}{
		"ResourceName": cmd.RequiredArgs.ResourceName,
		"StackName":    cmd.BuildpackStack,
		"User":         cmd.username,
	})

	warnings, err := cmd.Actor.UpdateBuildpackLabelsByBuildpackNameAndStack(cmd.RequiredArgs.ResourceName, cmd.BuildpackStack, cmd.labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd UnsetLabelCommand) executeOrg() error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	cmd.displayMessage()

	warnings, err := cmd.Actor.UpdateOrganizationLabelsByOrganizationName(cmd.RequiredArgs.ResourceName, cmd.labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd UnsetLabelCommand) executeServiceBroker() error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	cmd.displayMessage()
	warnings, err := cmd.Actor.UpdateServiceBrokerLabelsByServiceBrokerName(cmd.RequiredArgs.ResourceName, cmd.labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd UnsetLabelCommand) executeSpace() error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Removing label(s) for space {{.ResourceName}} in org {{.OrgName}} as {{.User}}...", map[string]interface{}{
		"ResourceName": cmd.RequiredArgs.ResourceName,
		"OrgName":      cmd.Config.TargetedOrganization().Name,
		"User":         cmd.username,
	})

	warnings, err := cmd.Actor.UpdateSpaceLabelsBySpaceName(cmd.RequiredArgs.ResourceName, cmd.Config.TargetedOrganization().GUID, cmd.labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd UnsetLabelCommand) executeStack() error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	cmd.displayMessage()

	warnings, err := cmd.Actor.UpdateStackLabelsByStackName(cmd.RequiredArgs.ResourceName, cmd.labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd UnsetLabelCommand) validateFlags() error {
	resourceTypeString := strings.ToLower(cmd.RequiredArgs.ResourceType)
	if cmd.BuildpackStack != "" && ResourceType(resourceTypeString) != Buildpack {
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				cmd.RequiredArgs.ResourceType, "--stack, -s",
			},
		}
	}
	return nil
}

func (cmd UnsetLabelCommand) displayMessage() {
	cmd.UI.DisplayTextWithFlavor("Removing label(s) for {{.ResourceType}} {{.ResourceName}} as {{.User}}...", map[string]interface{}{
		"ResourceType": strings.ToLower(cmd.RequiredArgs.ResourceType),
		"ResourceName": cmd.RequiredArgs.ResourceName,
		"User":         cmd.username,
	})
}
func (cmd UnsetLabelCommand) displayMessageWithOrgAndSpace() {
	cmd.UI.DisplayTextWithFlavor("Removing label(s) for {{.ResourceType}} {{.ResourceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.User}}...", map[string]interface{}{
		"ResourceType": strings.ToLower(cmd.RequiredArgs.ResourceType),
		"ResourceName": cmd.RequiredArgs.ResourceName,
		"OrgName":      cmd.Config.TargetedOrganization().Name,
		"SpaceName":    cmd.Config.TargetedSpace().Name,
		"User":         cmd.username,
	})
}
