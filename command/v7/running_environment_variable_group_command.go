package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/util/ui"
)

type RunningEnvironmentVariableGroupCommand struct {
	BaseCommand

	usage           interface{} `usage:"CF_NAME running-environment-variable-group"`
	relatedCommands interface{} `related_commands:"env, staging-environment-variable-group"`
}

func (cmd RunningEnvironmentVariableGroupCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting the running environment variable group as {{.Username}}...", map[string]interface{}{
		"Username": user.Name,
	})
	cmd.UI.DisplayNewline()

	envVars, warnings, err := cmd.Actor.GetEnvironmentVariableGroup(constant.RunningEnvironmentVariableGroup)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(envVars) == 0 {
		cmd.UI.DisplayTextWithFlavor("No running environment variable group has been set.")
	} else {
		cmd.displayTable(envVars)
	}

	return nil
}

func (cmd RunningEnvironmentVariableGroupCommand) displayTable(envVars v7action.EnvironmentVariableGroup) {
	var keyValueTable = [][]string{
		{"variable name", "assigned value"},
	}

	for envVarName, envVarValue := range envVars {
		keyValueTable = append(keyValueTable, []string{
			envVarName,
			envVarValue.Value,
		})
	}

	cmd.UI.DisplayTableWithHeader("", keyValueTable, ui.DefaultTableSpacePadding)
}
