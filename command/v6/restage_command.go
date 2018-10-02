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

//go:generate counterfeiter . RestageActor

type RestageActor interface {
	AppActor
	RestageApplication(app v2action.Application, client v2action.NOAAClient) (<-chan *v2action.LogMessage, <-chan error, <-chan v2action.ApplicationStateChange, <-chan string, <-chan error)
}

type RestageCommand struct {
	RequiredArgs        flag.AppName `positional-args:"yes"`
	usage               interface{}  `usage:"CF_NAME restage APP_NAME"`
	relatedCommands     interface{}  `related_commands:"restart"`
	envCFStagingTimeout interface{}  `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for buildpack staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout interface{}  `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`

	UI                      command.UI
	Config                  command.Config
	SharedActor             command.SharedActor
	Actor                   RestageActor
	ApplicationSummaryActor shared.ApplicationSummaryActor
	NOAAClient              *consumer.Consumer
}

func (cmd *RestageCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)
	sharedActor := sharedaction.NewActor(config)

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

	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)
	cmd.ApplicationSummaryActor = v2v3action.NewActor(v2Actor, v3Actor)

	cmd.NOAAClient = shared.NewNOAAClient(ccClient.DopplerEndpoint(), config, uaaClient, ui)

	return nil
}

func (cmd RestageCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Restaging app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
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

	messages, logErrs, appState, apiWarnings, errs := cmd.Actor.RestageApplication(app, cmd.NOAAClient)
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
		shared.NewAppSummaryDisplayer2(cmd.UI).AppDisplay(appSummary, true)
	}

	return nil
}
