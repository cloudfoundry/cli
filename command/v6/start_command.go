package v6

import (
	"github.com/cloudfoundry/noaa/consumer"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2v3action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	sharedV3 "code.cloudfoundry.org/cli/command/v3/shared"
	"code.cloudfoundry.org/cli/command/v6/shared"
	log "github.com/sirupsen/logrus"
)

//go:generate counterfeiter . StartActor

type StartActor interface {
	AppActor
	StartApplication(app v2action.Application, client v2action.NOAAClient) (<-chan *v2action.LogMessage, <-chan error, <-chan v2action.ApplicationStateChange, <-chan string, <-chan error)
}

type StartCommand struct {
	RequiredArgs        flag.AppName `positional-args:"yes"`
	usage               interface{}  `usage:"CF_NAME start APP_NAME"`
	envCFStagingTimeout interface{}  `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for buildpack staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout interface{}  `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`
	relatedCommands     interface{}  `related_commands:"apps, logs, scale, ssh, stop, restart, run-task"`

	UI                      command.UI
	Config                  command.Config
	SharedActor             command.SharedActor
	Actor                   StartActor // todo rename key to StartActor to avoid confusion
	ApplicationSummaryActor shared.ApplicationSummaryActor
	NOAAClient              *consumer.Consumer
}

func (cmd *StartCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}

	ccClientV3, _, err := sharedV3.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	v2Actor := v2action.NewActor(ccClient, uaaClient, config)
	v3Actor := v3action.NewActor(ccClientV3, config, sharedActor, nil)

	cmd.Actor = v2Actor

	cmd.ApplicationSummaryActor = v2v3action.NewActor(v2Actor, v3Actor)

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

	messages, logErrs, appState, apiWarnings, errs := cmd.Actor.StartApplication(app, cmd.NOAAClient)
	err = shared.PollStart(cmd.UI, cmd.Config, messages, logErrs, appState, apiWarnings, errs)
	if err != nil {
		return err
	}

	cmd.UI.DisplayNewline()

	if err := command.MinimumCCAPIVersionCheck(cmd.ApplicationSummaryActor.CloudControllerV3APIVersion(), ccversion.MinVersionApplicationFlowV3); err != nil {
		log.WithField("v3_api_version", cmd.ApplicationSummaryActor.CloudControllerV3APIVersion()).Debug("using v2 for app display")
		appSummary, warnings, err := cmd.Actor.GetApplicationSummaryByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
		shared.DisplayAppSummary(cmd.UI, appSummary, true)
	} else {
		log.WithField("v3_api_version", cmd.ApplicationSummaryActor.CloudControllerV3APIVersion()).Debug("using v3 for app display")
		appSummary, warnings, err := cmd.ApplicationSummaryActor.GetApplicationSummaryByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, true)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
		sharedV3.NewAppSummaryDisplayer2(cmd.UI).AppDisplay(appSummary, true)
	}
	return nil
}
