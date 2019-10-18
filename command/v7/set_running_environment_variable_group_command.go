package v7

import (
	"encoding/json"
	"fmt"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . SetRunningEnvironmentVariableGroupActor

type SetRunningEnvironmentVariableGroupActor interface {
	SetEnvironmentVariableGroup(group constant.EnvironmentVariableGroupName, envVars ccv3.EnvironmentVariables) (v7action.Warnings, error)
}

type SetRunningEnvironmentVariableGroupCommand struct {
	RequiredArgs    flag.SetEnvVarGroup `positional-args:"yes"`
	usage           interface{}         `usage:"CF_NAME set-running-environment-variable-group '{\"name\":\"value\",\"name\":\"value\"}'"`
	relatedCommands interface{}         `related_commands:"set-env, staging-environment-variable-group"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SetRunningEnvironmentVariableGroupActor
}

func (cmd *SetRunningEnvironmentVariableGroupCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())

	return nil
}

func (cmd SetRunningEnvironmentVariableGroupCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Setting the contents of the running environment variable group as {{.Username}}...", map[string]interface{}{
		"Username": user.Name,
	})

	var envVars ccv3.EnvironmentVariables
	err = json.Unmarshal([]byte(fmt.Sprintf(`{"var":%s}`, cmd.RequiredArgs.EnvVarGroupJson)), &envVars)
	if err != nil {
		return errors.New("Invalid environment variable group provided. Please provide a valid JSON object.")
	}

	warnings, err := cmd.Actor.SetEnvironmentVariableGroup(
		constant.RunningEnvironmentVariableGroup,
		envVars,
	)
	cmd.UI.DisplayWarningsV7(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
