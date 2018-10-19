package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

//go:generate counterfeiter . SetEnvActor

type SetEnvActor interface {
	CloudControllerAPIVersion() string
	SetEnvironmentVariableByApplicationNameAndSpace(appName string, spaceGUID string, envPair v7action.EnvironmentVariablePair) (v7action.Warnings, error)
}

type SetEnvCommand struct {
	RequiredArgs    flag.SetEnvironmentArgs `positional-args:"yes"`
	usage           interface{}             `usage:"CF_NAME set-env APP_NAME ENV_VAR_NAME ENV_VAR_VALUE"`
	relatedCommands interface{}             `related_commands:"v3-apps, env, v3-restart, set-running-environment-variable-group, set-staging-environment-variable-group, v3-stage, unset-env"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SetEnvActor
}

func (cmd *SetEnvCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, _, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil)

	return nil
}

func (cmd SetEnvCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	appName := cmd.RequiredArgs.AppName
	cmd.UI.DisplayTextWithFlavor("Setting env variable {{.EnvVarName}} for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":    appName,
		"EnvVarName": cmd.RequiredArgs.EnvironmentVariableName,
		"OrgName":    cmd.Config.TargetedOrganization().Name,
		"SpaceName":  cmd.Config.TargetedSpace().Name,
		"Username":   user.Name,
	})

	warnings, err := cmd.Actor.SetEnvironmentVariableByApplicationNameAndSpace(
		appName,
		cmd.Config.TargetedSpace().GUID,
		v7action.EnvironmentVariablePair{
			Key:   cmd.RequiredArgs.EnvironmentVariableName,
			Value: string(cmd.RequiredArgs.EnvironmentVariableValue),
		})
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayText("TIP: Use 'cf v3-stage {{.AppName}}' to ensure your env variable changes take effect.", map[string]interface{}{
		"AppName": appName,
	})

	return nil
}
