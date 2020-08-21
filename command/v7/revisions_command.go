package v7

import (
	"strconv"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/util/ui"
)

//go:generate counterfeiter . RevisionsActor

type RevisionsActor interface {
	GetRevisionsByApplicationNameAndSpace(appName string, spaceGUID string) (v7action.Revisions, v7action.Warnings, error)
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
	cmd.UI.DisplayOK()

	revisions, warnings, err := cmd.Actor.GetRevisionsByApplicationNameAndSpace(
		appName,
		cmd.Config.TargetedSpace().GUID,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	table := [][]string{{
		"version",
		"guid",
		"description",
		"deployable",
		"created at",
	}}
	for _, revision := range revisions {
		table = append(table, []string{strconv.Itoa(revision.Version), revision.GUID, revision.Description, strconv.FormatBool(revision.Deployable), revision.CreatedAt})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}
