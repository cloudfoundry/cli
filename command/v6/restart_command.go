package v6

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2v3action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v6/shared"
	sharedV3 "code.cloudfoundry.org/cli/command/v6/shared"
	"github.com/cloudfoundry/noaa/consumer"
	log "github.com/sirupsen/logrus"
)

//go:generate counterfeiter . RestartActor

type RestartActor interface {
	GetApplicationByNameAndSpace(name string, spaceGUID string) (v2action.Application, v2action.Warnings, error)
	GetApplicationSummaryByNameAndSpace(name string, spaceGUID string) (v2action.ApplicationSummary, v2action.Warnings, error)
	RestartApplication(app v2action.Application, client v2action.NOAAClient) (<-chan *v2action.LogMessage, <-chan error, <-chan v2action.ApplicationStateChange, <-chan string, <-chan error)
}

type RestartCommand struct {
	RequiredArgs        flag.AppName `positional-args:"yes"`
	usage               interface{}  `usage:"CF_NAME restart APP_NAME"`
	relatedCommands     interface{}  `related_commands:"restage, restart-app-instance"`
	envCFStagingTimeout interface{}  `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for buildpack staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout interface{}  `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`

	UI                      command.UI
	Config                  command.Config
	SharedActor             command.SharedActor
	Actor                   RestartActor
	ApplicationSummaryActor shared.ApplicationSummaryActor
	NOAAClient              *consumer.Consumer
}

func (cmd *RestartCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui)
	if err != nil {
		return err

	}
	ccClientV3, _, err := sharedV3.NewV3BasedClients(config, ui, true)
	if err != nil {
		return err
	}

	v2Actor := v2action.NewActor(ccClient, uaaClient, config)
	v3Actor := v3action.NewActor(ccClientV3, config, sharedActor, nil)

	cmd.Actor = v2Actor
	cmd.ApplicationSummaryActor = v2v3action.NewActor(v2Actor, v3Actor)

	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)
	cmd.ApplicationSummaryActor = v2v3action.NewActor(v2Actor, v3Actor)
	cmd.NOAAClient = shared.NewNOAAClient(ccClient.DopplerEndpoint(), config, uaaClient, ui)

	return nil
}

func (cmd RestartCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Restarting app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
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

	messages, logErrs, appState, apiWarnings, errs := cmd.Actor.RestartApplication(app, cmd.NOAAClient)
	err = shared.PollStart(cmd.UI, cmd.Config, messages, logErrs, appState, apiWarnings, errs)
	if err != nil {
		return err
	}

	cmd.UI.DisplayNewline()
	log.WithField("v3_api_version", cmd.ApplicationSummaryActor.CloudControllerV3APIVersion()).Debug("using v3 for app display")
	appSummary, v3Warnings, err := cmd.ApplicationSummaryActor.GetApplicationSummaryByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, true)
	cmd.UI.DisplayWarnings(v3Warnings)
	if err != nil {
		return err
	}
	shared.NewAppSummaryDisplayer2(cmd.UI).AppDisplay(appSummary, true)
	return nil
}
