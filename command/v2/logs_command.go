package v2

import (
	"github.com/cloudfoundry/noaa/consumer"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . LogsActor

type LogsActor interface {
	GetRecentLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client v2action.NOAAClient, config v2action.Config) ([]v2action.LogMessage, v2action.Warnings, error)
	GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client v2action.NOAAClient, config v2action.Config) (<-chan *v2action.LogMessage, <-chan error, v2action.Warnings, error)
}

type LogsCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	Recent          bool         `long:"recent" description:"Dump recent logs instead of tailing"`
	usage           interface{}  `usage:"CF_NAME logs APP_NAME"`
	relatedCommands interface{}  `related_commands:"app, apps, ssh"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       LogsActor
	NOAAClient  *consumer.Consumer
}

func (cmd *LogsCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	cmd.NOAAClient = shared.NewNOAAClient(ccClient.DopplerEndpoint(), config, uaaClient, ui)

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

	return cmd.streamLogs()
}

func (cmd LogsCommand) displayRecentLogs() error {
	messages, warnings, err := cmd.Actor.GetRecentLogsForApplicationByNameAndSpace(
		cmd.RequiredArgs.AppName,
		cmd.Config.TargetedSpace().GUID,
		cmd.NOAAClient,
		cmd.Config,
	)

	for _, message := range messages {
		cmd.UI.DisplayLogMessage(message, true)
	}

	cmd.UI.DisplayWarnings(warnings)
	return err
}

func (cmd LogsCommand) streamLogs() error {
	messages, logErrs, warnings, err := cmd.Actor.GetStreamingLogsForApplicationByNameAndSpace(
		cmd.RequiredArgs.AppName,
		cmd.Config.TargetedSpace().GUID,
		cmd.NOAAClient,
		cmd.Config,
	)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

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

			cmd.NOAAClient.Close()
			return logErr
		}

		if messagesClosed && errLogsClosed {
			break
		}
	}

	return nil
}
