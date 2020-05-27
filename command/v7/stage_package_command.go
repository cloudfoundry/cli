package v7

import (
	"strings"
	"time"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/resources"
)

type StagePackageCommand struct {
	BaseCommand

	RequiredArgs    flag.AppName `positional-args:"yes"`
	PackageGUID     string       `long:"package-guid" description:"The guid of the package to stage (default: latest ready package)"`
	usage           interface{}  `usage:"CF_NAME stage-package APP_NAME [--package-guid PACKAGE_GUID]"`
	relatedCommands interface{}  `related_commands:"app, create-package, droplets, packages, push, set-droplet"`

	envCFStagingTimeout interface{} `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for staging, in minutes" environmentDefault:"15"`

	LogCacheClient sharedaction.LogCacheClient
}

func (cmd *StagePackageCommand) Setup(config command.Config, ui command.UI) error {
	err := cmd.BaseCommand.Setup(config, ui)
	if err != nil {
		return err
	}

	logCacheEndpoint, _, err := cmd.Actor.GetLogCacheEndpoint()
	if err != nil {
		return err
	}
	cmd.LogCacheClient = command.NewLogCacheClient(logCacheEndpoint, config, ui)

	return nil
}

func (cmd StagePackageCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Staging package for {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})

	packageGUID := cmd.PackageGUID

	if packageGUID == "" {
		app, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		pkg, warnings, err := cmd.Actor.GetNewestReadyPackageForApplication(app)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		packageGUID = pkg.GUID
	}

	logStream, logErrStream, stopLogStreamFunc, logWarnings, logErr := cmd.Actor.GetStreamingLogsForApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, cmd.LogCacheClient)
	cmd.UI.DisplayWarnings(logWarnings)
	if logErr != nil {
		return logErr
	}
	defer stopLogStreamFunc()

	dropletStream, warningsStream, errStream := cmd.Actor.StagePackage(
		packageGUID,
		cmd.RequiredArgs.AppName,
		cmd.Config.TargetedSpace().GUID,
	)

	var droplet resources.Droplet
	droplet, err = shared.PollStage(dropletStream, warningsStream, errStream, logStream, logErrStream, cmd.UI)
	if err != nil {
		return err
	}

	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText("Package staged")

	t, err := time.Parse(time.RFC3339, droplet.CreatedAt)
	if err != nil {
		return err
	}

	table := [][]string{
		{cmd.UI.TranslateText("droplet guid:"), droplet.GUID},
		{cmd.UI.TranslateText("state:"), strings.ToLower(string(droplet.State))},
		{cmd.UI.TranslateText("created:"), cmd.UI.UserFriendlyDate(t)},
	}

	cmd.UI.DisplayKeyValueTable("", table, 3)
	return nil
}
