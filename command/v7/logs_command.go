package v7

import (
	"os"
	"os/signal"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

type LogsCommand struct {
	BaseCommand

	RequiredArgs    flag.AppName `positional-args:"yes"`
	Recent          bool         `long:"recent" description:"Dump recent logs instead of tailing"`
	usage           interface{}  `usage:"CF_NAME logs APP_NAME"`
	relatedCommands interface{}  `related_commands:"app, apps, ssh"`

	CC_Client      *ccv3.Client
	LogCacheClient sharedaction.LogCacheClient
}

func (cmd *LogsCommand) Setup(config command.Config, ui command.UI) error {
	err := cmd.BaseCommand.Setup(config, ui)
	if err != nil {
		return err
	}

	ccClient, _, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}

	cmd.LogCacheClient = command.NewLogCacheClient(ccClient.Info.LogCache(), config, ui)
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

func (cmd LogsCommand) handleLogErr(logErr error) {
	switch logErr.(type) {
	case actionerror.LogCacheTimeoutError:
		cmd.UI.DisplayWarning("timeout connecting to log server, no log will be shown")
	default:
		cmd.UI.DisplayWarning("Failed to retrieve logs from Log Cache: {{.Error}}", map[string]interface{}{
			"Error": logErr,
		})
	}
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

	defer stopStreaming()
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
			cmd.handleLogErr(logErr)
		case <-c:
			return nil
		}

		if messagesClosed && errLogsClosed {
			break
		}
	}

	return nil
}
