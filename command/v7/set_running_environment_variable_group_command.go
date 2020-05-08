package v7

import (
	"encoding/json"
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/command/flag"
)

type SetRunningEnvironmentVariableGroupCommand struct {
	command.BaseCommand

	RequiredArgs    flag.SetEnvVarGroup `positional-args:"yes"`
	usage           interface{}         `usage:"CF_NAME set-running-environment-variable-group '{\"name\":\"value\",\"name\":\"value\"}'"`
	relatedCommands interface{}         `related_commands:"set-env, running-environment-variable-group"`
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
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
