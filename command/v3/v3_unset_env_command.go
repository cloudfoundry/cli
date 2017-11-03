package v3

import (
	"net/http"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . V3UnsetEnvActor

type V3UnsetEnvActor interface {
	CloudControllerAPIVersion() string
	UnsetEnvironmentVariableByApplicationNameAndSpace(appName string, spaceGUID string, EnvironmentVariableName string) (v3action.Warnings, error)
}

type V3UnsetEnvCommand struct {
	RequiredArgs    flag.UnsetEnvironmentArgs `positional-args:"yes"`
	usage           interface{}               `usage:"CF_NAME v3-unset-env APP_NAME ENV_VAR_NAME"`
	relatedCommands interface{}               `related_commands:"v3-apps, v3-env, v3-restart, v3-set-env, v3-stage"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       V3UnsetEnvActor
}

func (cmd *V3UnsetEnvCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, _, err := shared.NewClients(config, ui, true)
	if err != nil {
		if v3Err, ok := err.(ccerror.V3UnexpectedResponseError); ok && v3Err.ResponseCode == http.StatusNotFound {
			return translatableerror.MinimumAPIVersionNotMetError{MinimumVersion: ccversion.MinVersionV3}
		}

		return err
	}
	cmd.Actor = v3action.NewActor(ccClient, config, nil, nil)

	return nil
}

func (cmd V3UnsetEnvCommand) Execute(args []string) error {
	cmd.UI.DisplayText(command.ExperimentalWarning)
	cmd.UI.DisplayNewline()

	err := command.MinimumAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), ccversion.MinVersionV3)
	if err != nil {
		return err
	}

	err = cmd.SharedActor.CheckTarget(true, true)
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
		cmd.UI.DisplayText("TIP: Use 'cf v3-stage {{.AppName}}' to ensure your env variable changes take effect.", map[string]interface{}{
			"AppName": appName,
		})
	}

	return nil
}
