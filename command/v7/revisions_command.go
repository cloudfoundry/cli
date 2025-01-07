package v7

import (
	"fmt"
	"strconv"

	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/command/flag"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/util/ui"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . RevisionsActor

type RevisionsActor interface {
	GetRevisionsByApplicationNameAndSpace(appName string, spaceGUID string) ([]resources.Revision, v7action.Warnings, error)
}

type RevisionsCommand struct {
	RequiredArgs flag.EnvironmentArgs `positional-args:"yes"`
	usage        interface{}          `usage:"CF_NAME revisions APP_NAME"`

	BaseCommand
	relatedCommands interface{} `related_commands:"revision, rollback"`
}

func (cmd RevisionsCommand) Execute(_ []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
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

	if app.Stopped() {
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText(fmt.Sprintf("Info: this app is in a stopped state. It is not possible to determine which revision is currently deployed."))
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

	revisionsDeployed, warnings, err := cmd.Actor.GetApplicationRevisionsDeployed(appGUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(revisionsDeployed) > 1 {
		cmd.UI.DisplayText("Info: this app is in the middle of a rolling deployment. More than one revision is deployed.")
		cmd.UI.DisplayNewline()
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
			[]string{decorateVersionWithDeployed(revision, revisionsDeployed),
				revision.Description,
				strconv.FormatBool(revision.Deployable),
				revision.GUID,
				revision.CreatedAt,
			})
	}

	if len(revisionsDeployed) > 1 {
		cmd.UI.DisplayText("Info: this app is in the middle of a deployment. More than one revision is deployed.")
		cmd.UI.DisplayNewline()
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}

func decorateVersionWithDeployed(revision resources.Revision, deployedRevisions []resources.Revision) string {
	for _, revDeployed := range deployedRevisions {
		if revDeployed.GUID == revision.GUID {
			return strconv.Itoa(revision.Version) + "(deployed)"
		}
	}
	return strconv.Itoa(revision.Version)
}
