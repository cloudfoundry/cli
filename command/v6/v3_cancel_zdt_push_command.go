package v6

import (
	"errors"

	"code.cloudfoundry.org/cli/v7/actor/sharedaction"
	"code.cloudfoundry.org/cli/v7/actor/v3action"
	"code.cloudfoundry.org/cli/v7/command"
	"code.cloudfoundry.org/cli/v7/command/flag"
	"code.cloudfoundry.org/cli/v7/command/v6/shared"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . V3CancelZdtPushActor

type V3CancelZdtPushActor interface {
	CancelDeploymentByAppNameAndSpace(appName string, spaceGUID string) (v3action.Warnings, error)
}

type V3CancelZdtPushCommand struct {
	RequiredArgs       flag.AppName `positional-args:"yes"`
	UI                 command.UI
	Config             command.Config
	CancelZdtPushActor V3CancelZdtPushActor
	SharedActor        command.SharedActor
	usage              interface{} `usage:"CF_NAME v3-cancel-zdt-push APP_NAME"`
}

func (cmd *V3CancelZdtPushCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewV3BasedClients(config, ui, true)
	if err != nil {
		return err
	}

	cmd.CancelZdtPushActor = v3action.NewActor(ccClient, config, sharedActor, uaaClient)
	cmd.SharedActor = sharedActor

	return nil
}

func (cmd V3CancelZdtPushCommand) Execute(args []string) error {
	cmd.UI.DisplayWarning(command.ExperimentalWarning)

	err := cmd.validateArgs()
	if err != nil {
		return err
	}

	err = cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	_, err = cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	warnings, err := cmd.CancelZdtPushActor.CancelDeploymentByAppNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayText("Deployment cancelled, rolling back")
	return nil
}

func (cmd V3CancelZdtPushCommand) validateArgs() error {
	if cmd.RequiredArgs.AppName == "" {
		return errors.New("No app name given")
	}
	return nil
}
