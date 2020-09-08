package v7

import (
	"strconv"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/ui"
)

//go:generate counterfeiter . RevisionsActor

type RevisionsActor interface {
	GetRevisionsByApplicationNameAndSpace(appName string, spaceGUID string) ([]resources.Revision, v7action.Warnings, error)
}

type RevisionsCommand struct {
	RequiredArgs flag.EnvironmentArgs `positional-args:"yes"`
	usage        interface{}          `usage:"CF_NAME revisions APP_NAME"`

	BaseCommand
}

func (cmd RevisionsCommand) Execute(_ []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	appName := cmd.RequiredArgs.AppName
	cmd.UI.DisplayTextWithFlavor("Getting revisions for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   appName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})

	app, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(appName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	appGUID := app.GUID
	revisionsFeature, warnings, err := cmd.Actor.GetAppFeature(appGUID, "revisions")
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if !revisionsFeature.Enabled {
		cmd.UI.DisplayWarning("Warning: Revisions for app '{{.AppName}}' are disabled. Updates to the app will not create new revisions.",
			map[string]interface{}{
				"AppName": appName,
			})
	}

	cmd.UI.DisplayNewline()

	revisions, warnings, err := cmd.Actor.GetRevisionsByApplicationNameAndSpace(
		appName,
		cmd.Config.TargetedSpace().GUID,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(revisions) == 0 {
		cmd.UI.DisplayText("No revisions found")
		return nil
	}

	table := [][]string{{
		"revision",
		"description",
		"deployable",
		"revision guid",
		"created at",
	}}
	for _, revision := range revisions {
		table = append(table,
			[]string{strconv.Itoa(revision.Version),
				revision.Description,
				strconv.FormatBool(revision.Deployable),
				revision.GUID,
				revision.CreatedAt,
			})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}
