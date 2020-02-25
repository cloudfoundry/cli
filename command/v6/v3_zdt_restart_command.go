package v6

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . V3ZeroDowntimeRestartActor

type V3ZeroDowntimeRestartActor interface {
	ZeroDowntimePollStart(appGUID string, warningsChannel chan<- v3action.Warnings) error
	CreateDeployment(appGUID, dropletGUID string) (string, v3action.Warnings, error)

	CloudControllerAPIVersion() string
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v3action.Application, v3action.Warnings, error)
	StartApplication(appGUID string) (v3action.Warnings, error)
}

type V3ZeroDowntimeRestartCommand struct {
	RequiredArgs flag.AppName `positional-args:"yes"`
	usage        interface{}  `usage:"CF_NAME v3-zdt-restart APP_NAME"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       V3ZeroDowntimeRestartActor
}

func (cmd *V3ZeroDowntimeRestartCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, _, err := shared.NewV3BasedClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v3action.NewActor(ccClient, config, nil, nil)

	return nil
}

func (cmd V3ZeroDowntimeRestartCommand) Execute(args []string) error {
	cmd.UI.DisplayWarning(command.ExperimentalWarning)

	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	app, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if app.Stopped() {
		cmd.UI.DisplayTextWithFlavor("Starting app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
			"AppName":   cmd.RequiredArgs.AppName,
			"OrgName":   cmd.Config.TargetedOrganization().Name,
			"SpaceName": cmd.Config.TargetedSpace().Name,
			"Username":  user.Name,
		})

		warnings, err = cmd.Actor.StartApplication(app.GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
	} else {
		cmd.UI.DisplayTextWithFlavor("Starting deployment for app {{.AppName}} in org {{.CurrentOrg}} / space {{.CurrentSpace}} as {{.CurrentUser}}...", map[string]interface{}{
			"AppName":      cmd.RequiredArgs.AppName,
			"CurrentSpace": cmd.Config.TargetedSpace().Name,
			"CurrentOrg":   cmd.Config.TargetedOrganization().Name,
			"CurrentUser":  user.Name,
		})

		_, warnings, err = cmd.Actor.CreateDeployment(app.GUID, "")
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		cmd.UI.DisplayText("Waiting for app to start...")

		warnings := make(chan v3action.Warnings)
		done := make(chan bool)
		go func() {
			for {
				select {
				case message := <-warnings:
					cmd.UI.DisplayWarnings(message) // untested
				case <-done:
					return
				}
			}
		}()

		err = cmd.Actor.ZeroDowntimePollStart(app.GUID, warnings)
		done <- true
		if err != nil {
			return err
		}
	}

	cmd.UI.DisplayOK()

	return nil
}
