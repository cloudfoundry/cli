package v7

import (
	"context"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . StageActor

type StageActor interface {
	GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client v7action.LogCacheClient) (<-chan v7action.LogMessage, <-chan error, context.CancelFunc, v7action.Warnings, error)
	StagePackage(packageGUID, appName, spaceGUID string) (<-chan v7action.Droplet, <-chan v7action.Warnings, <-chan error)
}

type StageCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	PackageGUID     string       `long:"package-guid" description:"The guid of the package to stage" required:"true"`
	usage           interface{}  `usage:"CF_NAME stage APP_NAME --package-guid PACKAGE_GUID"`
	relatedCommands interface{}  `related_commands:"app, create-package, droplets, packages, push, set-droplet, stage"`

	envCFStagingTimeout interface{} `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for staging, in minutes" environmentDefault:"15"`

	UI             command.UI
	Config         command.Config
	LogCacheClient v7action.LogCacheClient
	SharedActor    command.SharedActor
	Actor          StageActor
}

func (cmd *StageCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, _, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}

	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil, clock.NewClock())
	cmd.LogCacheClient = shared.NewLogCacheClient(ccClient.Info.LogCache(), config, ui)

	return nil
}

func (cmd StageCommand) Execute(args []string) error {
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

	logStream, logErrStream, stopLogStreamFunc, logWarnings, logErr := cmd.Actor.GetStreamingLogsForApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, cmd.LogCacheClient)
	cmd.UI.DisplayWarningsV7(logWarnings)
	if logErr != nil {
		return logErr
	}
	defer stopLogStreamFunc()

	dropletStream, warningsStream, errStream := cmd.Actor.StagePackage(
		cmd.PackageGUID,
		cmd.RequiredArgs.AppName,
		cmd.Config.TargetedSpace().GUID,
	)

	var droplet v7action.Droplet
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
