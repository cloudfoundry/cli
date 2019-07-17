package v7

import (
	"fmt"
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/ui"
)

type ResourceType string

const (
	App       ResourceType = "app"
	Buildpack ResourceType = "buildpack"
	Org       ResourceType = "org"
	Space     ResourceType = "space"
)

//go:generate counterfeiter . LabelsActor

type LabelsActor interface {
	GetApplicationLabels(appName string, spaceGUID string) (map[string]types.NullString, v7action.Warnings, error)
	GetOrganizationLabels(orgName string) (map[string]types.NullString, v7action.Warnings, error)
	GetSpaceLabels(spaceName string, orgGUID string) (map[string]types.NullString, v7action.Warnings, error)
	GetBuildpackLabels(buildpackName string, buildpackStack string) (map[string]types.NullString, v7action.Warnings, error)
}

type LabelsCommand struct {
	RequiredArgs flag.LabelsArgs `positional-args:"yes"`
	StackName    string          `long:"stack" short:"s" description:"required when more than one buildpack has the same name"`
	usage        interface{}     `usage:"CF_NAME labels RESOURCE RESOURCE_NAME\n\nEXAMPLES:\n   cf labels app dora \n\nRESOURCES:\n   app\n   buildpack\n   org\n   space\n\nSEE ALSO:\n   set-label, unset-label"`
	UI           command.UI
	Config       command.Config
	SharedActor  command.SharedActor
	Actor        LabelsActor
}

func (cmd *LabelsCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd LabelsCommand) Execute(args []string) error {
	var (
		labels   map[string]types.NullString
		warnings v7action.Warnings
	)
	username, err := cmd.Config.CurrentUserName()
	if err != nil {
		return err
	}

	resourceTypeString := strings.ToLower(cmd.RequiredArgs.ResourceType)
	switch ResourceType(resourceTypeString) {
	case App:
		labels, warnings, err = cmd.fetchAppLabels(username)
	case Org:
		labels, warnings, err = cmd.fetchOrgLabels(username)
	case Space:
		labels, warnings, err = cmd.fetchSpaceLabels(username)
	case Buildpack:
		labels, warnings, err = cmd.fetchBuildpackLabels(username)
	default:
		err = fmt.Errorf("Unsupported resource type of '%s'", cmd.RequiredArgs.ResourceType)
	}
	cmd.UI.DisplayWarnings(warnings)
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

func (cmd LabelsCommand) fetchBuildpackLabels(username string) (map[string]types.NullString, v7action.Warnings, error) {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return nil, nil, err
	}

	cmd.UI.DisplayTextWithFlavor("Getting labels for buildpack {{.BuildpackName}} as {{.Username}}...", map[string]interface{}{
		"BuildpackName": cmd.RequiredArgs.ResourceName,
		"Username":      username,
	})

	cmd.UI.DisplayNewline()

	return cmd.Actor.GetBuildpackLabels(cmd.RequiredArgs.ResourceName, cmd.StackName)
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
