package v7

import (
	"fmt"
	"strconv"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
)

type RevisionCommand struct {
	BaseCommand
	usage           interface{}   `usage:"CF_NAME revision APP_NAME [--version VERSION]"`
	RequiredArgs    flag.AppName  `positional-args:"yes"`
	Version         flag.Revision `long:"version" required:"true" description:"The integer representing the specific revision to show"`
	relatedCommands interface{}   `related_commands:"revisions, rollback"`
}

func (cmd RevisionCommand) Execute(_ []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	appName := cmd.RequiredArgs.AppName
	cmd.UI.DisplayTextWithFlavor("Showing revision {{.Version}} for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   appName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
		"Version":   cmd.Version.Value,
	})
	cmd.UI.DisplayNewline()

	app, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(appName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	deployedRevisions, warnings, err := cmd.Actor.GetApplicationRevisionsDeployed(app.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	revision, warnings, err := cmd.Actor.GetRevisionByApplicationAndVersion(
		app.GUID,
		cmd.Version.Value,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	isDeployed := revisionDeployed(revision, deployedRevisions)

	cmd.displayBasicRevisionInfo(revision, isDeployed)
	cmd.UI.DisplayNewline()

	cmd.UI.DisplayHeader("labels:")
	labels := revision.Metadata.Labels
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(labels) > 0 {
		cmd.displayMetaData(labels)
		cmd.UI.DisplayNewline()
	}

	cmd.UI.DisplayHeader("annotations:")
	annotations := revision.Metadata.Annotations
	cmd.UI.DisplayWarnings(warnings)

	if len(annotations) > 0 {
		cmd.displayMetaData(annotations)
		cmd.UI.DisplayNewline()
	}

	cmd.UI.DisplayHeader("application environment variables:")
	envVars, isPresent, warnings, _ := cmd.Actor.GetEnvironmentVariableGroupByRevision(revision)
	cmd.UI.DisplayWarnings(warnings)
	if isPresent {
		cmd.displayEnvVarGroup(envVars)
		cmd.UI.DisplayNewline()
	}

	return nil
}

func (cmd RevisionCommand) displayBasicRevisionInfo(revision resources.Revision, isDeployed bool) {
	keyValueTable := [][]string{
		{"revision:", fmt.Sprintf("%d", cmd.Version.Value)},
		{"deployed:", strconv.FormatBool(isDeployed)},
		{"description:", revision.Description},
		{"deployable:", strconv.FormatBool(revision.Deployable)},
		{"revision GUID:", revision.GUID},
		{"droplet GUID:", revision.Droplet.GUID},
		{"created on:", revision.CreatedAt},
	}
	cmd.UI.DisplayKeyValueTable("", keyValueTable, 3)
}

func (cmd RevisionCommand) displayEnvVarGroup(envVarGroup v7action.EnvironmentVariableGroup) {
	envVarTable := [][]string{}
	for k, v := range envVarGroup {
		envVarTable = append(envVarTable, []string{fmt.Sprintf("%s:", k), v.Value})
	}
	cmd.UI.DisplayKeyValueTable("", envVarTable, 3)
}

func (cmd RevisionCommand) displayMetaData(data map[string]types.NullString) {
	tableData := [][]string{}
	for k, v := range data {
		tableData = append(tableData, []string{fmt.Sprintf("%s:", k), v.Value})
	}
	cmd.UI.DisplayKeyValueTable("", tableData, 3)

}

func revisionDeployed(revision resources.Revision, deployedRevisions []resources.Revision) bool {
	for _, deployedRevision := range deployedRevisions {
		if revision.GUID == deployedRevision.GUID {
			return true
		}
	}
	return false
}
