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

	app, _, _ := cmd.Actor.GetApplicationByNameAndSpace(appName, cmd.Config.TargetedSpace().GUID)
	deployedRevisions, _, _ := cmd.Actor.GetApplicationRevisionsDeployed(app.GUID)
	revisions, _, _ := cmd.Actor.GetRevisionsByApplicationNameAndSpace(
		appName,
		cmd.Config.TargetedSpace().GUID,
	)
	for _, revision := range revisions {
		if revision.Version == cmd.Version.Value {
			deployed := revisionDeployed(revision, deployedRevisions)
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
	}
	return nil
}

func revisionDeployed(revision resources.Revision, deployedRevisions []resources.Revision) bool {
	for _, deployedRevision := range deployedRevisions {
		if revision.GUID == deployedRevision.GUID {
			return true
		}
	}
	return false
}
