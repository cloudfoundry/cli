package v3

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . V3StageActor

type V3StageActor interface {
	GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client v3action.NOAAClient) (<-chan *v3action.LogMessage, <-chan error, v3action.Warnings, error)
	StagePackage(packageGUID string, appName string) (<-chan v3action.Build, <-chan v3action.Warnings, <-chan error)
}

type V3StageCommand struct {
	RequiredArgs        flag.AppName `positional-args:"yes"`
	PackageGUID         string       `long:"package-guid" description:"The guid of the package to stage" required:"true"`
	usage               interface{}  `usage:"CF_NAME v3-stage APP_NAME --package-guid PACKAGE_GUID"`
	envCFStagingTimeout interface{}  `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for buildpack staging, in minutes" environmentDefault:"15"`

	UI          command.UI
	Config      command.Config
	NOAAClient  v3action.NOAAClient
	SharedActor command.SharedActor
	Actor       V3StageActor
}

func (cmd *V3StageCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor()

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}

	cmd.Actor = v3action.NewActor(ccClient, config)
	cmd.NOAAClient = shared.NewNOAAClient(ccClient.APIInfo.Logging(), config, uaaClient, ui)

	return nil
}

func (cmd V3StageCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(cmd.Config, true, true)
	if err != nil {
		return shared.HandleError(err)
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

	logStream, logErrStream, logWarnings, logErr := cmd.Actor.GetStreamingLogsForApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, cmd.NOAAClient)
	cmd.UI.DisplayWarnings(logWarnings)
	if logErr != nil {
		return shared.HandleError(logErr)
	}

	buildStream, warningsStream, errStream := cmd.Actor.StagePackage(cmd.PackageGUID, cmd.RequiredArgs.AppName)
	_, err = shared.PollStage(buildStream, warningsStream, errStream, logStream, logErrStream, cmd.UI)
	if err == nil {
		cmd.UI.DisplayOK()
	}

	return err
}
