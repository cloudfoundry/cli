package v2

import (
	"github.com/cloudfoundry/noaa/consumer"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . StartActor

type StartActor interface {
	AppActor
	StartApplication(app v2action.Application, client v2action.NOAAClient, config v2action.Config) (<-chan *v2action.LogMessage, <-chan error, <-chan v2action.ApplicationStateChange, <-chan string, <-chan error)
}

type StartCommand struct {
	RequiredArgs        flag.AppName `positional-args:"yes"`
	usage               interface{}  `usage:"CF_NAME start APP_NAME"`
	envCFStagingTimeout interface{}  `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for buildpack staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout interface{}  `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`
	relatedCommands     interface{}  `related_commands:"apps, logs, scale, ssh, stop, restart, run-task"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       StartActor
	NOAAClient  *consumer.Consumer
}

func (cmd *StartCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd StartCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Starting app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"AppName":     cmd.RequiredArgs.AppName,
			"OrgName":     cmd.Config.TargetedOrganization().Name,
			"SpaceName":   cmd.Config.TargetedSpace().Name,
			"CurrentUser": user.Name,
		})

	app, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if app.Started() {
		cmd.UI.DisplayText("App {{.AppName}} is already started",
			map[string]interface{}{
				"AppName": cmd.RequiredArgs.AppName,
			})
		return nil
	}

	messages, logErrs, appState, apiWarnings, errs := cmd.Actor.StartApplication(app, cmd.NOAAClient, cmd.Config)
	err = shared.PollStart(cmd.UI, cmd.Config, messages, logErrs, appState, apiWarnings, errs)
	if err != nil {
		return err
	}

	cmd.UI.DisplayNewline()

	appSummary, warnings, err := cmd.Actor.GetApplicationSummaryByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	shared.DisplayAppSummary(cmd.UI, appSummary, true)

	return nil
}
