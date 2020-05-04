package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/flag"
)

type UnsetEnvCommand struct {
	BaseCommand

	RequiredArgs    flag.UnsetEnvironmentArgs `positional-args:"yes"`
	usage           interface{}               `usage:"CF_NAME unset-env APP_NAME ENV_VAR_NAME"`
	relatedCommands interface{}               `related_commands:"apps, env, restart, set-env, stage"`
}

func (cmd UnsetEnvCommand) Execute(args []string) error {

	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	appName := cmd.RequiredArgs.AppName
	cmd.UI.DisplayTextWithFlavor("Removing env variable {{.EnvVarName}} from app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":    appName,
		"EnvVarName": cmd.RequiredArgs.EnvironmentVariableName,
		"OrgName":    cmd.Config.TargetedOrganization().Name,
		"SpaceName":  cmd.Config.TargetedSpace().Name,
		"Username":   user.Name,
	})

	warnings, err := cmd.Actor.UnsetEnvironmentVariableByApplicationNameAndSpace(
		appName,
		cmd.Config.TargetedSpace().GUID,
		cmd.RequiredArgs.EnvironmentVariableName,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		switch errVal := err.(type) {
		case actionerror.EnvironmentVariableNotSetError:
			cmd.UI.DisplayText(errVal.Error())
		default:
			return err
		}
	}

	cmd.UI.DisplayOK()
	if err == nil {
		cmd.UI.DisplayText("TIP: Use 'cf restage {{.AppName}}' to ensure your env variable changes take effect.", map[string]interface{}{
			"AppName": appName,
		})
	}

	return nil
}
