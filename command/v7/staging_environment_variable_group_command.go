package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/util/ui"
)

type StagingEnvironmentVariableGroupCommand struct {
	BaseCommand

	usage           interface{} `usage:"CF_NAME staging-environment-variable-group"`
	relatedCommands interface{} `related_commands:"env, running-environment-variable-group"`
}

func (cmd StagingEnvironmentVariableGroupCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting the staging environment variable group as {{.Username}}...", map[string]interface{}{
		"Username": user.Name,
	})
	cmd.UI.DisplayNewline()

	envVars, warnings, err := cmd.Actor.GetEnvironmentVariableGroup(constant.StagingEnvironmentVariableGroup)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(envVars) == 0 {
		cmd.UI.DisplayTextWithFlavor("No staging environment variable group has been set.")
	} else {
		table, err := buildEnvVarsTable(envVars)
		if err != nil {
			return err
		}

		cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
	}

	return nil
}

func buildEnvVarsTable(envVars v7action.EnvironmentVariableGroup) ([][]string, error) {
	var keyValueTable = [][]string{
		{"variable name", "assigned value"},
	}

	for envVarName, envVarValue := range envVars {
		keyValueTable = append(keyValueTable, []string{
			envVarName,
			envVarValue.Value,
		})
	}

	return keyValueTable, nil
}
