package v7

import (
	"fmt"
	"strconv"

	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/types"
)

type RevisionCommand struct {
	BaseCommand
	usage           interface{}   `usage:"CF_NAME revision APP_NAME [--version VERSION]"`
	RequiredArgs    flag.AppName  `positional-args:"yes"`
	Version         flag.Revision `long:"version" description:"The integer representing the specific revision to show"`
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
	if cmd.Version.Value > 0 {
		cmd.UI.DisplayTextWithFlavor("Showing revision {{.Version}} for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
			"AppName":   appName,
			"OrgName":   cmd.Config.TargetedOrganization().Name,
			"SpaceName": cmd.Config.TargetedSpace().Name,
			"Username":  user.Name,
			"Version":   cmd.Version.Value,
		})
	} else {
		cmd.UI.DisplayTextWithFlavor("Showing revisions for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
			"AppName":   appName,
			"OrgName":   cmd.Config.TargetedOrganization().Name,
			"SpaceName": cmd.Config.TargetedSpace().Name,
			"Username":  user.Name,
		})
	}

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

	if cmd.Version.Value > 0 {
		revision, warnings, err := cmd.Actor.GetRevisionByApplicationAndVersion(
			app.GUID,
			cmd.Version.Value,
		)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
		isDeployed := cmd.revisionDeployed(revision, deployedRevisions)

		err = cmd.displayRevisionInfo(revision, isDeployed)
		if err != nil {
			return err
		}
	} else {
		for _, deployedRevision := range deployedRevisions {
			err = cmd.displayRevisionInfo(deployedRevision, true)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (cmd RevisionCommand) displayRevisionInfo(revision resources.Revision, isDeployed bool) error {
	cmd.displayBasicRevisionInfo(revision, isDeployed)
	cmd.UI.DisplayNewline()

	cmd.UI.DisplayHeader("labels:")
	cmd.displayMetaData(revision.Metadata.Labels)

	cmd.UI.DisplayHeader("annotations:")
	cmd.displayMetaData(revision.Metadata.Annotations)

	cmd.UI.DisplayHeader("application environment variables:")
	envVars, isPresent, warnings, err := cmd.Actor.GetEnvironmentVariableGroupByRevision(revision)
	if err != nil {
		return err
	}
	cmd.UI.DisplayWarnings(warnings)
	if isPresent {
		cmd.displayEnvVarGroup(envVars)
		cmd.UI.DisplayNewline()
	}
	return nil
}

func (cmd RevisionCommand) displayBasicRevisionInfo(revision resources.Revision, isDeployed bool) {
	keyValueTable := [][]string{
		{"revision:", fmt.Sprintf("%d", revision.Version)},
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
	if len(data) > 0 {
		tableData := [][]string{}
		for k, v := range data {
			tableData = append(tableData, []string{fmt.Sprintf("%s:", k), v.Value})
		}
		cmd.UI.DisplayKeyValueTable("", tableData, 3)
		cmd.UI.DisplayNewline()
	}
}

func (cmd RevisionCommand) revisionDeployed(revision resources.Revision, deployedRevisions []resources.Revision) bool {
	for _, deployedRevision := range deployedRevisions {
		if revision.GUID == deployedRevision.GUID {
			return true
		}
	}
	return false
}
