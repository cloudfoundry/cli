package v6

import (
	"code.cloudfoundry.org/cli/actor/loggingaction"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2v3action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v6/shared"
	sharedV3 "code.cloudfoundry.org/cli/command/v6/shared"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"
	log "github.com/sirupsen/logrus"
)

//go:generate counterfeiter . RestageActor

type RestageActor interface {
	GetApplicationByNameAndSpace(name string, spaceGUID string) (v2action.Application, v2action.Warnings, error)
	GetApplicationSummaryByNameAndSpace(name string, spaceGUID string) (v2action.ApplicationSummary, v2action.Warnings, error)
	RestageApplication(app v2action.Application, client loggingaction.LogCacheClient) (<-chan loggingaction.LogMessage, <-chan error, <-chan v2action.ApplicationStateChange, <-chan string, <-chan error)
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
	LogCacheClient          *logcache.Client
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

	ccClientV3, _, err := sharedV3.NewV3BasedClients(config, ui, true, "")
	if err != nil {
		return err
	}

	v2Actor := v2action.NewActor(ccClient, uaaClient, config)
	v3Actor := v3action.NewActor(ccClientV3, config, sharedActor, nil)

	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)
	cmd.ApplicationSummaryActor = v2v3action.NewActor(v2Actor, v3Actor)

	cmd.LogCacheClient = shared.NewLogCacheClient(ccClient.LogCacheEndpoint(), config, ui)

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

	cmd.UI.DisplayWarning("This action will cause app downtime.")
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

	messages, logErrs, appState, apiWarnings, errs := cmd.Actor.RestageApplication(app, cmd.LogCacheClient)
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
