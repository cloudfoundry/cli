package v7

import (
	"fmt"
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/ui"
	"code.cloudfoundry.org/clock"
)

type ResourceType string

const (
	App       ResourceType = "app"
	Buildpack ResourceType = "buildpack"
	Domain    ResourceType = "domain"
	Org       ResourceType = "org"
	Space     ResourceType = "space"
	Stack     ResourceType = "stack"
)

//go:generate counterfeiter . LabelsActor

type LabelsActor interface {
	GetApplicationLabels(appName string, spaceGUID string) (map[string]types.NullString, v7action.Warnings, error)
	GetOrganizationLabels(orgName string) (map[string]types.NullString, v7action.Warnings, error)
	GetSpaceLabels(spaceName string, orgGUID string) (map[string]types.NullString, v7action.Warnings, error)
	GetBuildpackLabels(buildpackName string, buildpackStack string) (map[string]types.NullString, v7action.Warnings, error)
	GetStackLabels(stackName string) (map[string]types.NullString, v7action.Warnings, error)
}

type LabelsCommand struct {
	RequiredArgs    flag.LabelsArgs `positional-args:"yes"`
	BuildpackStack  string          `long:"stack" short:"s" description:"Specify stack to disambiguate buildpacks with the same name"`
	usage           interface{}     `usage:"CF_NAME labels RESOURCE RESOURCE_NAME\n\nEXAMPLES:\n   cf labels app dora\n   cf labels org business\n   cf labels buildpack go_buildpack --stack cflinuxfs3 \n\nRESOURCES:\n   app\n   buildpack\n   org\n   space\n   stack"`
	relatedCommands interface{}     `related_commands:"set-label, unset-label"`
	UI              command.UI
	Config          command.Config
	SharedActor     command.SharedActor
	Actor           LabelsActor
}

func (cmd *LabelsCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd LabelsCommand) Execute(args []string) error {
	var (
		labels   map[string]types.NullString
		warnings v7action.Warnings
	)
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
		labels, warnings, err = cmd.fetchAppLabels(username)
	case Buildpack:
		labels, warnings, err = cmd.fetchBuildpackLabels(username)
	case Org:
		labels, warnings, err = cmd.fetchOrgLabels(username)
	case Space:
		labels, warnings, err = cmd.fetchSpaceLabels(username)
	case Stack:
		labels, warnings, err = cmd.fetchStackLabels(username)
	default:
		err = fmt.Errorf("Unsupported resource type of '%s'", cmd.RequiredArgs.ResourceType)
	}
	cmd.UI.DisplayWarningsV7(warnings)
	if err != nil {
		return err
	}

	cmd.printLabels(labels)
	return nil
}

func (cmd LabelsCommand) fetchAppLabels(username string) (map[string]types.NullString, v7action.Warnings, error) {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return nil, nil, err
	}

	cmd.UI.DisplayTextWithFlavor("Getting labels for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.ResourceName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  username,
	})

	cmd.UI.DisplayNewline()
	return cmd.Actor.GetApplicationLabels(cmd.RequiredArgs.ResourceName, cmd.Config.TargetedSpace().GUID)
}

func (cmd LabelsCommand) fetchOrgLabels(username string) (map[string]types.NullString, v7action.Warnings, error) {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return nil, nil, err
	}

	cmd.UI.DisplayTextWithFlavor("Getting labels for org {{.OrgName}} as {{.Username}}...", map[string]interface{}{
		"OrgName":  cmd.RequiredArgs.ResourceName,
		"Username": username,
	})

	cmd.UI.DisplayNewline()

	return cmd.Actor.GetOrganizationLabels(cmd.RequiredArgs.ResourceName)
}

func (cmd LabelsCommand) fetchSpaceLabels(username string) (map[string]types.NullString, v7action.Warnings, error) {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return nil, nil, err
	}

	cmd.UI.DisplayTextWithFlavor("Getting labels for space {{.SpaceName}} in org {{.OrgName}} as {{.Username}}...", map[string]interface{}{
		"SpaceName": cmd.RequiredArgs.ResourceName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"Username":  username,
	})

	cmd.UI.DisplayNewline()

	return cmd.Actor.GetSpaceLabels(cmd.RequiredArgs.ResourceName, cmd.Config.TargetedOrganization().GUID)
}

func (cmd LabelsCommand) fetchStackLabels(username string) (map[string]types.NullString, v7action.Warnings, error) {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return nil, nil, err
	}

	cmd.UI.DisplayTextWithFlavor("Getting labels for stack {{.StackName}} as {{.Username}}...", map[string]interface{}{
		"StackName": cmd.RequiredArgs.ResourceName,
		"Username":  username,
	})

	cmd.UI.DisplayNewline()

	return cmd.Actor.GetStackLabels(cmd.RequiredArgs.ResourceName)
}

func (cmd LabelsCommand) fetchBuildpackLabels(username string) (map[string]types.NullString, v7action.Warnings, error) {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return nil, nil, err
	}

	var template string
	if cmd.BuildpackStack != "" {
		template = "Getting labels for %s {{.ResourceName}} with stack {{.StackName}} as {{.User}}..."
	} else {
		template = "Getting labels for %s {{.ResourceName}} as {{.User}}..."
	}
	preFlavoringText := fmt.Sprintf(template, cmd.RequiredArgs.ResourceType)
	cmd.UI.DisplayTextWithFlavor(
		preFlavoringText,
		map[string]interface{}{
			"ResourceName": cmd.RequiredArgs.ResourceName,
			"StackName":    cmd.BuildpackStack,
			"User":         username,
		},
	)

	cmd.UI.DisplayNewline()

	return cmd.Actor.GetBuildpackLabels(cmd.RequiredArgs.ResourceName, cmd.BuildpackStack)
}

func (cmd LabelsCommand) canonicalResourceTypeForName() ResourceType {
	return ResourceType(strings.ToLower(cmd.RequiredArgs.ResourceType))
}

func (cmd LabelsCommand) printLabels(labels map[string]types.NullString) {
	if len(labels) == 0 {
		cmd.UI.DisplayText("No labels found.")
		return
	}

	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	table := [][]string{
		{
			cmd.UI.TranslateText("key"),
			cmd.UI.TranslateText("value"),
		},
	}

	for _, key := range keys {
		table = append(table, []string{key, labels[key].Value})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
}

func (cmd LabelsCommand) validateFlags() error {
	if cmd.BuildpackStack != "" && cmd.canonicalResourceTypeForName() != Buildpack {
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				cmd.RequiredArgs.ResourceType, "--stack, -s",
			},
		}
	}
	return nil
}
