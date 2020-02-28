package v6

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . LogsActor

type LogsActor interface {
	GetRecentLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client sharedaction.LogCacheClient) ([]sharedaction.LogMessage, v2action.Warnings, error)
	GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc, v2action.Warnings, error)
	ScheduleTokenRefresh(func(time.Duration) <-chan time.Time, chan struct{}, chan struct{}) (<-chan error, error)
}

type LogsCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	Recent          bool         `long:"recent" description:"Dump recent logs instead of tailing"`
	usage           interface{}  `usage:"CF_NAME logs APP_NAME"`
	relatedCommands interface{}  `related_commands:"app, apps, ssh"`

	UI             command.UI
	Config         command.Config
	SharedActor    command.SharedActor
	Actor          LogsActor
	LogCacheClient sharedaction.LogCacheClient
}

func (cmd *LogsCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui)
	if err != nil {
		return err
	}

	ccClientV3, _, err := shared.NewV3BasedClients(config, ui, true)
	if err != nil {
		return err
	}

	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	v3Actor := v3action.NewActor(ccClientV3, config, nil, nil)
	logcacheURL := v3Actor.LogCacheURL()
	cmd.LogCacheClient = command.NewLogCacheClient(logcacheURL, config, ui)

	return nil
}

func (cmd LogsCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Retrieving logs for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"AppName":   cmd.RequiredArgs.AppName,
			"OrgName":   cmd.Config.TargetedOrganization().Name,
			"SpaceName": cmd.Config.TargetedSpace().Name,
			"Username":  user.Name,
		})
	cmd.UI.DisplayNewline()

	if cmd.Recent {
		return cmd.displayRecentLogs()
	}

	stop := make(chan struct{})
	stoppedRefreshing := make(chan struct{})
	stoppedOutputtingRefreshErrors := make(chan struct{})
	err = cmd.refreshTokenPeriodically(stop, stoppedRefreshing, stoppedOutputtingRefreshErrors)
	if err != nil {
		return err
	}

	err = cmd.streamLogs()

	close(stop)
	<-stoppedRefreshing
	<-stoppedOutputtingRefreshErrors

	return err
}

func (cmd LogsCommand) displayRecentLogs() error {
	messages, warnings, err := cmd.Actor.GetRecentLogsForApplicationByNameAndSpace(
		cmd.RequiredArgs.AppName,
		cmd.Config.TargetedSpace().GUID,
		cmd.LogCacheClient,
	)

	for _, message := range messages {
		cmd.UI.DisplayLogMessage(message, true)
	}

	cmd.UI.DisplayWarnings(warnings)
	return err
}

func (cmd LogsCommand) refreshTokenPeriodically(
	stop chan struct{},
	stoppedRefreshing chan struct{},
	stoppedOutputtingRefreshErrors chan struct{}) error {

	tokenRefreshErrors, err := cmd.Actor.ScheduleTokenRefresh(time.After, stop, stoppedRefreshing)
	if err != nil {
		return err
	}

	go func() {
		defer close(stoppedOutputtingRefreshErrors)

		for {
			select {
			case err := <-tokenRefreshErrors:
				cmd.UI.DisplayError(err)
			case <-stop:
				return
			}
		}
	}()

	return nil
}

func (cmd LogsCommand) streamLogs() error {
	messages, logErrs, stopStreaming, warnings, err := cmd.Actor.GetStreamingLogsForApplicationByNameAndSpace(
		cmd.RequiredArgs.AppName,
		cmd.Config.TargetedSpace().GUID,
		cmd.LogCacheClient,
	)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	var messagesClosed, errLogsClosed bool
	for {
		select {
		case message, ok := <-messages:
			if !ok {
				messagesClosed = true
				break
			}

			cmd.UI.DisplayLogMessage(message, true)
		case logErr, ok := <-logErrs:
			if !ok {
				errLogsClosed = true
				break
			}
			stopStreaming()
			return fmt.Errorf("Failed to retrieve logs from Log Cache: %s", logErr)
		case <-c:
			stopStreaming()
			return nil
		}

		if messagesClosed && errLogsClosed {
			break
		}
	}

	return nil
}
