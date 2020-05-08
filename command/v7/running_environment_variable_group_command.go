package v7

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/util/ui"
)

type RunningEnvironmentVariableGroupCommand struct {
	command.BaseCommand

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
		table, err := buildEnvVarsTable(envVars)
		if err != nil {
			return err
		}

		cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
	}

	return nil
}
