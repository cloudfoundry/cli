package v7

import (
	"fmt"
	"strconv"

	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/resources"
)

type RevisionCommand struct {
	BaseCommand
	usage           interface{}   `usage:"CF_NAME revision APP_NAME [--version VERSION]"`
	RequiredArgs    flag.AppName  `positional-args:"yes"`
	Version         flag.Revision `long:"version" required:"true" description:"The integer representing the specific revision to show"`
	relatedCommands interface{}   `related_commands:"revisions, rollback"`
}

func (cmd RevisionCommand) Execute(_ []string) error {
	cmd.UI.DisplayWarning(command.ExperimentalWarning)
	cmd.UI.DisplayNewline()
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

	app, _, _ := cmd.Actor.GetApplicationByNameAndSpace(appName, cmd.Config.TargetedSpace().GUID)
	deployedRevisions, _, _ := cmd.Actor.GetApplicationRevisionsDeployed(app.GUID)
	revisions, _, _ := cmd.Actor.GetRevisionsByApplicationNameAndSpace(
		appName,
		cmd.Config.TargetedSpace().GUID,
	)
	revision, _ := cmd.getSelectedRevision(revisions)
	deployed := revisionDeployed(revision, deployedRevisions)

	cmd.displayRevision(revision, deployed)
	return nil
}

func (cmd RevisionCommand) getSelectedRevision(revisions []resources.Revision) (resources.Revision, error) {
	for _, revision := range revisions {
		if revision.Version == cmd.Version.Value {
			return revision, nil
		}
	}
	return resources.Revision{}, nil
}

func (cmd RevisionCommand) displayRevision(revision resources.Revision, deployed bool) {
	cmd.displayBasicRevisionInfo(revision, deployed)
	cmd.UI.DisplayNewline()
	cmd.displayEnvVarGroupForRevision(revision)
	cmd.UI.DisplayNewline()
}

func (cmd RevisionCommand) displayBasicRevisionInfo(revision resources.Revision, deployed bool) {
	keyValueTable := [][]string{
		{"revision:", fmt.Sprintf("%d", cmd.Version.Value)},
		{"deployed:", strconv.FormatBool(deployed)},
		{"description:", revision.Description},
		{"deployable:", strconv.FormatBool(revision.Deployable)},
		{"revision GUID:", revision.GUID},
		{"droplet GUID:", revision.Droplet.GUID},
		{"created on:", revision.CreatedAt},
	}
	cmd.UI.DisplayKeyValueTable("", keyValueTable, 3)
}

func (cmd RevisionCommand) displayEnvVarGroupForRevision(revision resources.Revision) {
	envVarsURL, present := revision.Links["environment_variables"]
	if present == true {
		envVarGroup, _, _ := cmd.Actor.GetEnvironmentVariableGroupByRevision(envVarsURL.HREF)
		envVarTable := [][]string{}
		for k, v := range envVarGroup {
			envVarTable = append(envVarTable, []string{fmt.Sprintf("%s:", k), v.Value})
		}
		cmd.UI.DisplayText("application environment variables:")
		cmd.UI.DisplayKeyValueTable("", envVarTable, 3)
	} else {
		cmd.UI.DisplayWarning("Unable to retrieve environment variables for revision.")
	}
}

func revisionDeployed(revision resources.Revision, deployedRevisions []resources.Revision) bool {
	for _, deployedRevision := range deployedRevisions {
		if revision.GUID == deployedRevision.GUID {
			return true
		}
	}
	return false
}
