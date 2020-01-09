package v7

import (
	"errors"
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/types"
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

type ActionType string

const (
	Unset ActionType = "Removing"
	Set   ActionType = "Setting"
)

type TargetResource struct {
	ResourceType   string
	ResourceName   string
	BuildpackStack string
}

type LabelUpdater struct {
	targetResource TargetResource
	labels         map[string]types.NullString

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SetLabelActor

	Username string
	Action   ActionType
}

func (cmd *LabelUpdater) Execute(targetResource TargetResource, labels map[string]types.NullString) error {
	cmd.targetResource = targetResource
	cmd.labels = labels

	var err error
	cmd.Username, err = cmd.Config.CurrentUserName()
	if err != nil {
		return err
	}

	err = cmd.validateFlags()
	if err != nil {
		return err
	}

	resourceTypeString := strings.ToLower(cmd.targetResource.ResourceType)
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
		err = errors.New(cmd.UI.TranslateText("Unsupported resource type of '{{.ResourceType}}'", map[string]interface{}{"ResourceType": cmd.targetResource.ResourceType}))
	}

	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}

func (cmd LabelUpdater) validateFlags() error {
	resourceTypeString := strings.ToLower(cmd.targetResource.ResourceType)
	if cmd.targetResource.BuildpackStack != "" && ResourceType(resourceTypeString) != Buildpack {
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				cmd.targetResource.ResourceType, "--stack, -s",
			},
		}
	}
	return nil
}

func (cmd LabelUpdater) executeApp() error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	cmd.displayMessageWithOrgAndSpace()

	warnings, err := cmd.Actor.UpdateApplicationLabelsByApplicationName(cmd.targetResource.ResourceName, cmd.Config.TargetedSpace().GUID, cmd.labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd LabelUpdater) executeDomain() error {
	cmd.displayMessage()

	warnings, err := cmd.Actor.UpdateDomainLabelsByDomainName(cmd.targetResource.ResourceName, cmd.labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd LabelUpdater) executeRoute() error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	cmd.displayMessageWithOrgAndSpace()

	warnings, err := cmd.Actor.UpdateRouteLabels(cmd.targetResource.ResourceName, cmd.Config.TargetedSpace().GUID, cmd.labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd LabelUpdater) executeBuildpack() error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	var template string
	if cmd.targetResource.BuildpackStack == "" {
		template = "{{.Action}} label(s) for buildpack {{.ResourceName}} as {{.User}}..."
	} else {
		template = "{{.Action}} label(s) for buildpack {{.ResourceName}} with stack {{.StackName}} as {{.User}}..."
	}

	cmd.UI.DisplayTextWithFlavor(template, map[string]interface{}{
		"Action":       fmt.Sprintf("%s", cmd.Action),
		"ResourceName": cmd.targetResource.ResourceName,
		"StackName":    cmd.targetResource.BuildpackStack,
		"User":         cmd.Username,
	})

	warnings, err := cmd.Actor.UpdateBuildpackLabelsByBuildpackNameAndStack(cmd.targetResource.ResourceName, cmd.targetResource.BuildpackStack, cmd.labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd LabelUpdater) executeOrg() error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	cmd.displayMessage()

	warnings, err := cmd.Actor.UpdateOrganizationLabelsByOrganizationName(cmd.targetResource.ResourceName, cmd.labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd LabelUpdater) executeServiceBroker() error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	cmd.displayMessage()
	warnings, err := cmd.Actor.UpdateServiceBrokerLabelsByServiceBrokerName(cmd.targetResource.ResourceName, cmd.labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd LabelUpdater) executeSpace() error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("{{.Action}} label(s) for space {{.ResourceName}} in org {{.OrgName}} as {{.User}}...", map[string]interface{}{
		"Action":       cmd.Action,
		"ResourceName": cmd.targetResource.ResourceName,
		"OrgName":      cmd.Config.TargetedOrganization().Name,
		"User":         cmd.Username,
	})

	warnings, err := cmd.Actor.UpdateSpaceLabelsBySpaceName(cmd.targetResource.ResourceName, cmd.Config.TargetedOrganization().GUID, cmd.labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd LabelUpdater) executeStack() error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	cmd.displayMessage()

	warnings, err := cmd.Actor.UpdateStackLabelsByStackName(cmd.targetResource.ResourceName, cmd.labels)

	cmd.UI.DisplayWarnings(warnings)

	return err
}

func (cmd LabelUpdater) displayMessage() {
	cmd.UI.DisplayTextWithFlavor("{{.Action}} label(s) for {{.ResourceType}} {{.ResourceName}} as {{.User}}...", map[string]interface{}{
		"Action":       cmd.Action,
		"ResourceType": strings.ToLower(cmd.targetResource.ResourceType),
		"ResourceName": cmd.targetResource.ResourceName,
		"User":         cmd.Username,
	})
}

func (cmd LabelUpdater) displayMessageWithOrgAndSpace() {
	cmd.UI.DisplayTextWithFlavor("{{.Action}} label(s) for {{.ResourceType}} {{.ResourceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.User}}...", map[string]interface{}{
		"Action":       cmd.Action,
		"ResourceType": strings.ToLower(cmd.targetResource.ResourceType),
		"ResourceName": cmd.targetResource.ResourceName,
		"OrgName":      cmd.Config.TargetedOrganization().Name,
		"SpaceName":    cmd.Config.TargetedSpace().Name,
		"User":         cmd.Username,
	})
}
